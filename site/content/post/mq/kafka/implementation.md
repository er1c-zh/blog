---
title: "kafka的实现"
date: 2022-10-26T07:22:26+08:00
draft: false
tags:
    - mq
    - kafka
    - memo
    - what
    - way-to-kafka
order: 4
---

kafka的实现。

<!--more-->

{{% serial_index way-to-kafka %}}

# 网络层

网络层实现了简单的转发。

`sendfile`通过给`Message`增加`writeTo`接口来实现，
这使得存储在文件中的`Message`可以通过`transferTo`来实现发送而不再需要使用额外的buffer和复制。

线程模型是一个acceptor线程和N个处理固定链接数的工作线程。

# 消息

`Message`由：

1. 变长header
1. 变长opaque key byte array
1. 变长opaque value byte array

将数据设置为opaque可以支持更多的用途和使用方式。

# 消息格式

消息都被存储在**批量消息**(batch)中。
一个批量消息包含一个或多个消息。
批量消息和消息有各自的header。

## 批量消息 RecordBatch

```
baseOffset: int64
batchLength: int32
partitionLeaderEpoch: int32
magic: int8 (current magic value is 2)
crc: int32
attributes: int16
    bit 0~2:
        0: no compression
        1: gzip
        2: snappy
        3: lz4
        4: zstd
    bit 3: timestampType
    bit 4: isTransactional (0 means not transactional)
    bit 5: isControlBatch (0 means not a control batch)
    bit 6: hasDeleteHorizonMs (0 means baseTimestamp is not set as the delete horizon for compaction)
    bit 7~15: unused
lastOffsetDelta: int32
baseTimestamp: int64
maxTimestamp: int64
producerId: int64
producerEpoch: int16
baseSequence: int32
records: [Record]
```

## 消息 Record

```
length: varint
attributes: int8
    bit 0~7: unused
timestampDelta: varlong
offsetDelta: varint
keyLength: varint
key: byte[]
valueLen: varint
value: byte[]
Headers => [Header]
```

```
// RecordHeader
headerKeyLength: varint
headerKey: String
headerValueLength: varint
Value: byte[]
```

# Log

对于有两个partition的topic`mt`，
有两个目录存储数据。
(`mt_0`和`mt_1`)

log_file由一系列log_entry构成。

每个log文件名是文件中的第一个消息的offset。

log_entry开头是4byte的长度信息。

每个message由一个64位的offset来唯一标识。
这个标识符是这个partition从开始以来发送的所有数据的byte为单位的offset。

使用offset来唯一标记message是不同寻常的方案。
考虑到使用全局唯一的id以及和offset的映射表来做，需要巨大的存储和性能成本。

为了解决这个问题，一个自然的方案是，partition_id + node_id + per-partition atomic counter来作为id。
那么，使用offset也就是一个自然的选择了。（同样是跟着消息的写入自增、且唯一。）

## 写入

log文件支持从最后一个文件追加写入。
根据配置，如果文件大小超过了，就会产生一个新的log文件。

有两项配置用来控制可靠性：

- 最多多少条message触发刷盘
- 最多多长时间触发刷盘

## 读取

读取操作的入参是offset和需要读取的数据size。

如果超过了size，会不断的尝试翻倍size来尝试读取。
可以配置上限来丢弃过大的消息。

### 具体实现

定位到存储了这个offset的log segment file。
根据offset来计算segment文件中的offset。
搜索是在每个文件在内存中的范围上的简单的二分查找。

## 删除

数据的删除按照segment为单位。
log manager根据time和size来判断一个segment是否是可以删除的。
时间策略会看一个segment中最新的消息作为；
size策略默认关闭，如果开启，manager会持续的删除最旧的segment来保持所有的size不超过配置。

同时开启时，任意条件满足都会删除。

利用COW机制来实现读取和删除不冲突。

## 保证

log提供一个配置项用来控制是否没有刷盘的最大message量。

开始一个恢复过程中，遍历最新的segment，检查每个message entry

# 分发 Distribution

消费者跟踪记录每个partition消费到的最大的offset，
在重启的时候会继续消费。

kafka提供一个可选的方案：
将给定的consumer group的offset记录到一个特定的broker中，这个broker称作group coordinator。

consumer可以向任意broker来查询自己group的coordinator是哪个broker。

当coordinator变动时，consumer需要重新进行这个流程，来发现新的coordinator。

offset可以被自动或者手动的提交。

当coordinator收到了提交offset的请求时，
会将这个请求append到一个特别的、压缩过的kafka topic(*__consumer_offsets*)。
coordinator会在所有的分片返回写入成功之后，
返回给consumer一个写入成功的响应。
如果有任意分片写入失败，
coordinator会返回失败，
consumer可以重试提交。
考虑到只需要维护每个partition最新的offset，broker会周期性的压缩offset topic。
coordinator会将offset缓存下来，为了更快的查询。

coordinator被查询offset时，会直接从cache中返回offset的vector。
如果coordinator刚刚启动或者其他原因构建好cache，
会返回失败，直到将offset topic的数据加载到cache；consumer可以重试查询。

## Zookeeper如何协调consumer和broker

### borcker节点注册

```
/brokers/ids/[0...N] -> {"jmx_port":...,"timestamp":...,"endpoints":[...],"host":...,"version":...,"port":...} (ephemeral node)
```
每一项表示一个broker。
broker的配置中会给出一个逻辑id来唯一标识这个broker。
启动时，broker通过创建一个自己的逻辑id的znode来注册。
逻辑id是为了在broker从不同的物理集群迁移时，
不影响消费者。

因为是ephemeral节点，所以当broker失联的时候，
节点会消失。

### broker topic 注册

```
/brokers/topics/[topic]/partitions/[0...N]/state --> {"controller_epoch":...,"leader":...,"version":...,"leader_epoch":...,"isr":[...]} (ephemeral node)
```

### cluster id

集群id是一个唯一、不可变更的字符串。

# Reference

- [kafka IMPLEMENTATION](https://kafka.apache.org/documentation/#implementation)
- [Introducing the NIO SocketServer Implementation](https://web.archive.org/web/20120619234320/http://sna-projects.com/blog/2009/08/introducing-the-nio-socketserver-implementation/)

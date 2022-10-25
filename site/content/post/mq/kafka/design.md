---
title: "kafka设计"
date: 2022-10-24T20:08:12+08:00
draft: false
tags:
    - mq
    - kafka
    - memo
    - what
    - way-to-kafka
order: 3
---

kafka的设计思路。

<!--more-->

{{% serial_index way-to-kafka %}}

# kafka的设计目标与动机 *(Motivation)*

> We designed Kafka to be able to 
> act as a unified platform for 
> handling all the real-time data feeds a large company might have. 

一句话概括：kafka被设计为一个能够处理大公司产出的实时数据流的统一平台。
这意味着：

- 足够大的吞吐量，用来支持大流量的数据流，比如实时日志聚合。
- 优雅地支持较大数据量的积压 *(backlog)* ，比如周期性的离线数据的导入。
- 低延迟，用来满足传统的消息组件的功能。

除此之外，kafka还被希望能够：

- 支持实时的从分片、分布式的流中创建新的下游流。
- 良好的容灾机制，处理好机器失效等问题。

**这些目标使得kafka更像一个数据库日志系统而不是一个传统的消息系统。**

# 持久化 Persistence

## 使用磁盘与操作系统的文件系统

> Don't fear the filesystem!

考虑几个因素：

- 硬盘顺序写的带宽足够大。

    > As a result the performance of linear writes on a JBOD configuration with six 7200rpm SATA 
    > RAID-5 array is about 600MB/sec but the performance of random writes is only about 
    > 100k/sec—a difference of over 6000X. 

- 操作系统面向块存储设备提供的预读 *(read ahead)* 和批量写 *(write behind)* 机制、页缓存机制以及页缓存与磁盘中的一致性保证。
- （基于JVM）在内存中维护数据结构会增加内存占用（甚至翻倍），以及内存占用增大导致垃圾回收导致的性能问题。

又考虑，可以将缓存的一致性以及维护的问题交给操作系统来实现，
来降低kafka本身的复杂度。

所以kafka被设计为以页缓存 *(pagecache)* 为存储消息的核心的形式。
参考了[Varnish](http://varnish-cache.org/wiki/ArchitectNotes)的设计。

## 常数时间复杂度

考虑上述使用磁盘的情况，
O(logN)的存储数据结构（如BTree/B+Tree）不能像在内存中一样当作常数时间复杂度来处理，
另外考虑读写的并发问题。
所以选择将消息当作一个日志文件来存储，
消费消息是从日志文件头读取，
写入消息是追加数据。
同时实现了读写操作在常数时间复杂度内完成、
读写操作解耦无并发问题、
以及利用磁盘的顺序读写能力，
更低的存储成本与更高的读写性能（消息可以更久的被存储）。

这种实现相比BTree等的实现不足在于实现搜索等功能较复杂。

# 更好的效率

考虑到通常有至少一个消费者会消费到消息，
设计上将“消费”的消耗尽量的降低。

在使用机械硬盘的情况下，
有两种产生性能问题的情况：

1. 很多次小的IO操作。

    这个问题会在消费者、生产者以及服务端持久化消息时发生。
    为了解决这个问题，
    构造了 **消息集合** *(message set)* 的概念。
    这可以增加网络报文的大小，
    持久化时可以减少IO的次数，
    消费者也可以减少请求的次数。
    
1. 过多的数据拷贝。

    kafka定义了一个生产者、服务器、消费者三方共用的二进制格式来存储、传输消息。
    
    首先是能够跨越消息来进行压缩，
    这种机制被称作 **端到端的批量压缩** *(end-to-end batch compression)* 。
    举个例子，批量压缩可以将多个消息中相同的json的字段名称压缩，
    而在单个消息中进行压缩时无法进行这种压缩。

    另一项收益是，
    服务器操作消息时不需要解析消息的格式。
    具体的，
    服务器可以将一个个消息集合当作文件来处理，
    利用Linux操作系统提供的`sendfile`调用，
    将文件的内容直接放到socket的发送缓冲区中，
    减少了数据的拷贝。

    > 现代unix操作系统提供高效的方法用来将pagecache的数据传递给一个socket。

    常见的操作路径是：

    1. 操作系统从Disk读取数据到内核空间的pagecache。
    1. 应用从内核空间将数据读取（产生复制）到用户空间的buffer。
    1. 应用将用户空间的buffer的数据写入到（产生复制）在内核空间中的socket缓冲区。
    1. 操作系统从socket的缓冲区写入（产生复制）到NIC的缓冲区来完成发送。

    包含了四次复制和两次系统调用。
    改用`sendfile`之后，只需要一次系统调用和一次复制：从pagecache复制到NIC缓冲区。

# 生产者

## 负载均衡

生产者总是将数据发送给该分片的的leader broker。
为了帮助生产者完成这个设计，
每个kafka的节点都能回答所有节点的活性以及partition的leader。

生产者client决定将消息发送到那个partition。
负载均衡工作可以在这里完成，
支持随机或者按照语意进行划分。

一个实用的case：

生产者可以根据比如uid来选择分片发送，
这可以使得一个uid的消息都发送到一个分片中。
这使得consumer可以进行本地敏感*(locality-sensitive)*的处理。

## 异步发送

开启批量发送之后，生产者会尝试在内存中聚合数据并通过一个批量请求来提交。
可以设置不超过某个大小或者不等待超过某个时间。

这项设置可以实现通过牺牲一点延迟来换取更好的吞吐。

# 消费者

消费者通过发送 **fetch请求** 给想要消费的partition的leading broker来完成消费。
每个请求中，消费者设定想要开始消费的offset，接收到从该位置开始的一组数据。

显然易见的，消费者可以控制消费的位置以及如果需要的话，可以重新消费。

## 推或者拉？

kafka选择了传统的消息系统的设计方式：生产者推，消费者拉。

### push-based system

缺点：

- 难以处理多个消费者的消费速率。

### pull-based system

优势：

- 消费者自行控制消费的速率
- broker能够更激进的batch消息，
相比于push-based实现，需要在立即推送和聚合推送之间做出取舍。

不足：

- 如果没有数据的话会产生大量的无用轮询。

    解决方案：通过 **long poll** 请求，来等待数据的到来。
    *（这里也可以等待到足够的数据来实现batch）*

### store-and-forward 生产者

指生产者写入数据到local log，当consumer从broker尝试拉取数据时，
broker从producer拉取数据。

对于kafka的使用场景，有时需要支持数千的生产者。
数据分布在这么多的磁盘上将会是一个维护上的灾难。

## 消费位置 consumer position

**追踪消息被消费到哪个位置是消息系统的一个很关键的性能点。**

通常，消费情况的元数据被存储在borker上。
这是符合直觉的的方案：

1. broker能知道消费的发生以及ack，并且对于单一server的实现，没有其他合适的地方来存储。
1. 传统的消息系统没有很好的可伸缩性，broker可以将消费过的数据清除来保持较小的存储。

问题：如何在broker和consumer之间对于消费情况达成一致是一个不小的问题。

发送即标记消费实现最多一次会丢消息。

通过接受ACK来实现至少一次成本比较高：

- 如broker要维护每个消息的状态机：发送、消费完成、锁（不重复消费）。
- 如果永远没有收到ack，如何处理。

### kafka的方案

每个topic划分为有序的若干分片，
对于一个consumer group，一个分片同时有且只有一个consumer会消费。

因此consumer position只会有一个数字，不管是存储还是ack都非常的方便。

同时有一个side benefit，
消费者可以主动的回滚consumer position，
重复消费之前消费过的数据。

## Offline Data Load

## Static Membership

**静态成员** 用来帮助在 consumer group rebalance 协议上的程序提升可用性。

rebalance protocol依赖于group coordinator为group中的成员分配id。
这个id是临时的，每次成员重启或者重新加入会产生变化。

对于消费者类型的服务，这种变化会导致消息的大量reassign。
对于有状态的服务，这种调整会导致local的优化失效，进而导致系统部分或全部不可用。

为了解决这种问题，设计了Static Member机制。

# 消息投递语义

- *At most once*
- *At least once*
- *Exactly once*

可以分割为两个问题：

1. pub消息的可靠性
1. 消费消息的保证

## 写入

当一个消息写入的分片返回了 **committed** ，
只要有一个拥有该分片副本的borker是*alive*的，
那么这个消息就不会丢失。

### producer没有收到提交message的response

从0.11.0.0之后，kafka支持了幂等发送，
producer可以重复发送消息，但不会产生多个消息记录。

broker为每个producer赋予一个ID，
生产者发送消息的时候，会为每个消息带上一个序列号。

也是从0.11.0.0之后，producer支持使用类似事务的语义，
将消息发送到多个topic分片中：
要么都写入成功，要么都失败。

要求不是那么高的场景下，
也可以配置等待若干时间用来提交消息。
也可以，完全的异步发送，或者等待到只有leader分片（相对于follower分片）
拥有这个消息。

## 消费

### At least once

消费完成之后，记录消费的offset。

### At most once

开始消费之前，就记录消费的offset。

### Exactly once

对于kafka内部的场景，消费一个topic然后写入到另一个topic。
那么可以利用kafka提供的事务机制。
将offset和产生的结果作为一个原子写入到topic中。

对于消费topic写入外部系统中，
也需要使得记录offset和输出结果原子化的执行。
通常的实现方式，比如两步提交。
作为替代，可以将offset和输出存储到同一个地方。

# 副本 Replication

kafka根据配置的参数来控制每个分片的副本数量。

副本机制的运行单元是topic的分片。

Followers和普通的消费者一样，
从leader消费消息，应用到自身的日志中。

在分布式系统中，为了自动处理失败，需要一个精确的“存活”定义。
在kafka中，是这样定义的：

1. broker维持一个与controller的活跃session来接受元数据更新。
    - kRaft集群中，需要周期性的与controller发送心跳。
    - Zookeeper集群中，间接通过broker在初始化与zookeeper的会话时建立的虚节点。
1. broker保持维护的作为follower的分片与leader不会落后太远。

使用**in sync**来表示**alive**和**failed**之间的状态。
kafka维护一个ISR(in-sync replica)集合，
如果没有维持活跃session或者落后太多，那么这个broker就会被从ISR中移除。

**committed的含义**：
（如果消息是committed的，）
不需要担心如果leader fail之后，会丢失消息。
生产者需要权衡latency和可靠性来选择是否要等待message成为committed。

kafka保证committed的消息，在有至少一个in sync副本的情况下，
不会丢失。

kafka不解决网络分区问题。

## 副本log Quorums, ISRs, and State Machines

副本log是kafka分片的核心。
在分布式数据系统中，副本日志是一个很常见的基本元素。
副本log可以被其他的系统以状态机风格来实现其他的分布式系统。

副本日志实现了将一系列数据的顺序达成共识的过程。
最简单的方式是选出一个leader来决断所有的输入。

### 如何选出拥有所有committed消息的新leader

当leader无法正常工作时，我们需要选出一个up-to-date的follower。
kafka提供的保证需要分片复制算法能够在leader无法工作之后，
保证选出的leader拥有所有committed的消息。

考虑，如果在写入等待a个副本返回写入成功，选举需要检查b个副本，
a和b中保证有overlap，这个过程就是Quorum。

一个常见的方法是在写入和选举是都等待/检查超过半数的副本（以下称为过半选举）。
特别的，这种方法有一个天然的优势：
操作取决于latency最小的若干实例。
实现这种检查，有若干算法及其变种：ZAB, kRaft etc。

**过半选举**的不足在于只能容忍较少的节点失败或者需要更大的存储成本。

实践上，kafka选择了一个权衡之后的方案。
相比于过半选举，kafka动态维护了一个称作ISR(in-sync replicas)的集合来作为
quorum set。
ISR中的成员的数据与leader保持同步，也只有ISR中的成员有资格被选为新的leader。
所有写入只有在被所有的ISR副本处理才被认为是committed。
ISR成员维护在集群元数据中。
通过ISR模型和f+1个（ISR）副本，
kafka topic可以容忍f个节点失败而不丢失committed的消息。

ISR模型与过半选举在支持相同失败容忍度的时候，写入需要等待相同数量的副本的ack。
对于过半选举的只依赖低延迟实例的天然优点，
kafka选择通过让producer来决定是否等待follower ack来解决。
ISR模型的优势在于节省了很多带宽和存储，是值得的。

## kafka不要求崩溃的节点将所有的数据带回来

不同于普通的副本算法，kafka支持基于不可靠的存储运行。

实践中，磁盘错误是非常常见的情况。
另外，为了提高性能，kafka无法容忍每次都调用fsync来刷盘。

## 所有ISR节点失败后会发生什么？

kafka提供了两个对于一致性和可用性不同倾向的方案：

1. 等待任意ISR节点重新接入。
1. 选择一个ISR外的节点作为leader。

## 可用性与持久性

写入时，producer可以选择等待0、1、所有的（ISR）副本返回ack。

提供两个选项来提升持久性：

1. 禁用使用ISR外的节点作为主。
1. 设置最小ISR size。

## 副本管理

因为有很多副本，如果在broker之间进行选主会带来性能困扰。
许哦咦kafka选择使用controller进行选主，
controller自身失败，会自行选主。

# 日志压缩

TBD

# Quotas

TBD

# 参考

- [kafka DESIGN](https://kafka.apache.org/documentation/#design)
- [The Pathologies of Big Data](https://queue.acm.org/detail.cfm?id=1563874)
- [Notes from the Architect](http://varnish-cache.org/docs/trunk/phk/notes.html)
- [一个关于磁盘的介绍](https://zhuanlan.zhihu.com/p/534821258)

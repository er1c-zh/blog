---
title: "kafka设计"
date: 2021-09-30T20:08:12+08:00
draft: true
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

# 消费者

# 参考

- [kafka DESIGN](https://kafka.apache.org/30/documentation/#design)
- [The Pathologies of Big Data](https://queue.acm.org/detail.cfm?id=1563874)
- [Notes from the Architect](http://varnish-cache.org/docs/trunk/phk/notes.html)
- [一个关于磁盘的介绍](https://zhuanlan.zhihu.com/p/534821258)

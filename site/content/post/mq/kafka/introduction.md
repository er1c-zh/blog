---
title: "kafka基本概念"
date: 2021-09-30T10:39:50+08:00
draft: false
tags:
    - mq
    - kafka
    - memo
    - what
    - way-to-kafka
order: 1
---

关于kafka的基本概念。

<!--more-->

{{% serial_index way-to-kafka %}}

# 概念

kafka是一个**事件流平台** *(streaming platform)* 。
**事件流**是对现实中不断发生的“事件”的一种抽象。
技术上来说，
kafka可以从数据库、传感器、web服务等各种数据源捕获实时产生的**事件**，
并持久化这些**事件**用于后续在需要的时间和位置（比如某些服务器）消费与处理。

为了实现这些目标，
kafka由**Servers**和**Client**两部分构成。
kafka（的服务端或“中心”）由可以部署在多个机房的多个**server**集群构成。
其中的一些**server**作为存储层 *(storage layer)*，被称作**broker**。
其他的**server**作为运行Kafka Connect从存量的系统中输入输出事件流的数据。
Client提供了在分布式的应用上读写事件流的能力，
具有并行、可伸缩、容忍网络问题或机器问题导致的失败的特性。

**event**，事件，
kafka中存储的个体，
也会被称作 **message** 或 **record** ，
由`key`/`value`/`timestamp`/`optional metadata`组成。

**生产者**是向kafka写入事件的客户端；
**消费者**反之，是读取，或者说订阅，的客户端。

**topic** 是事件被组织、持久化存储的地方；
**topic** 与事件之间类似文件夹与文件的关系。
组织体现在生产者消费者或生产者订阅者会通过 **topic** 来写入或消费事件。
持久化存储体现在，与其他消息队列不同的，事件在消费之后不会被清理，
而会根据配置来存储一定的时间，在消息存续的时间中，可以再次消费，查看。
性能与存储的数据大小没有直接的关系，所以长时间的存储消息是可以接受的。

topic是**分片的**。
一个topic被分割为若干个**bucket**放置在不同的broker上。
分片为kafka提供了伸缩的能力，
因为客户端能同时从多个分片上来读写数据。
事件根据它的`key`会被且仅会写入到一个具体的分片上，
相同`key`的信息会被写入到同一分片上。
特别的，
kafka保证一个给定的分片的消费者读取到的事件的顺序与事件被写入的顺序相同。

为了实现数据的高可用与容灾，
可以为topic创建**副本** *（topic can be replicated）* 。
实现上，一份数据会被存储到不同的broker实例上。

# 参考

- [kafka Introduction](https://kafka.apache.org/intro)

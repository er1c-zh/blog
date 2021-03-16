---
title: "Redis Cluster练习"
date: 2021-03-11T11:09:02+08:00
draft: true
tags:
  - redis
  - how
  - redis-cluster-step-by-step
order: 1
---

哇哦，终于可以学习、实践一下Redis Cluster了！

<!--more-->

{{% serial_index redis-cluster-step-by-step %}}

# Redis的集群与Redis Cluster

单实例满足不了业务需求的时候，
一个解决方案是通过集群来分担压力。

`Redis Cluster`是官方提供的集群化方案，
能够在网络分区的情况下，提供一定程度的可用性 *(availability)* 。

除了RC之外，还有如[Codis](https://github.com/CodisLabs/codis)等实现方案。

主要会介绍如下的几个方面的偏向实践角度的观点：

- 对RC能够提供的服务的感性认识
- 简化的集群工作原理 
- 一致性相关的信息

下面的内容更多的是对[Redis cluster tutorial](https://redis.io/topics/cluster-tutorial)的~~机械~~翻译。

## Redis Cluster能提供什么功能？

1. 自动化的分配数据到不同节点

    这是我们选择集群的根本原因，
    通过分配功能、存储压力到不同的节点，来提高对外提供服务的能力。

1. 在网络分区的情况下，提供一定程度的可用性

    RC能够在网络分区或某些节点无法正常提供服务的情况下，继续提供服务。

## RC的实现思路

### 节点间的通信

首先是节点之间的通信，
RC中的实例拥有两个端口，分别是普通的、向客户端提供服务的和用于集群间通信的端口。

利用“通信的端口”的节点间信道被称作`集群总线`，
各个节点之间在这个总线上，
利用二进制协议来完成探活、配置更新、故障恢复鉴权等维护集群正常的功能需要的信息交流。

### 分片

采用hash槽的方式来进行分片。
将key的CRC16模16384作为分片的id，
特别的，可以使用相同的`hash tag`来令不同的key保存在同一个槽上。

利用`hash tag`的特性，可以一定程度的支持操作多个key的指令。

### 主从模式保证了可用性

通过给每个hash槽配备主节点和若干副本来实现在主节点无法正常工作时提升从节点来保证可用。

如果一个hash槽的所有实例都失效了，那么就无法正常工作。

## RC一致性相关的信息

RC不提供强一致性 *(strong consistency)* 保证。
在某些场景下，即便服务器返回了写入的ACK，但数据还是会丢失。

写入到集群的操作流程大概是：

1. client向master写入数据，master返回ack
1. master向从节点广播这个写入
1. 从节点写入这个修改

产生丢失的一个情况是：
如果在 “master向从节点广播这个写入” 的时候，master发生了crash。
参考数据库的redo log，可以将写入提前到回应成功来避免这种情景下的丢失。

另一个会产生丢失的场景是当一个主实例与其他的节点发生网络分区，无法通信。
在该失效的实例发生问题到发现自己无法正常工作之间的写入，
都无法被访问另一分区的客户端读取到，且如果没有特别的处理，这部分数据会发生丢失（因为没有复制到新的主节点）。
特别的，这个时间叫做**node timeout**。

# 启动一个测试用的Redis Cluster集群

以一个三主三从的集群为例子。


# 参考

- [Redis cluster tutorial](https://redis.io/topics/cluster-tutorial)
- [Redis cluster specification](https://redis.io/topics/cluster-spec)

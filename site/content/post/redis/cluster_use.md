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


# 参考

- [Redis cluster tutorial](https://redis.io/topics/cluster-tutorial)

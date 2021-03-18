---
title: "Redis Cluster初始化过程"
date: 2021-03-18T11:38:54+08:00
draft: true
tags:
  - redis
  - how
  - redis-cluster-step-by-step
order: 2
---

之前学习了RC的基本操作，
接下来我打算从集群的启动指令开始，
一步步了解源码层面的实现。

这一篇暂定是包括实例启动集群模式的过程。

<!--more-->

{{% serial_index redis-cluster-step-by-step %}}

# 引子

启动集群的时候，利用了指令：

```shell
redis-server --protected-mode no --daemonize yes --port 7002 \
--cluster-enabled yes --cluster-config-file nodes2.conf --cluster-node-timeout 5000 --appendonly yes
```

其中的`--cluster-enabled yes`显然是作为集群模式启动的关键。

来看源码：

```c
void initServer(void) {
    // ...
    if (server.cluster_enabled) clusterInit();
    // ...
}
```

初始化函数`initServer`中，判断`cluster_enabled`来选择是否执行`clusterInit()`。
这和我们之前的猜测是一致的，继续来看下`clusterInit()`做了什么工作。

# 一切开始的地方`clusterInit()`

`clusterInit`初始化了实例的集群相关的功能。

1. 首先初始化了`server.cluster`，一个`clusterState`对象，
里面存储了集群的状态、节点、失败转移(failover)等信息。
1. 尝试加载配置文件(`clusterLoadConfig(server.cluster_configfile)`)，
如果不存在的话，为自身新建一个节点(`createClusterNode()`)并保存配置文件(`clusterSaveConfigOrDie()`)。
1. 固定的，计算`配置的端口+10000`作为`cluster bus`通信的端口，检查是否合法。
1. 监听每个监听的地址(bind addr)的通信端口，并向事件循环注册处理函数`clusterAcceptHandler`。
1. 最后做一些清理的工作，完成初始化。

# 集群，启动！

初次运行，在启动了若干实例后，还需要如下指令来构建集群：

```shell
redis-cli --cluster create 127.0.0.1:7000 127.0.0.1:7001 \
    127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
    --cluster-replicas 1
```

构建集群由redis-cli和服务端合作完成。

客户端接收到`create`指令之后，具体操作由`clusterManagerCommandCreate()`完成。
这一篇的关注点更多的偏向于服务器端，
所以直接给出简单结论，客户端有如下工作：

1. 解析实例的地址，检查是否存活。
1. 根据参数计算出主节点、从节点、拓扑关系和hash槽的分配情况。
1. 



















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
    1. 最后会由`clusterReadHandler`来处理链接的数据。
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
1. 打散、重排实例，优化集群拓扑结构。
1. 根据参数计算出主节点、从节点、拓扑关系和hash槽的分配情况。
1. 为每个实例设置epoch、调用`MEET`命令。
1. 一些清理和检查的工作。

可以看到与服务端交互的操作有两个：

1. `cluster SET-config-epoch`设置epoch。
1. 调用`cluster MEET`命令。

回到服务端的代码，可以看到`cluster`指令都是由`clusterCommand(client *c)`来处理的。
这里先关注涉及到的两个子指令。

## `cluster MEET`

> "MEET <ip> <port> [bus-port] -- Connect nodes into a working cluster."

该指令用于将节点联入集群中，接受的参数是实例的地址和端口。

实现上首先解析了参数，然后调用`clusterStartHandshake`来继续后续的工作。

```c
void clusterCommand(client *c) {
    //...
    if (clusterStartHandshake(c->argv[2]->ptr,port,cport) == 0 &&
                errno == EINVAL)
    //...
}
```

1. 首先检查了ip地址和端口的合法性并格式化。
1. 检查目标节点是否在握手中。
1. 保存新节点的信息。

```c
int clusterStartHandshake(char *ip, int port, int cport) {
    clusterNode *n;
    char norm_ip[NET_IP_STR_LEN];
    struct sockaddr_storage sa;

    /* IP sanity check */
    // ...
    /* Port sanity check */
    // ...
    // 格式化地址

    // 检查是否目标节点是否已经在握手中了
    if (clusterHandshakeInProgress(norm_ip,port,cport)) {
        errno = EAGAIN;
        return 0;
    }

    // 一个新的节点，那么就存储信息，并将信息加入到集群的dict中
    n = createClusterNode(NULL,CLUSTER_NODE_HANDSHAKE|CLUSTER_NODE_MEET);
    memcpy(n->ip,norm_ip,sizeof(n->ip));
    n->port = port;
    n->cport = cport;
    clusterAddNode(n);
    return 1;
}
```

看到这里，属实令人一头雾水，怎么就结束了呢，集群的信息并没有交互起来哇。
redis除了接收到命令响应式的处理之外，还有时间驱动的事件，
会不会是有定时任务来完成后续的工作呢？

## 新节点加入集群的广播

果然，在`serverCron`中，如果实例是集群模式运行的，就会调用`clusterCron`。
继续看下去，不出意外，有遍历`server.cluster->nodes`检查链接的循环。
`cluster MEET`未竟的事业便是在这里完成的。

```c
void clusterCron(void) {
    // ...
    di = dictGetSafeIterator(server.cluster->nodes);
    server.cluster->stats_pfail_nodes = 0;
    while((de = dictNext(di)) != NULL) {
        clusterNode *node = dictGetVal(de);

        // 检查节点的状态

        if (node->link == NULL) { // 没有建立链接
            clusterLink *link = createClusterLink(node); // 创建一个clusterLink对象
            link->conn = server.tls_cluster ? connCreateTLS() : connCreateSocket(); // 根据配置创建不同的链接
            connSetPrivateData(link->conn, link); // 准备数据
            if (connConnect(link->conn, node->ip, node->cport, NET_FIRST_BIND_ADDR,
                        clusterLinkConnectHandler) == -1) { // 链接目标节点的控制端口，回调函数处理
                // 出现异常的处理
                continue;
            }
            node->link = link;
        }
    }
    // ...
}
```

建立链接的回调由`clusterLinkConnectHandler`处理，
工作比较简单：

1. 注册读取信息的回调 `clusterReadHandler`，
和监听自身的控制端口的回调函数是同一个。
1. 发起ping-pong `clusterSendPing(CLUSTER_TYPE_PONG)`































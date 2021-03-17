---
title: "Redis Cluster练习"
date: 2021-03-11T11:09:02+08:00
draft: false
tags:
  - redis
  - how
  - redis-cluster-step-by-step
order: 1
---

哇哦，终于可以学习、实践一下Redis Cluster了！

这一篇主要是简单的概念与集群搭建与简单操作。

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

我打包了一个docker镜像，可以直接使用。
出于简化考虑，将所有实例放在一起了，避免处理多个容器互联的问题。

```shell
docker run -dp 7000-7005:7000-7005 --name redis_cluster er1cz/redis_cluster
```

镜像暴露了7000-7005六个端口，分别是六个实例的端口，然后在本机可以使用redis-cli正常操作。

```shell
redis-cli -c -p 7000
127.0.0.1:7000> get key1
-> Redirected to slot [9189] located at 127.0.0.1:7001
(nil)
127.0.0.1:7001> set key1 value1
OK
127.0.0.1:7001> get key1
"value1"
127.0.0.1:7001>
```

上述命令，`-c`指明客户端使用集群模式，连接7000端口的服务器。
后续执行了读写操作，没有任何问题。

## 执行reshard操作

执行`redis-cli --cluster reshard 127.0.0.1:7000`开始重新分片。

1. 首先会要求输入多少个hash槽，这里为了方便演示选择了一个。
1. 然后需要输入哪个节点需要接受这些槽，这里输入了7000端口上的实例的id。
1. 然后是从哪些节点移除槽，可以一个一个的输入节点id，随后输入done；
也可以输入all选择所有节点来输出，这可以用在有新节点加入的情况。这里选择了7002端口的实例来提供槽。
1. 最后redis会提供一个移动方案，输入yes来开始执行reshard操作。

```shell
redis-cli --cluster reshard 127.0.0.1:7000
>>> Performing Cluster Check (using node 127.0.0.1:7000)
M: 13c73f2de3d6a2ac4fd4d99d359a9bd3693a7e5f 127.0.0.1:7000
   slots:[0-5460] (5461 slots) master
   1 additional replica(s)
S: 92dccaed8da01ceaea3e711e456ca76f78c0e55e 127.0.0.1:7005
   slots: (0 slots) slave
   replicates 080227d20bfb70af6603d68c4a94e3bd73050142
S: b6c39a0699dd33247dc8d94f21030f5fecedac76 127.0.0.1:7003
   slots: (0 slots) slave
   replicates 977028968d051b051ddb094a87cf3f71d37b9b75
M: 977028968d051b051ddb094a87cf3f71d37b9b75 127.0.0.1:7002
   slots:[10923-16383] (5461 slots) master
   1 additional replica(s)
S: cbddaf7e40f2cc3651a9ec62493efe37aee85506 127.0.0.1:7004
   slots: (0 slots) slave
   replicates 13c73f2de3d6a2ac4fd4d99d359a9bd3693a7e5f
M: 080227d20bfb70af6603d68c4a94e3bd73050142 127.0.0.1:7001
   slots:[5461-10922] (5462 slots) master
   1 additional replica(s)
[OK] All nodes agree about slots configuration.
>>> Check for open slots...
>>> Check slots coverage...
[OK] All 16384 slots covered.
How many slots do you want to move (from 1 to 16384)? 1
What is the receiving node ID? 13c73f2de3d6a2ac4fd4d99d359a9bd3693a7e5f
Please enter all the source node IDs.
  Type 'all' to use all the nodes as source nodes for the hash slots.
  Type 'done' once you entered all the source nodes IDs.
Source node #1: 977028968d051b051ddb094a87cf3f71d37b9b75
Source node #2: done

Ready to move 1 slots.
  Source nodes:
    M: 977028968d051b051ddb094a87cf3f71d37b9b75 127.0.0.1:7002
       slots:[10923-16383] (5461 slots) master
       1 additional replica(s)
  Destination node:
    M: 13c73f2de3d6a2ac4fd4d99d359a9bd3693a7e5f 127.0.0.1:7000
       slots:[0-5460] (5461 slots) master
       1 additional replica(s)
  Resharding plan:
    Moving slot 10923 from 977028968d051b051ddb094a87cf3f71d37b9b75
Do you want to proceed with the proposed reshard plan (yes/no)? yes
Moving slot 10923 from 127.0.0.1:7002 to 127.0.0.1:7000:
```

然后执行`redis-cli --cluster check 127.0.0.1:7000`查看集群信息，
可以看到7000端口的实例的槽从5461个变成了5462个，7002端口的实例则减少了一个。

```shell
redis-cli --cluster check 127.0.0.1:7000
127.0.0.1:7000 (13c73f2d...) -> 0 keys | 5462 slots | 1 slaves.
127.0.0.1:7002 (97702896...) -> 0 keys | 5460 slots | 1 slaves.
127.0.0.1:7001 (080227d2...) -> 1 keys | 5462 slots | 1 slaves.
[OK] 1 keys in 3 masters.
0.00 keys per slot on average.
>>> Performing Cluster Check (using node 127.0.0.1:7000)
M: 13c73f2de3d6a2ac4fd4d99d359a9bd3693a7e5f 127.0.0.1:7000
   slots:[0-5460],[10923] (5462 slots) master
   1 additional replica(s)
S: 92dccaed8da01ceaea3e711e456ca76f78c0e55e 127.0.0.1:7005
   slots: (0 slots) slave
   replicates 080227d20bfb70af6603d68c4a94e3bd73050142
S: b6c39a0699dd33247dc8d94f21030f5fecedac76 127.0.0.1:7003
   slots: (0 slots) slave
   replicates 977028968d051b051ddb094a87cf3f71d37b9b75
M: 977028968d051b051ddb094a87cf3f71d37b9b75 127.0.0.1:7002
   slots:[10924-16383] (5460 slots) master
   1 additional replica(s)
S: cbddaf7e40f2cc3651a9ec62493efe37aee85506 127.0.0.1:7004
   slots: (0 slots) slave
   replicates 13c73f2de3d6a2ac4fd4d99d359a9bd3693a7e5f
M: 080227d20bfb70af6603d68c4a94e3bd73050142 127.0.0.1:7001
   slots:[5461-10922] (5462 slots) master
   1 additional replica(s)
[OK] All nodes agree about slots configuration.
>>> Check for open slots...
>>> Check slots coverage...
[OK] All 16384 slots covered.
```

## 测试节点失效

我首先写入`key1`到集群中，发现被存储在7001实例上。

```shell
127.0.0.1:7000> get key1
-> Redirected to slot [9189] located at 127.0.0.1:7001
"value1"
```

查看拓扑结构，可以看到7005是7001的从节点。

```shell
redis-cli --cluster check 127.0.0.1:7000
# ...
127.0.0.1:7001 (080227d2...) -> 1 keys | 5462 slots | 1 slaves.
[OK] 1 keys in 3 masters.
0.00 keys per slot on average.
S: 92dccaed8da01ceaea3e711e456ca76f78c0e55e 127.0.0.1:7005
   slots: (0 slots) slave
   replicates 080227d20bfb70af6603d68c4a94e3bd73050142
# ...
M: 080227d20bfb70af6603d68c4a94e3bd73050142 127.0.0.1:7001
   slots:[5461-10922] (5462 slots) master
   1 additional replica(s)
[OK] All nodes agree about slots configuration.
>>> Check for open slots...
>>> Check slots coverage...
[OK] All 16384 slots covered.
```

接下来利用`redis-cli -p 7001 debug segfault`来模拟7001实例失效。

```shell
redis-cli -p 7001 debug segfault
Error: Server closed the connection
```

然后在去查询`key1`，可以看到，现在是由7005实例来提供服务。

```shell
redis-cli -p 7000 -c
127.0.0.1:7000> get key1
-> Redirected to slot [9189] located at 127.0.0.1:7005
"value1"
127.0.0.1:7005>
```

这个时候如果启动7001实例，会看到实例自动接入到集群，并成为7005的从节点。

```shell
# 在容器中
./redis/src/redis-server --protected-mode no --daemonize yes --port 7001 --cluster-enabled yes --cluster-config-file nodes1.conf --cluster-node-timeout 5000 --appendonly yes
# 检查集群
redis-cli --cluster check 127.0.0.1:7000

# ...
M: 92dccaed8da01ceaea3e711e456ca76f78c0e55e 127.0.0.1:7005
   slots:[5461-10922] (5462 slots) master
   1 additional replica(s)
#...
S: 080227d20bfb70af6603d68c4a94e3bd73050142 127.0.0.1:7001
   slots: (0 slots) slave
   replicates 92dccaed8da01ceaea3e711e456ca76f78c0e55e
#...
```

## 添加新的节点到集群中

添加节点到集群中可以用：

```shell
redis-cli --cluster add-node 127.0.0.1:7006 127.0.0.1:7000
```

默认的，会将新节点当作master节点，但没有分配槽，可以通过reshard来手动的分配。

如果想将新节点作为从节点，那么可以增加参数：

```shell
redis-cli --cluster add-node 127.0.0.1:7006 127.0.0.1:7000 --cluster-slave
```

也可以在增加参数指定主节点：

```shell
redis-cli --cluster add-node 127.0.0.1:7006 127.0.0.1:7000 --cluster-slave --cluster-master-id $CLUSTER_MASTER_ID
```

## 移除节点

```shell
redis-cli --cluster del-node 127.0.0.1:7000 $CLUSTER_WAIT_REMOVE_ID
```

特别的，如果是主节点，那么需要清空后才可以移除掉。

# 参考

- [Redis cluster tutorial](https://redis.io/topics/cluster-tutorial)
- [Redis cluster specification](https://redis.io/topics/cluster-spec)

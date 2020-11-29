---
title: "有序集合的实现"
date: 2020-11-28T10:57:19+08:00
draft: false
tags:
  - redis
  - what
  - how-redis-work
order: 5
---

redis有序集合相关指令的具体实现。

<!--more-->

{{% serial_index how-redis-work %}}

分类按照命令表的flag来确定，有序集合的是包含`@sortedset`标签的指令。

*基于redis 6.0*

# 数据结构

在数据量较小的时候，底层由压缩列表来承载数据；
数据量较大时，使用`zset`类型来承载数据。
数据大小的判断由参数`zset_max_ziplist_value`来决定。

[这里]({{< relref "data_struct.md#压缩列表" >}})有压缩列表结构的简单介绍。

先来看看`zset`的实现。
`zset`由一个哈希表和一个跳表组成，
和那道“O(1)读取的链表”的面试题很像，
哈希表用于搜索，跳表用于遍历。
```c
typedef struct zset {
  dict *dict; // 哈希表
  zskiplist *zsl; // zset版本的跳表
} zset;
```

哈希表可以参考[这篇文章]({{< relref "data_struct.md#哈希表" >}})，这里就不赘述了。

跳表也可以参考[这里]({{< relref "data_struct.md#跳表" >}})。

最后再来看创建的代码就比较清晰了：

```c
// 创建zset的数据结构
// src/t_zset.c#zaddGenericCommand#1603
if (server.zset_max_ziplist_entries == 0 ||
  server.zset_max_ziplist_value < sdslen(c->argv[scoreidx+1]->ptr))
{
  // 数据量较大时，直接创建zset类型
  zobj = createZsetObject();
} else {
  // 如果要插入的元素数量小于利用压缩列表的数量上限
  // 创建一个压缩列表来承载数据
  zobj = createZsetZiplistObject();
}
dbAdd(c->db,key,zobj);
```

# 普遍的实现思路

有序集合在元素容量较小时，采用压缩列表来承载；
反之，由上述的`zset`来存储数据。

两者其实类似，都有一个链表式的结构来维护顺序。
为了优化元素较多时查询的性能，增加了一个哈希表，
为了优化元素较多时定位的性能，使用跳表来降低时间复杂度。

# `ZADD`/`ZINCRBY`

两者均由`zaddGenericCommnad`实现。

首先检查解析参数，随后获取`zset`实例，如果不存在则新建一个。

一切就绪后，遍历要加入的元素，调用`zsetAdd`来插入到集合中。

`zsetAdd`中处理了检查元素是否存在、若干flag的检查，以及插入操作。
类似集合，当元素的数量超过阈值时，
会通过`zsetConvert`将底层数据结果由压缩列表转换成`zset`实现。

# `ZREM`

通过`zsetDel`删除元素，如果删除后为空，类似列表、集合，会将有序集合的key删除。

# `ZUNIONSTORE`/`ZINTERSTORE`

`zunionInterGenericCommand`

类似集合的交并操作，会先将若干集合按照大小进行排序来避免不必要的运算量。

对于交集，用数量最小的集合来驱动，若一个元素在每个集合中都出现了，
按照权重和聚合函数来计算新的score，
将结果加入到目标有序集合。

对于并集，会先遍历每个集合，增加到辅助哈希表中，
这个过程中会计算新的得分。
在遍历完成后，开始将辅助哈希表中的元素填充到目标有序集合中。

# `ZRANGE`/`ZREVRANGE`

`zrangeGenericCommand`，根据底层类型的不同，进行了相应的操作。

压缩列表比较简单，根据正/反，先调用`ziplistIndex`定位，
然后调用`zzlNext`/`zzlPrev`来进行遍历。

类似的，`zset`类型的利用`zslGetElementByRank`来进行定位，
随后使用跳表的`forward`/`backward`指针进行遍历。

# `ZRANGEBYSCORE`/`ZREVRANGEBYSCORE`

与`ZRANGE`的实现思路类似。

`genericZrangebyscoreCommand`首先会利用`zslParseRange`解析指令中的分数范围，
随后根据数据结构的不同采取相应的操作来遍历，并返回结果。
核心的逻辑还是通过range来查找对应范围的遍历的起点，随后常规的遍历直到分数不符合预期结束。

一个有意思的点是在扫描完成之前，
`ZRANGEBYSCORE`指令无法得知有多少元素需要返回。
所以在写入返回值时，
使用了`addReplyDeferredLen`来处理这种特殊情况，
等到最后确定长度后，调用`setDeferredArrayLen`来设置长度。

# `ZRANGEBYLEX`/`ZREVRANGEBYLEX`

与前面的range类型的指令相似。

需要指出几个不同或有意思的点是：

1. 字典序的规则是 A C a b c 这样，即大小写敏感的。
1. 如果有序集合中有多个分数，这个结果是无意义的。

# `ZREMRANGEBYRANK`/`ZREMRANGEBYSCORE`/`ZREMRANGEBYLEX`

由`zremrangeGenericCommand`实现。
分别是根据下标、分数、字典序来删除符合条件的数据。
移除的与前面的range函数查询到的结果一致，
这里就不赘述了。

# `ZCOUNT`/`ZLEXCOUNT`

`ZCOUNT`对于压缩列表，通过遍历实现；
`zset`会通过查找传入的分数范围定位到头尾两个节点，
根据span计算出两个节点中的元素数量。

字典序的命令实现类似。

# `ZCARD`

返回长度。

# `ZSCORE`

通过遍历（压缩列表）或哈希表查找（`zset`），来查询到节点，再获得分数。

# `ZRANK`/`ZREVRANK`

类似`ZSCORE`的方式，定位到节点，然后返回下标。

# `ZPOPMIN`/`ZPOPMAX`

落到`genericZpopCommand`。

核心实现还是很简单清晰的，
循环需要pop的个数，
循环中定位到头/尾，
写入客户端，
最后从有序集合中删除。

# `BZPOPMIN`/`BZPOPMAX`

落到`blockingGenericZpopCommand`。
如果由集合不为空，会调用`genericZpopCommand`来正常的执行。
如果集合为空，那么使用`blockForKeys`来等待。

[这里]({{< relref "data_struct.md#阻塞指令的实现" >}})有`blockForKeys`的简单分析。

# `ZSCAN`

最终落到`scanGenericCommand`上来完成。
可以看下这篇文章[scan相关的实现]({{< relref "how_redis_work_common.md#scan相关的实现" >}})。

# `SORT`

与列表的实现类似。可以参考[列表的`SORT`接口实现]({{< relref "how_redis_work_list.md#sort" >}})

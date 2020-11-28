---
title: "有序集合的实现"
date: 2020-11-28T10:57:19+08:00
draft: true
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

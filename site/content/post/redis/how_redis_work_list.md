---
title: "列表相关的实现"
date: 2020-11-22T23:26:48+08:00
draft: false
tags:
  - redis
  - what
  - how-redis-work
order: 3
---

redis列表相关指令的具体实现。

<!--more-->

{{% serial_index how-redis-work %}}

分类按照命令表的flag来确定，字符串类的是包含`@list`标签的指令。

*基于redis 6.0*

# 数据结构

讨论具体接口的实现之前，我想简单介绍下redis用来实现列表相关的数据结构。

可以从PUSH类的接口的具体的实现函数`pushGenericCommand`看到，
如果之前没有具体的值，那么会调用`createQuicklistObject`来创建一个快速列表对象。

大概的看了下，似乎没再使用压缩列表了。

对于快速列表可以看之前的[文章]({{< relref "data_struct.md#快速列表-quicklist" >}})有简单的介绍。

# `LPUSH`/`RPUSH`/`LPUSHX`/`RPUSHX`

两个不带X的指令实现从头/尾插入数据。
具体的实施由`pushGenericCommand`完成。

函数中首先查询是否有目标的值，
如果没有使用`createQuicklistObject`和`dbAdd`来插入。

随后遍历参数，依次调用`listTypePush`来插入对象。
函数的实现最终依赖快速列表的接口。

然后写入列表剩余的长度。

最后触发事件。

对于带X的两个指令基于`pushxGenericCommand`实现，
与前面两个指令的区别在于如果不存在队列，那么就不会创建，且直接返回0.

# `LINSERT`

**O(n)**

利用`listTypeInitIterator`获得列表遍历的迭代器，
迭代并用`listTypeEqual`与参数中的`pivot`比较，
查找到位置后利用`listTypeInsert`进行插入。

返回结果，-1没插入，成功返回列表长度。

# `LPOP`/`RPOP`/`BLPOP`/`BRPOP`

对于POP类型的指令分为阻塞与非阻塞两种。

非阻塞的基于`popGenericCommand`实现。

通过`listTypePop`返回结果，
如果为空，返回null，否则写入返回值并触发事件。

阻塞模式基于`blockingPopGenericCommand`实现。
函数按照目标列表有数据或无数据来采取不同的行为。
对于列表中有数据的情况，阻塞不再有意义，
便使用类似非阻塞的实现，特别的，处理过程会将指令改写为非阻塞的。
如果列表为空，就会利用`blockForKeys`来阻塞在列表上。

有意思的一个点是，列表为空时会删除这个键值对，
就可以按照等待这个键来实现阻塞。

# `RPOPLPUSH`/`BRPOPLPUSH`

非阻塞模式的指令，通过`listTypePop`获取一个数据。
然后用`rpoplpushHandlePush`来压到需要的列表。

`rpoplpushHandlePush`的内容比较清晰，如果目标列表不存在，那么创建，
然后利用`listTypePush`插入数据。最后触发事件，写入结果。

最后检查如果列表为空，那么删除键。

阻塞模式的实现与之前的`BLPOP`的实现类似，
依赖`blockForKeys`来实现阻塞，
区别在于会写入之前没有的`target`字段来表明插入需要放在哪个键中。

# `LLEN`

乏善可陈，直接返回列表长度。

# `LINDEX`

调用`quicklistIndex`来实现核心逻辑。

首先先找到在哪个节点：

```c
while (likely(n)) { // n是快速列表的节点
  if ((accum /* 之前遍历过的数量 */ + n->count /* 计算加上这个节点的值的数量 */ ) > index) {
    break; // 超过目标index，意味着目标在这个节点里
  } else {
    D("Skipping over (%p) %u at accum %lld", (void *)n, n->count,
      accum);
    accum += n->count; // 迭代计数
    n = forward ? n->next : n->prev; // 按照方向迭代
  }
}
```

然后在节点里查找：

```c
quicklistDecompressNodeForUse(entry->node); // 解压缩中间节点
// 在压缩列表中查找
// 实现比较清除，按照压缩列表的格式来处理
entry->zi = ziplistIndex(entry->node->zl, entry->offset);
```

这个指令的复杂度是O(n)，但是系数较小。

# `LSET`

指令核心由`quicklistReplaceAtIndex`实现，
首先用`quicklistIndex`确定快速列表的节点，
这个时候得到了对应的一个压缩表，
接下来就用`ziplistDelete`和`ziplistInsert`替换对应值。

# `LRANGE`

利用迭代器去遍历，这个过程没有什么特别的。

利用`listTypeInitIterator`生成从指定位置开始迭代的迭代器。
最终落到快速列表的`quicklistGetIteratorAtIdx`上：

```c
quicklistIter *quicklistGetIteratorAtIdx(const quicklist *quicklist,
                                         const int direction,
                                         const long long idx) {
    quicklistEntry entry;

    if (quicklistIndex(quicklist, idx, &entry)) { // 确定具体的节点
        // 构建迭代器
        quicklistIter *base = quicklistGetIterator(quicklist, direction);
        base->zi = NULL;
        base->current = entry.node;
        base->offset = entry.offset;
        return base;
    } else {
        return NULL;
    }
}
```

# `LTRIM`

核心是利用快速列表的`quicklistDelRange`接口移除前后的其他数据。

# `LPOS`

```redis
LPOS key element [RANK rank] [COUNT num-matches] [MAXLEN len]
```

- 返回element的位置
- 返回第rank个匹配的位置
- 返回num-matches个答案
- 扫描最多len个元素，0扫全表

实现显然可以简单的通过迭代器来完成。

# `LREM`

实现显然可以简单的通过迭代器来查找目标对象，
随后调用`listTypeDelete`移除。

# `SORT`

**O(N+Mlog(M))** N是列表中的元素数量，M是要返回的数量。

```redis
SORT key [BY pattern] [LIMIT offset count] [GET pattern [GET pattern ...]] [ASC|DESC] [ALPHA] [STORE destination]
```

- ALPHA 按照字典序排序
- LIMIT offset count 类似SQL，表示从offset开始，查询count个
- BY pattern 指令按照pattern和列表中元素的值，生成key，然后按照这个key存储的值作为排序的依据进行排序
- GET pattern 指令会返回按照pattern和列表中元素的值，生成key，然后按照排序的结果依序返回生成的key对应的值

首先解析请求参数。

随后按照不同的类型（列表/集合/有续集合），读取出容器中的数据。这部分有一些不需要排序情况下的fast-path，暂且略过。

获取容器中的数据后，按照不同的参数，去获取用来排序的值：
- 设置了`BY pattern`，使用`lookupKeyByPattern`来构建键并查找对应的值。

然后就是排序，对于要全部排序的采用`qsort`；
对于部分排序使用redis自己实现的`pqsort`
(可以看[另一篇文章]({{< relref "redis_pqsort.md" >}})简单分析了具体的实现)。

最后是返回结果。
没有制定`GET`的话，直接将排序值返回即可；
否则需要利用`lookupKeyByPattern`来查找需要的值。

## 参考

- [redis sort](https://redis.io/commands/sort)

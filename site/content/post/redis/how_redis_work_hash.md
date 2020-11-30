---
title: "哈希表的实现"
date: 2020-11-30T21:52:00+08:00
draft: false
tags:
  - redis
  - what
  - how-redis-work
order: 6
---

redis哈希表相关指令的具体实现。

<!--more-->

{{% serial_index how-redis-work %}}

分类按照命令表的flag来确定，哈希表的是包含`@hash`标签的指令。

*基于redis 6.0*

# 数据结构

哈希表的实现也是有两种数据结构，
在数据量小的时候，
是压缩列表，
数据量大的时候，就转化成哈希表。

是否转换有两种情况：
1. 元素的数量大于`hash_max_ziplist_entries`
    ```c
    for (i = start; i <= end; i++) {
      if (sdsEncodedObject(argv[i]) &&
        sdslen(argv[i]->ptr) > server.hash_max_ziplist_value) // 这里进行的比较
      {
        hashTypeConvert(o, OBJ_ENCODING_HT);
        break;
      }
    }
    ```
1. 元素的长度大于`hash_max_ziplist_value`
    ```c
    if (hashTypeLength(o) > server.hash_max_ziplist_entries)
      hashTypeConvert(o, OBJ_ENCODING_HT);
    ```

这两种情况也很合理，避免了一个元素过长导致遍历的时间长的问题。

两种数据结构的简单介绍可以看下[压缩列表]({{< relref "data_struct.md#压缩列表" >}})或[哈希表]({{< relref "data_struct.md#哈希表" >}})。

**对于压缩列表，存储是按照键、值、键、值这样存储的。**

# `HSET`/`HMSET`

哈希表的实现比较简单，
首先查找目标key对应的`entry`，
如果没有找到，就简单的增加对应的`entry`。

压缩列表要复杂些，
首先要遍历来尝试找出是否有目标键值对，
如果没有，那么就在尾部增加键、值；
如果有，需要将原来的值删掉，然后将新值插入到旧值的地方。
压缩列表的插入操作十分的繁杂，需要考虑空间、节点之间的关系等。

通过循环完成多个键值的插入。

# `HSETNX`

先检查键值是否不存在，然后调用`HSET`用过的`hashTypeSet`插入。

# `HGET`/`HMGET`

没有什么特别的地方，简单的调用对应数据结构的get方法，返回结果。

批量版本的通过循环来完成。

# `HINCRBY`/`HINCRBYFLOAT`

是`INCRBY`的哈希表版，通过组合`HGET`、加法、`HSET`来完成。

浮点数版本类似。

# `HDEL`

实现上就是通过调用对应的删除函数来删除键值对。
对于删除后为空的情况，会把整个哈希表移除。

# `HLEN`

返回长度，没有什么特别的。

# `HSTRLEN`

`STRLEN`的哈希表版本，查询到对应的值后，返回结果。

# `HKEYS`/`HVALS`/`HGETALL`

三者使用`genericHgetallCommand`来实现，
通过传入不同的flag来控制返回那些值。

遍历的过程比较简单，生成对应的迭代器，然后遍历即可。

# `HEXISTS`

查找对应的key是否存在，实现上利用了get类型的函数，根据是否读取到返回结果。

# `HSCAN`

最终落到`scanGenericCommand`上来完成。
可以看下这篇文章[scan相关的实现]({{< relref "how_redis_work_common.md#scan相关的实现" >}})。

---
title: "列表相关的实现"
date: 2020-11-22T23:26:48+08:00
draft: true
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

# `LPOP`/`RPOP`

基于`popGenericCommand`实现。

通过`listTypePop`返回结果，
如果为空，返回null，否则写入返回值并触发事件。

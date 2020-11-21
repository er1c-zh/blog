---
title: "redis指令的实现-字符串"
date: 2020-11-11T16:59:20+08:00
draft: false
tags:
  - redis
  - what
  - how-redis-work
order: 1
---

redis字符串相关指令的具体实现。

<!--more-->

{{% serial_index how-redis-work %}}

分类按照命令表的flag来确定，字符串类的是包含`@string`标签的指令。

*基于redis 6.0*

*这个希望能坚持下来，做成一个系列，来简单介绍各个类型的指令的具体实现。*

# `GET`/`MGET`

*src/t_string.c#179*

对于`GET`，
核心实现为`getGenericCommand`方法。
首先调用了`lookupKeyReadOrReply`，获取需要的对象。
然后判断类型，如果不是`OBJ_STRING`，那么就返回错误(`addReply`)。
最后调用`addReplyBulk`返回结果。

`MGET`的实现类似，通过循环依次使用`lookupKeyRead`来查找、检查、返回结果。

# `SET key value [EX seconds|PX milliseconds|KEEPTTL] [NX|XX] [GET]`

*src/t_string.c#97*

首先使用若干if判断来解析`NX`和`EX`等参数，
随后调用`setGenericCommand`执行具体的操作：

1. 如果需要的话，解析过期时间(`getLongLongFromObjectOrReply`)。
1. 通过`lookupKeyWrite`来查询对应的key是否存在来检查是否符合`NX`或`XX`。
1. 调用`genericSetKey`来设置键值对。
1. 如果需要，调用`setExpire`设置过期时间。
1. 调用`notifyKeyspaceEvent`来通知事件。
1. 调用`addReply`返回结果。

# `SETNX`/`SETEX`/`PSETEX`

这些功能的具体实现都是调用`setGenericCommand`来实现的，
不同功能由传递的参数来实现。

# `MSET`/`MSETNX`

两者使用`msetGenericCommand`作为实现。
依次遍历输入的参数，（检查NX），使用`setKey`来设置值，并触发事件。

# `GETSET`

首先使用`getGenericCommand`获取并设置返回结果。
随后使用`setKey`来设置值（最后落到`genericSetKey`），最后触发相应的事件。

# `APPEND`

*src/t_string.c*

1. 首先查找是否存在目标key，如果不存在，直接调用`dbAdd`增加。
1. 检查查找到的内容，是否是合规，如果没有问题，那么调用`sdscatlen`来实现append。
1. 返回append后的长度。

# `STRLEN`

返回调用`stringObjectLen`来计算长度，数字返回十进制的长度。

# `SETRANGE`/`GETRANGE`/`SUBSTR`

对于`SETRANGE`来说，如果key不存在，或者offset超过原有的长度，会将前面的部分填0。 

`GETRANGE`和`SUBSTR`完全相同，两者使用`getrangeCommand`作为处理函数，
对于超过范围的范围，会规范到有意义的长度，见下面的代码。

```c
// t_string.c#279
if (start < 0) start = 0;
if (end < 0) end = 0;
if ((unsigned long long)end >= strlen) end = strlen-1;
```

# `INCR`/`DECR`/`INCRBY`/`DECRBY`

这些指令最终都由`incrDecrCommand`来实现，
其内部实现清晰：
1. 获取之前的值，如果没有则新建，检查参数与计算后的参数
1. 写入结果
1. 触发相应的事件

比较有意思的是在变更原有的值时，会检查有无其他地方引用，
如果有则会重新建立一个对象来存储相应的值。

# `INCRBYFLOAT`

`incrbyfloatCommand`具体实现类似`incrDecrCommand`。
比较有意思的一点是浮点数的精度问题，在主节点中利用语言的`long double`完成计算后，
将指令替换成`SET`，值设为计算的结果，以此来避免AOF重放或副本与正式格式的不一致。

# `STRALGO` string_algorithms

命令用于应用算法到字符串上，第一个参数用来表示使用什么算法。

目前（6.0）只支持最长公共子串算法。

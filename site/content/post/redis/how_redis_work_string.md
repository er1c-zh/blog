---
title: "redis指令的实现-字符串"
date: 2020-11-11T16:59:20+08:00
draft: true
tags:
    - redis
    - what
    - how-redis-work
---

redis字符串相关指令的具体实现。

分类按照命令表的flag来确定，字符串类的是包含`@string`标签的指令。

*基于redis 6.0*

*这个希望能坚持下来，做成一个系列，来简单介绍各个类型的指令的具体实现。*

- `GET key`

*src/t_string.c#179*

核心实现为`getGenericCommand`方法。
首先调用了`lookupKeyReadOrReply`，获取需要的对象。
然后判断类型，如果不是`OBJ_STRING`，那么就返回错误(`addReply`)。
最后调用`addReplyBulk`返回结果。

- `SET key value [EX seconds|PX milliseconds|KEEPTTL] [NX|XX] [GET]`

*src/t_string.c#97*

首先使用若干if判断来解析`NX`和`EX`等参数，
随后调用`setGenericCommand`执行具体的操作：

1. 如果需要的话，解析过期时间(`getLongLongFromObjectOrReply`)。
1. 通过`lookupKeyWrite`来查询对应的key是否存在来检查是否符合`NX`或`XX`。
1. 调用`genericSetKey`来设置键值对。
1. 如果需要，调用`setExpire`设置过期时间。
1. 调用`notifyKeyspaceEvent`来通知事件。
1. 调用`addReply`返回结果。


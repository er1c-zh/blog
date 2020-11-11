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

- `get`

*src/t_string.c#179*

核心实现为`getGenericCommand`方法。
首先调用了`lookupKeyReadOrReply`，获取需要的对象。
然后判断类型，如果不是`OBJ_STRING`，那么就返回错误(`addReply`)。
最后调用`addReplyBulk`返回结果。

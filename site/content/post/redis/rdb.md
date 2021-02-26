---
title: "redis RDB的实现"
date: 2021-02-07T22:00:12+08:00
draft: true
tags:
  - redis
  - how
---

*基于redis 6.0版本*

rdb是Redis Database Backup的缩写。

Redis Database Backup File是用来保存快照式的redis的数据的文件。

# 触发保存的时机

保存RDB由`rdbSave`实现。
`rdbSave`是一个同步阻塞的函数，
有一个异步版本包装函数`rdebSaveBacground`，
包装中尝试初始化一个通信用的`pipe`后利用`fork`机制开启一个进程来运行`rdbSave`。

```c
// src/rdb.c#L1314
int rdbSave(char *filename, rdbSaveInfo *rsi)
// src/rdb.c#L1382
int rdbSaveBackground(char *filename, rdbSaveInfo *rsi)
```

通过追踪两个函数的调用，可以发现有多个触发的机制：

1. 通过Redis的`SAVE`或`BGSAVE`指令来触发。
1. `FLUSHALL`指令也会执行。
1. `prepareForShutdown`中也会调用。
1. 注册在时间事件中的`serverCron`也会调用。

# RDB的机制

RDB的核心逻辑在`rdbSave`函数中实现，
`rdbSave`接受两个参数：
1. 一个字符串`filename`
1. 一个`rdbSaveInfo`的指针`rsi`

首先会尝试创建一个名为`temp-{{pid}}.rdb`的临时文件。

然后初始化`rio`对象(`rioInitWithFile`)，根据配置调整好`rio`。

> `rio`是一个面向流的简单抽象，提供了读写流的接口抽象。

一切准备就绪后，调用`rdbSaveRio`执行真正的保存操作。

`rdbSaveRio`执行成功之后，依次调用`fflush`/`fsync`/`fclose`来保证数据落盘。

最后调用`rename`将临时文件移动到`filename`完成整个流程。

## `rdbSaveRio`

`rdbSaveRio`有四个入参，分别是`rio`的实例、异常、flags和`rdbSaveInfo`实例。

准备的事情都在`rdbSave`中完成了，这里就是很纯粹的按照格式写入文件。

首先写入了生成该rdb文件的版本号和一些辅助字段和信息(`rdbSaveAuxField`/`rdbSaveModulesAux`)。

随后遍历所有的`db`：
1. 通过`dictGetSafeIterator`获得一个`dict`的迭代器
1. 写入选择数据库的指令
1. 写入数据库大小的指令和大小
1. 遍历所有所有的entry
  1. 调用`rdbSaveKeyValuePair`保存数据

最后写入结束的标记。

## `rdbSaveKeyValuePair`

数据存储的实现。

1. 如果有过期时间，保存过期时间
1. 保存LRU信息 *(IdleTime)*
1. 保存LFU信息
1. 保存键值对的值
  1. `rdbSaveObjectType` 保存值的类型
    通过一个switch保存了值的类型。
  1. `rdbSaveStringObject` 保存键
  1. `rdbSaveObject` 保存值，实现是一个长长长长长的多个if判断
    - 字符串就是正常的保存
    - 列表会遍历`QUICK_LIST`的节点，对于未压缩的节点就直接保存，否则会当作Blob保存
    - 对于SET，哈希表实现会通过迭代器遍历并保存，整数集合会当作Blob保存
    - 对于有序集合，也是类似的处理方式
    - 对于哈希表，也是该遍历遍历，该存成Blob就存成Blob

# 一些分析

## 后台存储线程和主线程同时访问数据不会产生并发问题吗？

这个非常的巧妙，通过`fork`出新的线程，linux会利用COW来避免无用的内存复制，
这相当于子线程创建了一个执行`fork`时的一个快照。

这样通过利用系统的机制来解决了并发与复制的问题，十分巧妙。

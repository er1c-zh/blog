---
title: "redis RDB的实现"
date: 2021-02-07T22:00:12+08:00
draft: false
tags:
  - redis
  - how
---

*基于redis 6.0版本*

对于redis的两种持久化方法听说很久了，
但一直没有认真的了解过，
所以找机会来学习下。

rdb是Redis Database Backup的缩写。

Redis Database Backup File是用来保存快照式的redis的数据的文件。

<!--more-->

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
1. 一个`rdbSaveInfo`的指针`rsi`，用于传递一些附加功能的参数

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

## 如何设计一个存储数据的格式

或者说，设计一个序列化数据的格式。
不光是简单的存储，比如经常使用的thrift和protobuf也是类似的。

首先是扩展性。在业务和功能不断扩展的情况下，很难说能够一次性设计好需要的格式。
扩展性包括向前兼容和向后兼容，分别指旧的程序支持后续生成的数据（向前兼容 Forwards Compatibility）
和新的程序支持旧的数据（向后兼容 Backwards Compatibility），
换句话说是（程序）向前/后兼容（数据）。

向后兼容是通常要做也可以做到的，而向前兼容则比较难。
如果要作向前兼容比较困难一点，特别是在一个不断发展的格式上。

退而求其次，在程序面对新的未知的格式时能优雅的做出响应是一个可以接受的处理方式。

然后是编解码的速度和数据大小。
不约而同的，很多编码格式都会把数据的长度放在存储开始的地方。
这样就能够顺序遍历一遍后便 **不多不少的** 读取到全部的数据。
而速度与产生的数据大小是跷跷板的两端，
相对的，如果不执行压缩，生成与读取的速度会便会但数据量会变大，反之亦然。

## RDB文件格式

可以通过分析读取RDB的函数来分析RDB文件的格式。

加载RDB文件由`rdbLoadRio`完成。

首先会读取9个字节，检查是否是`REDIS{{version}}`并判断version是否是本实例支持的版本。

随后就进入一个大循环，在这个大循环中完成了每一条记录的读取。

对于每一条记录，首先调用`rdbLoadType`来读取一个字节，
解释为无符号整数，业务含义是“类型”，标记接下来的数据的格式。

类型分为两大类分别是`RDB_OPCODE_*`存储非kv对的信息，和`RDB_TYPE_*`，存储kv对的数据。

一个巨大的if首先来判断是否是`OPCODE`型的记录，如果不是那么就认为是kv对数据。

### KV对数据

通过`rdbGenericLoadStringObject`读取字符串作为key，
通过`rdbLoadObject`读取对应的对象。

1. 读取key

    `rdbGenericLoadStringObject`会解析rdb中的数据，并按照入参解析成普通字符串、压缩字符串或数字。

    字符串存储先是一个变长的字节序列标记字符串的长度，`rdbLoadLenByRef`负责处理长度的读取。
    首先读取高2bit来判断表示长度的字节序列有多长，随后便按照对应格式来读取成长度。
    除此之外，`rdbLoadLenByRef`还会解析出字符串是否是经过编码。

    所谓的“编码”有数字和经过压缩的字符串两种情况，
    分别由`rdbLoadIntegerObject`和`rdbLoadLzfStringObject`来接管。

    数字的处理比较简单，就是读取、转换。

    `LzfString`的存储分为三部分，压缩长度、解压后长度和数据。
    首先读取两个长度，然后根据压缩长度读取数据，调用`lzf_decompress`来解压，最后根据flags来返回需要的对象。

    对于无编码的情况，根据入参来读取数据直接返回或构建一个sds对象或string对象。

1. 读取Object

    一个大if来根据type处理对应的数据类型。

    - RDB_TYPE_STRING

        字符串类型的对象解析与key的解析用的是同一个函数，这里就不再重复分析了。

    - RDB_TYPE_LIST

        要加载的格式是`quick_list`。
        首先是用过很多遍的`rdbLoadLen`读取到列表的长度。
        随后根据读取到的长度来继续加载数据，并解析，最后追加到列表中。

    - SET

        和列表十分的类似，会读取长度，加载数据，追加到数据结构中。
        特别的，会根据长度来选择是用数字集合还是哈希表来承载。

    - ZSET

        特别的，除了读取值之外，还会调用`rdbLoadBinaryDoubleValue`或`rdbLoadDoubleValue`来读取分数。

    - HASH

        没有什么特别的，与其他的实现思路类似。

    - LIST_QUICKLIST

        与LIST类似，不同在于使用了压缩列表

    - RDB_TYPE_HASH_ZIPMAP | RDB_TYPE_LIST_ZIPLIST | RDB_TYPE_SET_INTSET | RDB_TYPE_ZSET_ZIPLIST | RDB_TYPE_HASH_ZIPLIST

        这几个类型与前述的区别在于，读取到的是压缩过的字符串，随后会被解析到对应的类型。

### 指令

这里只记录几个的含义。

- `RDB_OPCODE_EXPIRETIME` `RDB_OPCODE_EXPIRETIME_MS`

    通过`rdbLoadTime`或`rdbLoadMillisecondTime`读取4字节解释为整数。

- `RDB_OPCODE_FREQ` `RDB_OPCODE_IDLE`

    用于LRU和LFU的数据。

- `RDB_OPCODE_EOF`

    结束的标志。

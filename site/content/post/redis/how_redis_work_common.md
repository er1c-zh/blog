---
title: "redis指令的实现-基础接口"
date: 2020-11-11T20:21:00+08:00
draft: true
tags:
    - redis
    - what
    - how-redis-work
---

这里主要用来放置数据库接口等基础接口的简单分析。

# 数据库提供的获取key的接口

首先看下基本的调用关系

```ditaa
@startuml
ditaa
                        +-----------+
            +---------->| lookupKey |<-----------+
            |           +-----------+            |
            |                                    |
+-----------+----------+             +-----------+-----------+
|lookupKeyReadWithFlags|             |lookupKeyWriteWithFlags|
+-----------+----------+             +-----------+-----------+
            ^                                    ^               
            |                                    |               
            |                                    |               
+-----------+----------+             +-----------+-----------+
|    lookupKeyRead     |             |     lookupKeyWrite    |
+-----------+----------+             +-----------------------+
            ^                                    ^               
            |                                    |               
            |                                    |               
+-----------+----------+             +-----------+-----------+
| lookupKeyReadOrReply |             | lookupKeyWriteOrReply |
+-----------+----------+             +-----------------------+
@enduml
```

最基本的实现是`lookupKey`。
函数调用了哈希表的接口`dictFind`，获得对应的`dictEntry`或如果不存在的话返回空。
如果key不存在，直接返回；
如果存在，首先通过`dictGetVal`宏获得对应的`robj`，然后根据逐出策略来更新LFU的相关统计或用于LRU的最后访问时间。

调用`lookupKey`的有三个函数：
- lookupKeyReadWithFlags
- lookupKeyWriteWithFlags
- scanDatabaseForReadyLists

首先来看下`lookupKeyReadWithFlags`，这是所有读取为目的查询接口的底层实现。

`lookupKeyReadWithFlags`首先调用了`expireIfNeeded`来判断key是否过期。
函数中首先判断是否过期。
如果过期了，对于从实例，会直接返回结果；
对于主实例，会：
  1. 更新统计数据
  1. 调用`propagateExpire`将变化传递给AOF和从实例
  1. 调用`notifyKeyspaceEvent`将过期事件进行广播
  1. 根据配置`lazyfree_lazy_expire`，会尝试异步(`dbAsyncDelete`)或同步(`dbSyncDelete`)的将这个key移除
      - 同步删除通过调用哈希表的`dictDelete`接口分别删除过期map和主map中的key，对于集群模式，还会调用`slotToKeyDel`来移除key和slot的关系
      - 异步删除整个流程类似与同步的删除，只是在判断大小如果足够大的时候，会异步释放，而不是同步释放。
  1. 调用`signalModifiedKey`触发相关的钩子

然后调用`lookupKey`来查找，最后返回结果。

然后看下`lookupKeyWriteWithFlags`，其实现非常简单，调用了`expireIfNeeded`检查过期情况，然后调用`lookupKey`来查找，返回。

还有一些更高级的接口，来提供一些封装，如：

- `lookupKeyRead`对`lookupKeyReadWithFlags`进行简单封装，提供了常用的查询
- `lookupKeyReadOrReply(client, key, reply)`如果查询不到就向客户端返回结果`reply`
- `lookupKeyWrite`和`lookupKeyWriteOrReply`类似read版本

# 返回结果

*src/networking.c*

还是先看下调用关系。

最基础的两个`_addReplyToBuffer`和`_addReplyProtoToList`，分别将数据直接写入buffer或者加入到client的`reply`链表中。

下面是细节一些的实现。


```c
// src/networking.c
/* -----------------------------------------------------------------------------
 * Low level functions to add more data to output buffers.
 * -------------------------------------------------------------------------- */

int _addReplyToBuffer(client *c, const char *s, size_t len) {
    size_t available = sizeof(c->buf)-c->bufpos; // 获取client的buf剩余长度

    if (c->flags & CLIENT_CLOSE_AFTER_REPLY) return C_OK; // 如果已经关闭了，直接返回

    // 如果reply列表不为空（即用了另一个接口），那么不能再用写入到buffer的这个接口
    if (listLength(c->reply) > 0) return C_ERR;

    // 如果buffer的剩余长度不足，那么返回错误
    if (len > available) return C_ERR;

    // 将数据拷贝过去
    memcpy(c->buf+c->bufpos,s,len);
    c->bufpos+=len;
    return C_OK;
}

void _addReplyProtoToList(client *c, const char *s, size_t len) {
    if (c->flags & CLIENT_CLOSE_AFTER_REPLY) return; // 如果已经关闭了，直接返回

    listNode *ln = listLast(c->reply); // 获取reply链表的尾部节点
    clientReplyBlock *tail = ln? listNodeValue(ln): NULL; // 如果没有插入过，那有可能为空

    /* Append to tail string when possible. */
    if (tail) {
        // 首先尝试去尽量填充到尾部节点的空间
        size_t avail = tail->size - tail->used;
        size_t copy = avail >= len? len: avail;
        memcpy(tail->buf + tail->used, s, copy);
        tail->used += copy;
        s += copy;
        len -= copy;
    }
    if (len) {
        // 如果还有剩余的数据，那创建一个新的节点，然后将数据填充进去
        size_t size = len < PROTO_REPLY_CHUNK_BYTES? PROTO_REPLY_CHUNK_BYTES: len;
        tail = zmalloc(size + sizeof(clientReplyBlock)); // 分配新的内存
        // 计算相应的数据
        tail->size = zmalloc_usable(tail) - sizeof(clientReplyBlock);
        tail->used = len;
        memcpy(tail->buf, s, len); // 拷贝
        listAddNodeTail(c->reply, tail); // 把新节点接到尾部
        c->reply_bytes += tail->size;
    }
    asyncCloseClientOnOutputBufferLimitReached(c); // todo
}
```

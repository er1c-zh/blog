---
title: "redis指令的实现-基础接口"
date: 2020-11-11T20:21:00+08:00
draft: false
tags:
  - redis
  - what
  - how-redis-work
order: 0
---


这里主要用来放置数据库接口等基础接口的简单分析。

<!--more--> 

{{% serial_index how-redis-work %}}

# 基础的参数格式

对于每个`client`中有两个变量来存储参数的个数和具体的值：

```c
  // ...
  // server.h#800
    int argc;               /* 当前命令的参数的个数 */
    robj **argv;            /* 参数的数组，0是指令名（如get），后续依次递增 */
  // ...
```

# 获取key对应的对象

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

`lookupKeyReadWithFlags`首先调用了`expireIfNeeded`来判断key是否过期并进行相应处理。

然后调用`lookupKey`来查找，最后返回结果。

然后看下`lookupKeyWriteWithFlags`，其实现非常简单，调用了`expireIfNeeded`检查过期情况，然后调用`lookupKey`来查找，返回。

还有一些更高级的接口，来提供一些封装，如：

- `lookupKeyRead`对`lookupKeyReadWithFlags`进行简单封装，提供了常用的查询
- `lookupKeyReadOrReply(client, key, reply)`如果查询不到就向客户端返回结果`reply`
- `lookupKeyWrite`和`lookupKeyWriteOrReply`类似read版本

# 返回结果到客户端

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

然后在简单看下每次向客户端写入数据都会调用的`prepareClientToWrite`。
函数检查客户端的若干参数，来判断是否允许写入，
以及之前如果没有注册的话，将“写入数据到socket”这个handler注册到事件循环中(`clientInstallWriteHandler`)。

```c
// src/networking.c#233
int prepareClientToWrite(client *c) {
    if (c->flags & (CLIENT_LUA|CLIENT_MODULE)) return C_OK; // 对于LUA脚本，不需要注册事件
    if (c->flags & CLIENT_CLOSE_ASAP) return C_ERR; // 如果尽快关闭标志被设置，直接返回错误
    // 如果 不要发送结果到客户端 或 跳过该指令的回复 被设置，直接返回错误
    if (c->flags & (CLIENT_REPLY_OFF|CLIENT_REPLY_SKIP)) return C_ERR;
    // 如果对于同步用的主节点客户端，不需要写入
    if ((c->flags & CLIENT_MASTER) &&
        // 除非设置了强制写入
        !(c->flags & CLIENT_MASTER_FORCE_REPLY)) return C_ERR;

    // 对于AOF载入的客户端，不需要写入
    if (!c->conn) return C_ERR; /* Fake client for AOF loading. */

    // 如果之前没有注册过写入handler，那么注册
    if (!clientHasPendingReplies(c)) clientInstallWriteHandler(c);

    // 一切ok 等待写入
    return C_OK;
}
```

具体命令的实现是使用下述的几个更高级别的函数。

- `addReply(client *c, robj *obj)` 将字符串对象`obj`(字符串和数字)尝试写入`buffer`，如果失败，则尝试写入`reply`链表。
- `addReplySds(client *c, sds s)` 将`s`写入回复，并释放。
- etc

# 操作数据库的键值对

## 设置key

首先看下高级别的接口。

`src/db.c#244`的`genericSetKey`方法提供了在db中设置键值对的功能。

1. 首先通过`lookupKeyWrite`判断对应的键是否存在，来决定采用`dbAdd`还是`dbOverwrite`来操作。
1. 调用`incrRefCount`来增加引用数量。
1. 除非特别设置，否则使用`removeExpire`移除过期时间。
1. 通过`signalModifedKey`来通知`WATCH`的等待者。


# 过期时间的实现

过期相关的核心接口都在`src/db.c#1173`处开始。

redis通过将有过期时间的键作为键、将过期时间的绝对时间戳作为值写到另一个字典`db.expires`中。

对过期时间的增删改查由`setExpire`/`removeExpire`/`setExpire`/`getExpire`实现，
其中的逻辑比较清晰和简单，就此略过。

系统还提供了检查键是否过期的`keyIsExpired`和用来检查并处理已经过期的键的`expireIfNeeded`。

`expireIfNeeded`通常在键被访问的地方调用，函数中首先判断是否过期。
如果过期了，对于从实例，会直接返回结果；
对于主实例，会：
  1. 更新统计数据
  1. 调用`propagateExpire`将变化传递给AOF和从实例
  1. 调用`notifyKeyspaceEvent`将过期事件进行广播
  1. 根据配置`lazyfree_lazy_expire`，会尝试异步(`dbAsyncDelete`)或同步(`dbSyncDelete`)的将这个key移除
      - 同步删除通过调用哈希表的`dictDelete`接口分别删除过期map(`db.expires`)和主map中的key，对于集群模式，还会调用`slotToKeyDel`来移除key和slot的关系
      - 异步删除整个流程类似与同步的删除，只是在判断大小如果足够大的时候，会异步**释放**，而不是同步**释放**。
  1. 调用`signalModifiedKey`触发相关的钩子

对于过期的键值对的处理同时需要应用在AOF和从实例上，这个功能由`propagateExpire`来实现。
两者的操作相似，通过生成一个`DEL`或`UNLINK`命令来实现标记对应的键为过期的目的。



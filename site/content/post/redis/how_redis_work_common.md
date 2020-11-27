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

下述的三个更高级别的函数利用上述的两个基础函数完成写入：

- `addReply(client *c, robj *obj)` 将字符串对象`obj`(字符串和数字)尝试写入`buffer`，如果失败，则尝试写入`reply`链表。
- `addReplySds(client *c, sds s)` 将`s`写入回复，并释放。
- `addReplyProto(client *c, const char *s, size_t len)` 相较于其他的方法更加的高效。

其他的方法均基于上述的三个方法实现，业务使用这些包装方法来实现业务。
这些包装的方法与协议息息相关，我认为可以看作是协议返回相关的具体实现。

选择一个方法作为例子，简单介绍下：

`addReplyBulk`写入bulk类型的数据

```c
void addReplyBulk(client *c, robj *obj) {
  addReplyBulkLen(c,obj); // 会先写入协议中要求的$字符和长度
  addReply(c,obj); // 写入结果
  addReply(c,shared.crlf); // 写入结尾的CRLF
}
```


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

对于过期的键的剔除有两种途径：被动删除与主动删除。

对于主动删除，redis会每秒十次的尝试清理过期键。
通过后台的定时循环调用`activeExpireCycle`实现。

```c
void activeExpireCycle(int type) {
  // ...
  for (j = 0; j < dbs_per_call && timelimit_exit == 0; j++) {
    // 迭代适当数量的数据库
    // ...
    redisDb *db = server.db+(current_db % server.dbnum); // 选择出要执行的数据库
    // ...

    // 如果每次循环检查，有超过一定比例的键过期，那么就再次查找
    do {
      // 如果没有设置了过期事件的键、hash表过于稀疏，停止检查
      // ...
      // 开始抽样检查
      // ...
      while (sampled < num && checked_buckets < max_buckets) {
        // ...
                  // 检查是否过期并处理
                  ttl = dictGetSignedIntegerVal(e)-now;
                  if (activeExpireCycleTryExpire(db,e,now)) expired++;
        // ...
      }
      // ...
      if ((iteration & 0xf) == 0) { /* check once every 16 iterations. */
        // 如果消耗过长时间就放弃
        // ...
      }
    } while (sampled == 0 || /* 如果每次循环检查，有超过一定比例的键过期，那么就再次查找 */
             (expired*100/sampled) > config_cycle_acceptable_stale);
  }
  // ...
}
```

被动删除通过`expireIfNeeded`通常在键被访问的地方调用来实现。
函数中首先判断是否过期，
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

# 阻塞指令的实现

比如`BLPOP`这种指令，在没有数据的情况下，会阻塞直到有数据可以返回。

具体的实现上，区分了有无数据的情况：

- 对于有数据的情况，会转而执行指令的非阻塞版本，返回结果。
- 队伍无数据的情况，会将客户端放到`db.blocking_keys`要等待的键的链表上。
  随后会等待插入的指令来唤醒阻塞的客户端。

简单看下阻塞的实现：
```c
void blockForKeys(client *c, int btype, robj **keys, int numkeys, mstime_t timeout, robj *target, streamID *ids) {
  dictEntry *de;
  list *l;
  int j;

  c->bpop.timeout = timeout; // 要等待到该时间戳
  c->bpop.target = target; // 需要接收到数据的键

  if (target != NULL) incrRefCount(target);

  for (j = 0; j < numkeys; j++) { // 依次遍历要等待的key
    bkinfo *bki = zmalloc(sizeof(*bki)); // 构建一个阻塞信息节点对象
    if (btype == BLOCKED_STREAM)
      bki->stream_id = ids[j]; // 对于阻塞在流上，记录id

    if (dictAdd(c->bpop.keys,keys[j],bki) != DICT_OK) {
      zfree(bki);
      continue; // 如果目标key已经存在，忽略
    }
    incrRefCount(keys[j]);

    de = dictFind(c->db->blocking_keys,keys[j]); // 查找要等待的key的等待列表
    if (de == NULL) { // 如果不存在，那么新建一个列表
      int retval;

      l = listCreate();
      retval = dictAdd(c->db->blocking_keys,keys[j],l);
      incrRefCount(keys[j]);
      serverAssertWithInfo(c,keys[j],retval == DICT_OK);
    } else {
      l = dictGetVal(de); // 如果存在，直接获取对应的列表
    }
    listAddNodeTail(l,c); // 将等待信息追加到尾部
    bki->listnode = listLast(l); // 在阻塞信息记录放置的节点
  }
  blockClient(c,btype); // 主要是设置了客户端的阻塞标志，
                        // 并加入到超时检查表中server.clients_timeout_table
                        // 其实现是一个基数树 radix tree，能够快速的查找到长整数
}
```

随后阻塞会被`handleClientsBlockedOnKeys`唤醒。
这个函数在两个场景会被调用：

- 每个指令执行结束且有处于就绪状态的键
- 每个主循环的`beforeSleep`方法会执行一次

```c
void handleClientsBlockedOnKeys(void) {
    while(listLength(server.ready_keys) != 0) {
        // ...
        // 存下就绪的key
        l = server.ready_keys;
        server.ready_keys = listCreate();
        // ...

        while(listLength(l) != 0) { // 遍历每个就绪的键
            listNode *ln = listFirst(l);
            readyList *rl = ln->value;

            dictDelete(rl->db->ready_keys,rl->key); // 移除每个db中的就绪key

            robj *o = lookupKeyWrite(rl->db,rl->key); // 获取数据

            if (o != NULL) {
                // 根据类型唤醒对应的阻塞的客户端
                // 方法内部按照FIFO唤醒阻塞的客户端
                if (o->type == OBJ_LIST)
                    serveClientsBlockedOnListKey(o,rl);
                else if (o->type == OBJ_ZSET)
                    serveClientsBlockedOnSortedSetKey(o,rl);
                else if (o->type == OBJ_STREAM)
                    serveClientsBlockedOnStreamKey(o,rl);
            }
            // ...
        }
        // ...
    }
}
```

最后简单介绍下超时的实现。
redis会在每次`beforeSleep`中调用`handleBlockedClientsTimeout`来处理阻塞超时的客户端。
其核心是利用`server.clients_timeout_table`的迭代器，遍历等待客户端，
直到当前时间戳为止。

# scan相关的实现

`SCAN`,`HSCAN`,`SSCAN`和`ZSCAN`四个scan命令最终都由`scanGenericCommand`实现。

实现中，对于数据量可控的情况，采取了直接返回全部数据的方式，降低了复杂度，提高了性能。
数据量大的情况均为hash表承载，由hash表的接口`dictScan`实现功能。

```c
// o 是要扫描的哈希表/集合/有序集合
void scanGenericCommand(client *c, robj *o, unsigned long cursor) {
    // ...

    // 首先解析了参数，分别存到count/pat,use_pattern/type中
    // ...

    // 检查是否是hash表作为存储底层，如果是将ht指向对应的hash表。
    ht = NULL;
    if (o == NULL) {
        // 如果传入的o为空，那么选择整个数据库
        ht = c->db->dict;
    } else if (o->type == OBJ_SET && o->encoding == OBJ_ENCODING_HT) {
        // 由hash表表示的set
        ht = o->ptr;
    } else if (o->type == OBJ_HASH && o->encoding == OBJ_ENCODING_HT) {
        // 由hash表表示的哈希表
        ht = o->ptr;
        count *= 2; /* We return key / value for this type. */
    } else if (o->type == OBJ_ZSET && o->encoding == OBJ_ENCODING_SKIPLIST) {
        // 由hash表表示的zset
        zset *zs = o->ptr;
        ht = zs->dict;
        count *= 2; /* We return key / value for this type. */
    }

    if (ht) {
        // ...
        long maxiterations = count*10; // 设置迭代次数的上限，避免过长时间阻塞
        do {
            // 依赖hash表的接口来扫描
            cursor = dictScan(ht, cursor, scanCallback, NULL, privdata);
        } while (cursor &&
              maxiterations-- &&
              listLength(keys) < (unsigned long)count);
    } else if (o->type == OBJ_SET) {
        int pos = 0;
        int64_t ll;

        // 非hash表的实现数据量比较小
        // 遍历全部，全部取出是可控的
        while(intsetGet(o->ptr,pos++,&ll)) 
            listAddNodeTail(keys,createStringObjectFromLongLong(ll));
        cursor = 0;
    } else if (o->type == OBJ_HASH || o->type == OBJ_ZSET) {
        unsigned char *p = ziplistIndex(o->ptr,0);
        unsigned char *vstr;
        unsigned int vlen;
        long long vll;

        // 非hash表的实现数据量比较小
        // 遍历全部，全部取出是可控的
        while(p) {
            ziplistGet(p,&vstr,&vlen,&vll);
            listAddNodeTail(keys,
                (vstr != NULL) ? createStringObject((char*)vstr,vlen) :
                                 createStringObjectFromLongLong(vll));
            p = ziplistNext(o->ptr,p);
        }
        cursor = 0;
    } else {
        serverPanic("Not handled encoding in SCAN.");
    }

    node = listFirst(keys);
    while (node) {
        robj *kobj = listNodeValue(node);
        nextnode = listNextNode(node);
        int filter = 0;

        // 过滤pattern规定的格式 
        // 过滤type
        // 过滤过期键
        // 具体实现不是重点
        // ...

        if (filter) {
            decrRefCount(kobj);
            // 如果不符合预期，那么从答案列表中移除
            listDelNode(keys, node);
        }

        if (o && (o->type == OBJ_ZSET || o->type == OBJ_HASH)) {
            // 对于要返回键值对的情况，那么值也要清除
            node = nextnode;
            nextnode = listNextNode(node);
            if (filter) {
                kobj = listNodeValue(node);
                decrRefCount(kobj);
                listDelNode(keys, node);
            }
        }
        node = nextnode;
    }

    // 返回结果
    // ...
    addReplyBulkLongLong(c,cursor); // 新的游标
    // ...
    while ((node = listFirst(keys)) != NULL) {
        // 返回结果
        // ...
    }

cleanup:
    // 清理资源
    // ...
}
```

# 参考

- [how-redis-expires-keys](https://redis.io/commands/expire#how-redis-expires-keys)
- [redis protocol](https://redis.io/topics/protocol)

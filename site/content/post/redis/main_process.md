---
title: "redis处理请求的流程"
date: 2020-11-03T03:59:02+08:00
draft: false
tags:
    - redis
    - what
---


*基于redis 4.0版本*

出于学习的目的，简单记录下redis处理连接与请求的流程。

最后，简单介绍了6.0版本引入的多线程IO的实现方式。

# 接受链接

redis使用一个简单的事件驱动框架（以下简称框架，之前的文章有对于该框架的简单分析）来监听服务端链接的事件。

首先在初始化服务器的时候，监听对应的地址。

然后向框架注册对应的事件和处理函数`acceptTcpHandler`。在有新的链接到达时，框架会调用该方法。

```c
// server.c
void initServer(void) {
  // ...
  /* Open the TCP listening socket for the user commands. */
  if (server.port != 0 &&
    listenToPort(server.port,server.ipfd,&server.ipfd_count) == C_ERR)
    exit(1);

  // ...

  /* Create an event handler for accepting new connections in TCP and Unix
   * domain sockets. */
  for (j = 0; j < server.ipfd_count; j++) {
      if (aeCreateFileEvent(server.el, server.ipfd[j], AE_READABLE,
          acceptTcpHandler,NULL) == AE_ERR)
          {
              serverPanic(
                  "Unrecoverable error creating server.ipfd file event.");
          }
  }
  // ...
}
```

`acceptTcpHandler`中调用若干方法，最终使用`accept`接受链接。然后将控制权交给`acceptCommonHandler`方法。

```c
// src/networking.c
void acceptTcpHandler(aeEventLoop *el, int fd, void *privdata, int mask) {
    int cport, cfd, max = MAX_ACCEPTS_PER_CALL;
    char cip[NET_IP_STR_LEN];
    UNUSED(el);
    UNUSED(mask);
    UNUSED(privdata);

    while(max--) {
        cfd = anetTcpAccept(server.neterr, fd, cip, sizeof(cip), &cport);
        if (cfd == ANET_ERR) {
            if (errno != EWOULDBLOCK)
                serverLog(LL_WARNING,
                    "Accepting client connection: %s", server.neterr);
            return;
        }
        serverLog(LL_VERBOSE,"Accepted %s:%d", cip, cport);
        acceptCommonHandler(cfd,0,cip);
    }
}
```

`acceptCommonHandler`中调用`createClient`处理业务逻辑，并记录若干统计数据和进行安全检查。

```c
// src/networking.c
static void acceptCommonHandler(int fd, int flags, char *ip) {
    client *c;
    if ((c = createClient(fd)) == NULL) {
      // ...
    }
    // ...
}
```

`createClient`设置链接的属性，将处理函数与fd注册到框架中，最后构建`client`对象。

```c
// src/networking.c
client *createClient(int fd) {
    client *c = zmalloc(sizeof(client)); // 分配内存

    // ...
    anetNonBlock(NULL,fd); // 设置socket为非阻塞模式
    anetEnableTcpNoDelay(NULL,fd); // 设置TCP_NODELAY，减少小包的延迟
    if (server.tcpkeepalive)
        anetKeepAlive(NULL,fd,server.tcpkeepalive); // 设置长链接
    if (aeCreateFileEvent(server.el,fd,AE_READABLE,
        readQueryFromClient, c) == AE_ERR) // 注册事件
    {
        // ...
    }
    // ... 进行了一些初始化客户端的操作
}
```

# 读取请求

`readQueryFromClient`根据`client->reqtype`，使用不同方式处理单行命令(PROTO_REQ_INLINE)和一组命令(PROTO_REQ_MULTIBULK)。
随后将程序交给`processInputBuffer`，执行后续的操作。

```c
// src/networking.c
void readQueryFromClient(aeEventLoop *el, int fd, void *privdata, int mask) {
  // ...
  readlen = PROTO_IOBUF_LEN; // 需要读取的数量
  // ...
  c->querybuf = sdsMakeRoomFor(c->querybuf, readlen); // 修改querybuf的长度，流出空间。
  nread = read(fd, c->querybuf+qblen, readlen); // 读取数据到查询指令缓冲区
  // ...
  if (!(c->flags & CLIENT_MASTER)) {
        // 如果是普通的客户端（区分于主从复制的的master服务器），只需要处理即可
        processInputBuffer(c);
    } else {
        // 请求是主从复制的主发送过来的
        size_t prev_offset = c->reploff;
        processInputBuffer(c);
        size_t applied = c->reploff - prev_offset;
        if (applied) {
            // 如果应用了主推送过来的变化，那么把变化发送给自己的从。
            replicationFeedSlavesFromMasterStream(server.slaves,
                    c->pending_querybuf, applied);
            sdsrange(c->pending_querybuf,applied,-1);
        }
    }
}
```

# 分析请求

`processInputBuffer`根据不同的类型请求，调用`processInlineBuffer`或`processMultibulkBuffer`解析参数。随后调用`processCommand`执行请求。

# 进行各个指令的操作

`processCommand`通过`call`调用每个指令在指令表中的处理函数来进行指令的具体操作。

```c
// src/server.c#127
// 指令表
struct redisCommand redisCommandTable[] = {
    {"module",moduleCommand,-2,"as",0,NULL,0,0,0,0,0},
    {"get",getCommand,2,"rF",0,NULL,1,1,1,0,0},
    {"set",setCommand,-3,"wm",0,NULL,1,1,1,0,0},
    // ...
}
```

## `redisComand`

每个指令都有一个`redisCommand`实例，来表明如何执行该指令。

```c
// src/server.h#1500
struct redisCommand {
    char *name; // 指令的名称
    redisCommandProc *proc; // 具体的执行函数
    int arity; // 参数个数
    char *sflags;   /* 字符串形式的标记 */
    uint64_t flags; /* 由sflags计算来的数字形式的标记 */
    /* Use a function to determine keys arguments in a command line.
     * Used for Redis Cluster redirect. */
    redisGetKeysProc *getkeys_proc; // 用于从指令中提取出所有的键，用于集群的转发
                                    // 只有当下面三个参数无法确定键时才会被使用
    /* What keys should be loaded in background when calling this command? */
    // 那些键值对需要在后台加载
    int firstkey; /* 第一个作为key的参数的位置 */
    int lastkey;  /* 最后一个作为key的参数的位置 */
    int keystep;  /* 每个key之间的间隔 */
    long long microseconds, calls; // microseconds 执行这个指令所耗费的总时间
                                   // calls 统计数据，对这个指令调用的次数
    int id;     /* 指令的id 用于ACL检查或其他因素 */
};
```

# 返回结果

对于不同的执行路径，最终都使用类似`addReply`的方法添加结果到输出缓冲区。

每次事件循环的中间，都会调用的`beforeSleep`方法中的`handleClientsWithPendingWrites`方法会处理有等待输出的client。

1. 首先调用`writeToClient`方法，尝试同步写入。
1. 如果写入成功之后，检查发现还是有数据需要写入（比如是达到了写入限制`NET_MAX_WRITES_PER_EVENT`），那么就向事务框架中注册`AE_WRITABLE`事件和处理函数`sendReplyToClient`，最终会再次调用`writeToClient`尝试写入结果。

# redis 6.0引入的并发IO的简单介绍

在读请求的时候，`readQueryFromClient`引入一个方法`postponeClientRead`检查是否进行异步读。

在异步读开启时，会将该延迟读加到`redisSever.clients_pending_read`链表中。

每次事件循环间隙，会调用`handleClientsWithPendingReadsUsingThreads`方法 *（src/server.c#2136）*，将需要进行读取的请求，分发给每个IO线程。

```c
int handleClientsWithPendingReadsUsingThreads(void) {
    // ...
    while((ln = listNext(&li))) {
        client *c = listNodeValue(ln);
        int target_id = item_id % server.io_threads_num; // 保证均匀的分发，这里是单线程的处理，无需加锁。
        listAddNodeTail(io_threads_list[target_id],c);
        item_id++;
    }
    // ...
}
```

写请求也是类似的处理方式，具体的处理方法是`handleClientsWithPendingWritesUsingThreads`。

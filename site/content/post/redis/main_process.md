---
title: "redis处理请求的流程"
date: 2020-11-03T03:59:02+08:00
draft: true
tags:
    - redis
    - what
---


*基于redis 4.0版本*

出于学习的目的，简单记录下redis处理连接与请求的流程。

# 接受连接

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

# 分析请求

# 进行各个指令的操作

# 返回结果

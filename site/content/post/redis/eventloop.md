---
title: "redis的事件循环简单分析"
date: 2020-11-01T19:30:35+08:00
draft: false
tags:
    - redis
    - what
---

*基于redis 4.0版本*

redis使用了一个称为`ae`的事件框架来处理所有的事件，包含时间事件和文件事件。

出于学习的目的，记录下阅读源码的过程。

# 数据结构的定义

*src/ae.h*

## ```EventLoop```

`EventLoop`存储了整个事件循环相关的状态。

```c
typedef struct aeEventLoop {
    int maxfd;   /* 当前注册中的最大文件描述符 */
    int setsize; /* 最大的可被追踪的文件描述符数值 */
    long long timeEventNextId;
    time_t lastTime;     /* Used to detect system clock skew 探测系统时钟的偏移 */
    aeFileEvent *events; /* 文件事件的数组 根据setsize初始化 */
    aeFiredEvent *fired; /* 被激活的事件 每次主循环被poll更新 */
    aeTimeEvent *timeEventHead; /* 时间事件的链表头 */
    int stop;
    void *apidata; /* 用于存储与底层多路复用库相关的数据 */
    aeBeforeSleepProc *beforesleep;
    aeBeforeSleepProc *aftersleep;
} aeEventLoop
```

## 文件事件

```c
/* File event structure */
typedef struct aeFileEvent {
    int mask; /* one of AE_(READABLE|WRITABLE|BARRIER) 事件的状态 可读/可写/？ */
    aeFileProc *rfileProc; // 两个事件处理函数
    aeFileProc *wfileProc;
    void *clientData;
} aeFileEvent;
```

## 时间事件

```c
/* Time event structure */
typedef struct aeTimeEvent {
    long long id; /* time event identifier. */
    long when_sec; /* seconds */
    long when_ms; /* milliseconds */
    aeTimeProc *timeProc;
    aeEventFinalizerProc *finalizerProc;
    void *clientData;
    struct aeTimeEvent *prev;
    struct aeTimeEvent *next;
} aeTimeEvent;
```

# 生命周期

这一部分我们参考redis的使用来进行分析。

1. 初始化一个事件循环

    在初始化过程中调用`aeEventLoop *aeCreateEventLoop(int setsize)`初始化一个`EventLoop`对象并设置追踪的文件描述符的最大数量。

    `aeCreateEventLoop`调用了特定的多路复用框架的初始化方法。epoll初始化了一个与`EventLoop`中events大小相同的`epoll_event`数组。

    ```c
    void initServer(void) {
      // ...
      // src/server.c#1846
      server.el = aeCreateEventLoop(server.maxclients+CONFIG_FDSET_INCR);
      // ...
    }
    ```

1. 注册事件

    ```c
    void initServer(void) {
      // ...
      // src/server.c#1921
      // 注册时间事件处理函数
      if (aeCreateTimeEvent(server.el/* eventloop */, 
          1 /* 1毫秒 */, serverCron /* 处理函数 */ , NULL, NULL) == AE_ERR) {
          serverPanic("Can't create event loop timers.");
          exit(1);
      }

      // 注册若干被动的TCP链接文件描述符
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

    1. `aeCreateTimeEvent`将接受到的处理函数包装成时间事件对象，并衔接到传入的`EventLoop`对象的时间事件对象链表上。

    1. `aeCreateFileEvent`将接收到的文件事件处理函数包装成文件事件对象，加入到`EventLoop`对象的`events`数组中。

        1. 根据所在环境的不同，分别使用不同的多路复用的库，包括evport/epoll/kqueue/select。通过定义的预处理常量引入不同的文件来实现。这里用`src/ae_epoll.c`作为分析的例子。

        1. 使用`aeApiAddEvent`函数(src/ae_epoll.c#73)来实现新增文件事件的监听。处理了新增与修改监听事件种类的各种case。

            ```c
            static int aeApiAddEvent(aeEventLoop *eventLoop, int fd, int mask) {
                aeApiState *state = eventLoop->apidata;
                struct epoll_event ee = {0}; /* avoid valgrind warning */
                /* If the fd was already monitored for some event, we need a MOD
                 * operation. Otherwise we need an ADD operation. */
                int op = eventLoop->events[fd].mask == AE_NONE ?
                        EPOLL_CTL_ADD : EPOLL_CTL_MOD; // 如果没有就增加，否则修改

                ee.events = 0;
                mask |= eventLoop->events[fd].mask; /* Merge old events */
                if (mask & AE_READABLE) ee.events |= EPOLLIN;
                if (mask & AE_WRITABLE) ee.events |= EPOLLOUT;
                ee.data.fd = fd;
                if (epoll_ctl(state->epfd,op,fd,&ee) == -1) return -1;
                return 0;
            }
            ```
1. 注册两个时间点的函数

    两个函数在每次事件主循环中，都会在前后调用。

    ```c
    void initServer(void) {
        // ...
        // src/server.c#3902
        aeSetBeforeSleepProc(server.el,beforeSleep);
        aeSetAfterSleepProc(server.el,afterSleep);
        // ...
    }
    ```

1. 开始事件循环

    ```c
    void initServer(void) {
        // ...
        // src/server.c#3904
        aeMain(server.el);
        // ...
    }
    ```

    1. 一个简单的循环，调用`beforeSleep`与`aeProcessEvents`。

        ```c
        void aeMain(aeEventLoop *eventLoop) {
            eventLoop->stop = 0;
            while (!eventLoop->stop) {
                if (eventLoop->beforesleep != NULL)
                    eventLoop->beforesleep(eventLoop);
                aeProcessEvents(eventLoop, AE_ALL_EVENTS|AE_CALL_AFTER_SLEEP);
            }
        }
        ```

    1. `aeProcessEvents`

        对于事件采取先处理读，然后处理写的策略，
        这样可以在读取一个命令并执行完成后，迅速的写入到链接中。
        但例外的是，可以通过传入flag `AE_BARRIER` 来反转上述的策略，
        用于在读取执行命令和写入结果之间需要通过`beforeSleep`或`afterSleep`对结果做一些处理的情况。

        ```plantuml
        @startuml
        title aeProcessEvents

        start
        :计算最近发生的时间事件距今需要多长时间tvp;
        :调用系统相关的aeApiPoll接口，等待tvp到达或者其他文件事件被触发;
        :调用aftersleep处理函数;

        repeat
        :调用读事件handler;
        :调用写事件handler;
        repeat while(poll返回的事件仍有未处理完成的)

        :调用processTimeEvents处理时间事件;
        stop
        @enduml
        ```

1. 事后清理

```c
void initServer(void) {
    // ...
    // src/server.c#3905
    aeDeleteEventLoop(server.el);
    return 0;
}
```

# `beforeSleep`、`afterSleep`与`serverCron`

简单记录了各个“定时任务”做的工作，
不包含集群、模组等的相关的内容。

## `beforeSleep`

1. 允许时，调用`activeExpireCycle`来尝试一轮主动的快速过期键检查。
具体实现在[这篇文章]({{< relref "how_redis_work_common.md#过期时间的实现" >}})有介绍。
1. `flushAppendOnlyFile`刷新AOF。
1. `handleClientsWithPendingWrites`将需要发送到客户端的缓冲区发送并清空，
具体实现在[这篇文章]({{< relref "main_process.md#返回结果" >}})有介绍。

## `afterSleep`

模组相关的工作。

## `serverCron`

1. 首先是做了一些统计、监控方面的工作：

    - 占用的资源
    - 不同db的大小、包含的key的数量
    - 客户端的信息

1. `clientsCron`
1. `databasesCron`
1. 如果有触发后台重写aof，会调用`rewriteAppendOnlyFileBackground`
1. 检查子进程是否有中途失败的情况。
1. 检查是否需要进行rdb或aof的持久化操作。
1. 如果需要，刷新aof缓存。
1. 清理客户端。
1. 如果需要，开始一个后台的rdb保存工作。

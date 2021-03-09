---
title: "redis bio是什么与如何实现的"
date: 2021-03-09T08:59:52+08:00
draft: false
tags:
    - redis
    - how
    - what
---

在看aof刷盘的代码时，
发现异步刷盘是通过一个叫`bio`的组件完成的。
正好研究一下。

<!--more-->

# 是什么

`bio` *（大概）* 是`background I/O`的缩写。

根据注释描述，是用来异步处理一些IO相关的工作。

总的来说，`bio`通过任务队列和若干消费者进程来实现异步任务框架。

# 实现的细节

任务框架自然需要的一个结构体`bio_job`来表示：

```c
struct bio_job {
    time_t time; /* Time at which the job was created. */
    /* Job specific arguments pointers. If we need to pass more than three
     * arguments we can just pass a pointer to a structure or alike. */
    void *arg1, *arg2, *arg3;
}
```

比较有意思的是其中没有任务类型，
任务的区分是由不同的任务队列来完成的，
有`bioInit`函数来完成包括若干任务队列在内的整个框架的初始化。

初始化函数主要作了以下几个事情：

1. 根据`BIO_NUM_OPS`创建对应的任务列表和等待计数。
    - 列表是一个普通的双向链表
1. 设置进程的栈大小。
1. 根据`BIO_NUM_OPS`创建对应的消费者线程。
    - `bioProcessBackgroundJobs`

接下来自然就是创建任务了——`bioCreateBackgroundJob`：

函数的实现非常清晰，
创建一个任务对象，加锁后用`listAddNodeTail`追加任务节点，
增加等待计数，调用`pthread_cond_signal`发出信号，最后解锁。

最后，也是最重要的，处理任务由`bioProcessBackgroundJobs`来完成。

## `bioProcessBackgroundJobs`

对于每个任务类型有且只有一个处理线程处理，这简化了并发的工作。

忽略pthread相关的工作，那么整个实现就很简单了：

1. 利用`listFirst`获取一个任务。
1. 根据类型作出需要的处理。
1. 释放任务节点。


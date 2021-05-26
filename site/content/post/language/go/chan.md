---
title: "go channel原理分析"
date: 2021-05-11T16:30:05+08:00
draft: false
tags:
    - go-src
    - go
    - what
order: 2
---

从功能出发，分析channel的原理。

<!--more-->

{{% serial_index go-src %}}

*go1.16*

*src/runtime/chan.go*

# TLDR

## 原理

channel内部有写入､读取两个等待队列，
当数据溢出或饥饿时，
会阻塞并追加到写入､读取队列上。

如果配置了buf，那么有一个环形数组来实现缓冲区。

## 使用上的点

### 读取写入是有序的

### 关闭nil的channel或重复关闭channel会导致panic

### 只由写入方关闭channel

向一个已关闭的channel写入数据，
或（因channel缓冲已满）等待写入时channel被关闭都会产生panic。

所以关闭channel应该保证没有写入方会写入时再关闭，
简单地，可以用“只由写入方”关闭来理解。

# 数据结构 

从功能来看，
channel可以简单看作一个传递数据的管道，
可以暂存预定义上限数量的数据，
并发安全，允许并行的读写。

如果暂无数据或缓存溢出，
还会将读取或写入操作阻塞，
直到可读可写。

channel支持关闭，
对一个关闭了的channel进行操作有固定的结果。

可以猜想一个channel对象中：
- 有一个区域来缓存数据
- 有goroutine的等待队列来维持阻塞
- 有锁或其他机制来处理并发问题
- 有状态变量来维护是否关闭

首先来看下channel的结构：

```go
type hchan struct {
    qcount   uint           // 缓冲区中目前的数量
    dataqsiz uint           // chan的缓冲区的元素数量上限
    buf      unsafe.Pointer // 一个环形缓冲区
    elemsize uint16
    closed   uint32
    elemtype *_type // element type
    sendx    uint   // send index 下一个要插入的下标
    recvx    uint   // receive index
    recvq    waitq  // list of recv waiters
    sendq    waitq  // list of send waiters

    // lock protects all fields in hchan, as well as several
    // fields in sudogs blocked on this channel.
    //
    // Do not change another G's status while holding this lock
    // (in particular, do not ready a G), as this can deadlock
    // with stack shrinking.
    lock mutex
}
```

字段不多，用途也比较明确：
- `sendx` `recvx` `elemtype` `qcount` `dataqsiz` `buf` `elemsize` 支持暂存功能
- `closed` 存储是否关闭
-  有两个等待队列 `recvq` `sendq`
- 有一个锁 `lock` 用来保护chan的全部字段，根据注释，还保护了阻塞在该chan上的`g`中的某些字段

# 功能

channel相关的功能有：

- 新建
- 从chan读取
- 写入chan
- 通过`select`写入或读取
- 关闭

本文接下来会依次分析相关功能的实现原理。

## 新建

代码中的`ch := make(chan interface{})`最终由函数
`func makechan(t *chantype, size int) *hchan`
执行。

`makechan`需要做：
1. 初始化`hchan`和内部的暂存内存
1. 初始化锁与`elemsize`等信息

`makechan`根据chan传输的数据类型是否包含指针来选择具体的分配内存的方式。

如果是不包含指针的情况，
会直接调用`mallocgc`在noscan的span上分配`hchan`和暂存区的内存。
反之，会先调用`new`来初始化`hchan`，
随后调用`mallocgc`正常初始化暂存区。

内存分配完成之后，
配置好传输的数据类型的大小､类型，通过`lockInit`初始化好锁。

新建一个chan到此就完成了。

## 写入与读取

读写chan有两种情况：
1. 直接读写
1. 通过select来监听多个“IO事件”

这里先分析第一种情况，
对于通过`select`来操作chan的情况后面会具体分析。

从写入的视角来看，
主要有三种情况：

1. chan上阻塞了等待数据的读取goroutine
1. chan上没有等待数据的goroutine，那么就尝试写入缓冲区
1. 缓冲区满，（如果需要）那么就将该goroutine阻塞

反之，读取也是类似的。

### 写入 `chansend`

写入操作的核心流程由`func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool`
实现，接受四个参数：

1. c 要写入的chan
1. ep
1. block 在无法写入的时候是否阻塞
1. callerpc

返回是否写入成功。

1. 不加锁的检查

    首先检查目标chan是否为nil，
    如果是nil，
    会根据`block`直接返回`false`或调用`gopark`挂起，
    这种情况的挂起会导致该goroutine永久的挂起。
    
    然后在**不加锁的情况下**检查chan是否**立即可写**，
    如果参数`block`为false且不是立即可写的，
    直接返回false。
    
    **立即可写**需要满足channel未关闭且channel的缓冲区未满。

1. 获取锁

    获取chan的锁，然后检查chan是否已经关闭，
    如果已经关闭抛出panic。

1. 写入-有等待的读goroutine

    通过`c.recvq.dequeue()`尝试获取一个等待读取的goroutine，
    如果成功，调用`send`发送数据，完成写入，返回`true`。
    `send`的具体实现暂且放过，等后续流程分析完成后，
    在结合发送部分一同分析。

1. 写入-缓冲区

    检查`c.qcount < c.dataqsiz`来判断缓冲区是否已满，写入。

    `c.qcount`是目前缓冲区中有的未读取的元素数量；
    `c.dataqsiz`是上限。

    具体的分析在下面的注释中了。

    ```go
    if c.qcount < c.dataqsiz { // 如果没有满
        // Space is available in the channel buffer. Enqueue the element to send.
        qp := chanbuf(c, c.sendx) // 计算出要插入的地址
        if raceenabled {
            racenotify(c, c.sendx, nil)
        }
        // 拷贝要插入的数据到刚刚计算出的地址
        typedmemmove(c.elemtype, qp, ep)
        // 更新环形缓冲区的指针
        c.sendx++
        if c.sendx == c.dataqsiz {
            c.sendx = 0
        }
        // 更新chan缓冲的元素数量
        c.qcount++
        // 解锁
        unlock(&c.lock)
        // 完成写入
        return true
    }
    ```

1. 写入-等待写入队列

    到这个时候，已经无法立即写入了，
    如果`block`是`false`，那么便立即返回；
    反之，将该goroutine加入到chan的等待队列上。

    首先获取一个当前goroutine的`sudog`对象，
    填充相关信息，追加到`c.sendq`上。

    `gopark(chanparkcommit, ...)`挂起。

    这个时候该goroutine就开始等待直到读取的goroutine唤醒。

    特别的，利用`KeepAlive`来保证在消费者“拥有”要传输的对象之前不会被清理。

    ```go
    gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanSend, traceEvGoBlockSend, 2)
    // Ensure the value being sent is kept alive until the
    // receiver copies it out. The sudog has a pointer to the
    // stack object, but sudogs aren't considered as roots of the
    // stack tracer.
    KeepAlive(ep)
    ```

1. 写入-阻塞后被唤醒

    清理goroutine等待的标记。
    
    释放`sudog`。

    检查是否传输成功，根据channel的状态产生异常：

    - 如果channel被关闭，那么抛出panic
    - 反之，产生非预期唤醒的fatal：`chansend: spurious wakeup`

1. 完成

    返回`true`。

### 读取 `chanrecv`

读取由`func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool)`完成。

接受三个参数：

- c 要读取的chan
- ep 读取到的元素存储的地址
- block 如果不能立即读取，是否阻塞

返回两个值：

- selected 如果读取有结果（channel关闭了或读取到数据），返回`true`
- received 如果读取到了数据，返回`true`

类似于写入，读取也有三种情况：

- 有等待的写入goroutine
- 缓冲区有待读取的数据
- 没有数据可以读取，（如果需要）阻塞在chan上等待

1. 加锁前的检查

    如果chan为nil，非阻塞情况返回；允许阻塞的goroutine则会永久的挂起在该goroutine上。

    如果不能立即读取到数据且不允许阻塞的情况下，直接返回结果。

1. 加锁及检查channel是否关闭

    获取chan的锁，
    随后检查chan状态，
    如果chan已经关闭**且没有未读取的数据**，
    返回“**无数据可读取､chan已经关闭的情况**” _(`selected = true, received = false`)_。

1. 读取-写入等待队列

    尝试从`c.sendq`获取一个等待写入的goroutine，
    如果存在，那么调用`recv`获取数据，返回“**读取到数据**”。

    `func recv(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int)`

    函数接受目标chan`c`，pop出来的等待写入的sudog`sg`，
    目标地址，解锁函数，`goready`需要的参数`skip`。

    `recv`针对有无缓冲区进行不同的操作。

    如果没有缓冲区，调用`recvDirect`直接从弹出的sudog的中读取数据；
    反之，从缓冲区中读取第一个数据，将`sg`的数据写入到缓冲区。

    解锁，唤醒`sg`对应的goroutine，返回。

    前述提到的`send`实现与`recv`类似，
    但是只有直接写入到等待读取的goroutine的部分，
    由`sendDirect`完成。最后解锁，唤醒等待读取的goroutine，返回。

1. 读取-缓冲区

    如果缓冲区中有数据，就从其中读取一个。

    ```go
    if c.qcount > 0 {
        // Receive directly from queue
        qp := chanbuf(c, c.recvx) // 获取一个缓冲区中的数据
        if raceenabled {
            racenotify(c, c.recvx, nil)
        }
        if ep != nil {
            typedmemmove(c.elemtype, ep, qp) // 如果不为空，复制数据
        }
        typedmemclr(c.elemtype, qp) // 声明变量的类型
        // 更新缓冲区标记数据
        c.recvx++
        if c.recvx == c.dataqsiz {
            c.recvx = 0
        }
        c.qcount--
        unlock(&c.lock)
        // 返回成功
        return true, true
    }
    ```

1. 读取-阻塞

    类似写入，获取､构建一个`sudog`，
    追加到`c.recvq`中。
    调用`gopark`阻塞。

    等待直到一个写入的goroutine传输好数据，
    唤醒。

    清理､释放`sudog`，
    检查是否传输成功(`mysg.success // mysg是sudog`)，
    返回结果。

## 关闭channel

关闭channel由`func closechan(c *hchan)`完成。

首先检查目标chan是否为nil或已经关闭，
如果是抛出panic，
其中，检查是否已经关闭之前加锁。

依次遍历`sendq`和`recvq`，
唤醒所有的goroutine，
对于等待读取的goroutine返回“已关闭”，
对于等待写入的程序，直接返回（会触发panic）。


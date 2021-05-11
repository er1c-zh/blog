---
title: "go channel原理分析"
date: 2021-05-11T16:30:05+08:00
draft: true
tags:
    - go-src
    - go
    - what
order: 1
---

从功能出发，分析channel的原理。

<!--more-->

{{% serial_index go-src %}}

*go1.16*

*src/runtime/chan.go*

# TLDR

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
	qcount   uint           // total data in the queue
	dataqsiz uint           // size of the circular queue
	buf      unsafe.Pointer // points to an array of dataqsiz elements
	elemsize uint16
	closed   uint32
	elemtype *_type // element type
	sendx    uint   // send index
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




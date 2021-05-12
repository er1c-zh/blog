---
title: "go内存模型与sync包"
date: 2021-05-01T03:06:06+08:00
draft: false
tags:
    - go
    - what
    - go-src
order: 1
---

go中的Happen-before保证与`sync`包原理的简单分析。

<!--more-->

{{% serial_index go-src %}}

go的内存模型与并发访问也离不开`Happens Before`概念。

简单来说，A先行发生于B表示B可以看到A造成的影响。

在go中有哪些`Happens Before`相关的保证呢？

- 初始化的时候，被依赖的包的`init()`函数先行发生于依赖它的包的`init()`。
- 任何`init()`先行发生于`main.main()`（主函数）。
- 开启新goroutine的`go`原语先行发生于要执行的函数。
- 一个goroutine的退出**不**保证先行发生于任何事件。
- channel
  - 对于一个channel上传递的对象，写入这个对象先行发生于接收到这个对象完成。
  - channel的关闭先行发生于接收到因channel关闭而返回的零值。
  - 对于一个没有buffer的channel，从这个channel上接受先行发生于在这个channel写入的完成。
  - 对于一个容量为C的channel，第k次接受先行发生于k+C次发送完成。
- sync包
  - 对于sync包中的`Mutex`和`RWMutex`，第n次`Unlock()`先行发生于第n+1次`Lock()`。
  - 对于`sync.Once`，任何`once.Do(f)`的`f()`的返回先行发生于其他`once.Do(f)`的返回。

# 一些常见的错误

## Double-Checked locking

这是一种常见的双重检测锁定来实现单例模式。

```go
var a string
var done bool

func setup() {
    a = "hello, world"
    done = true
}

func doprint() {
    if !done {
        once.Do(setup)
    }
    print(a)
}

func main() {
    go doprint() // a
    go doprint() // b
    // 这里可能打印出一个空白字符串
}
```

分析如下：

1. 假定a先被调度执行到`!done`判断并执行进入初始化。
1. 因为不同goroutine中，看到的其他goroutine的执行顺序是无序的。在b中可能先观察到`done`被设置为true，但a还未被设置。

# sync包的简单分析

sync包提供了若干同步用的原语，通常来说，建议使用channel来进行同步。

## `sync.Mutex`

```go
src/sync.mutex.go
type Mutex struct {
    state int32
    sema  uint32
}
```

`Mutex`的定义十分简单，一个标记状态用的`state`，一个信号量。

`Mutex`有两个工作模式：正常模式与饥饿模式。两种模式分别来处理竞争较少的情况，和竞争较多的情况。

### 加锁流程

对于最简单的情况，使用CAS将state设置成锁住的状态。

其他情况进入慢速路径`lockSlow()`。

```go
func (m *Mutex) lockSlow() {
    var waitStartTime int64
    starving := false // 是否是饥饿模式
    awoke := false
    iter := 0
    old := m.state // 旧有的state
    for {
        // Don't spin in starvation mode, ownership is handed off to waiters
        // so we won't be able to acquire the mutex anyway.
        // 锁定 && 饥饿模式 && 如果系统支持自旋，那么尝试自旋。
        if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) /* 判断是否要进行自旋 具体实现在 runtime/proc.go#5521*/ {
            // Active spinning makes sense.
            // Try to set mutexWoken flag to inform Unlock
            // to not wake other blocked goroutines.
            // 标记好正在自旋，使释放锁时，不需要再通知其他阻塞等待的goroutine
            if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
                atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
                awoke = true
            }
            runtime_doSpin()
            iter++
            old = m.state
            continue
        }
        new := old
        if old&mutexStarving == 0 {
        // 饥饿模式直接进入队尾
            new |= mutexLocked
        }
        if old&(mutexLocked|mutexStarving) != 0 {
        // 处于加锁状态，增加等待计数
            new += 1 << mutexWaiterShift // 常量 用于将state高位当作等待计数器
        }
        // The current goroutine switches mutex to starvation mode.
        // But if the mutex is currently unlocked, don't do the switch.
        // Unlock expects that starving mutex has waiters, which will not
        // be true in this case.
        // 在锁住且需要转换时，转换到饥饿模式
        if starving && old&mutexLocked != 0 {
            new |= mutexStarving
        }
        if awoke {
            // The goroutine has been woken from sleep,
            // so we need to reset the flag in either case.
            if new&mutexWoken == 0 {
                throw("sync: inconsistent mutex state")
            }
            new &^= mutexWoken
        }
        if atomic.CompareAndSwapInt32(&m.state, old, new /* new必然被设置为已加锁 */) {
            // 尝试加锁
            if old&(mutexLocked|mutexStarving) == 0 {
                // 加锁成功
                break // locked the mutex with CAS
            }
            // If we were already waiting before, queue at the front of the queue.
            queueLifo := waitStartTime != 0
            if waitStartTime == 0 {
                // 记录等待开始的时间
                waitStartTime = runtime_nanotime()
            }
            runtime_SemacquireMutex(&m.sema, queueLifo, 1) // 等待在信号量上
            starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs // 如果等待的时间超过了阈值，标记需要进入饥饿模式
            old = m.state
            if old&mutexStarving != 0 {
                // If this goroutine was woken and mutex is in starvation mode,
                // ownership was handed off to us but mutex is in somewhat
                // inconsistent state: mutexLocked is not set and we are still
                // accounted as waiter. Fix that.
                if old&(mutexLocked|mutexWoken) != 0 || old>>mutexWaiterShift == 0 {
                    throw("sync: inconsistent mutex state")
                }
                delta := int32(mutexLocked - 1<<mutexWaiterShift)
                if !starving || old>>mutexWaiterShift == 1 {
                    // 如果只有自己在等待，关闭饥饿模式
                    delta -= mutexStarving
                }
                atomic.AddInt32(&m.state, delta)
                break
            }
            awoke = true // 至少被唤醒一次
            iter = 0
        } else {
            old = m.state
        }
    }

    if race.Enabled {
        race.Acquire(unsafe.Pointer(m))
    }
}
```

### 释放锁

释放的流程比较简单，区别在于不同的运行模式上。对于普通模式，通过信号量来唤醒一个等待的goroutine；对于饥饿模式，唤醒下一个等待的goroutine，并直接将锁的所有权交给它。

## `Cond`

不同的waiter可以等待在该实例上，其他线程通过`Signal`或`Broadcast`来唤醒一个或者唤醒全部的waiter。

## `RWMutex`

一个读写锁。

## `WaitGroup`

`WaitGroup`维护一个计数器，线程可以通过调用`Wait`等待计数器到0。

其他线程可以通过`Add`来增加计数器的数量，也可以通过`Done`来使计数器减一。

## `Map`

线程安全的map。

为两种场景进行的优化：
1. 一个键写入后，基本不修改，但读取很多次。
1. 并发读写的时候总是在不相交的键上修改。

### 原理

`Map`将数据分为两部分：`read`和`dirty`。
`dirty`包含了`read`的所有内容。
两者均为key是普通的key，value是`entry`类型的指针，两部分的value的`entry`的指针相同（意味着两者修改时可以同时生效）。

通过`Load`函数读取时：
1. 首先尝试去`read`查找，如果找到，返回结果。
1. 如果没有找到，则获取全局锁，作二次检查。
1. 如果没有，去`dirty`查找。

通过`Store`函数保存时：

1. 首先尝试去`read`查看key是否存在并没有被删除，如果存在那么尝试通过CAS修改。
1. 上一步没有成功时，获取全局锁。作二次尝试。
1. 如果`dirty`中有该`key`，那么将该键值对保存到`dirty`中。
1. 如果`dirty`为nil，那么从`read`拷贝出来作为新的`dirty`，然后保存该键值对。

新建`dirty`的时候，通过range遍历`read`，对于每个`entry`，首先尝试CAS删除，成功后，加入到新的`dirty`中。

如果过多的需要加锁来确认是否有数据在`dirty`上，那么就将现在的`dirty`“升级”成`read`。

总结一下，`Map`利用了冷热分离的想法，将存量数据与新增的数据进行了分离，对于存量数据的读取，能做到很好的性能。

## `Once`

维护一个计数器，来保证`Do`传入的函数只会执行依次。

## `Pool`

一个资源池。

## `PoolDequeue`

一个无锁单生产者多消费者的队列。

生产者在队头增减，消费者在队尾获取。

# 参考

- [The Go Memory Model](https://golang.org/ref/mem)

# Goroutines and Channels

[TOC]

## Goroutines

- 并发的执行单元
- 程序启动时将会创建一个用于主函数运行的goroutine，当主函数返回时，所有的goroutines被打断。

```go
go doSomethingInNewGorutines()
```

## Channels

- Goroutines之间的通信方式

### usage

```go
ch := make(chan int) // ch's type is "chan it", no buffer
ch = make(chan int, 0) // channel without buffer
ch = make(chan int, 3) // channel with buffer,capacity 3

ch <- x		// send msg to channel
x = <-ch	// receive from channel

close(ch)	// close channel

// ch <- x 
// sending to a closed channel will create a panic
x, ok = <-ch	// receiving from a closed channel is ok
if !ok {
  // channel had been closed and has no msg
}
```

- channel是底层数据结构的引用
- 使用==比较channel的含义是两个channel引用的对象是否相同

### unbuffered channel

- blocked until sender sending a msg or receiver receiving msg
- **pipeline** 

### 单方向的channel

```go
var out chan<- int	// 只发送
var in <-chan int		// 只接受
```

### buffered channel

- cap函数可以知道channel的缓冲区容量
- len函数可以知道缓冲区中有多少msg
- 当缓冲区满或空时，进行写入或读取操作会导致阻塞

### channel多路复用

```go
select {
  case <-ch1:
  case x := <-ch2:
  case ch3 <- y:
  default:
  
}
```


# pkg context

Context用于为多routine协作提供了一种方式。

## 场景

如一个web请求进入服务器，acceptor接受请求。

1. 开一个新的goroutine调用handler处理请求。
2. handler也会开启其他的goroutine进行相应的操作。
3. 这个过程不断递归的发生，将会产生一个goroutine树，树中的各个goroutine事实上共享一个运行"背景"。

**运行"背景"**其实就是上下文，各个goroutine在web请求这个场景下，会共享相同的请求参数、超时时间等。

context.Context就是为此而生的。

## 功能

- 发送一个"取消"信号
- 设定一个超时时间
- 添加一个kv对

## 关键点

- **并发安全**

- 取消/超时/获取kv对的操作都是"符合直观"的递归操作

  - 取消操作会取消该context的所有子context
  - 超时操作在设置超时时间时，检查父context的超时时间，如果自己的超时时间更晚，则不会设置超时时间 _（通过父context的超时时间来递归的取消自己）_
  - 获取kv会递归向父寻找

- ```context.WithDeadline()``` 和 ```context.WithTimeout()```实现一致为```timerCtx```，都依赖```cancelCtx```实现

- ```context.Background()``` 用于返回一个**空context**，不会被取消、超时，无kv对。

  - 用于作为所有context的根
  - _用于main function_
  - 用于初始化
  - 用于测试

- ```context.TODO()``` 返回一个空context

  - > Code should use context.TODO when it's unclear which Context to use or it is not yet available (because the surrounding function has not yet been extended to accept a Context parameter).

## usage

```go
// 创建一个新的ctx
// cancelFunc用于被调用以取消ctx
ctx, cancelFunc := context.WithCancel(context.Background())
```




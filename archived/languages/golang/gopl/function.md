# function

[TOC]

## declare

``` go
func name(parameter_list) (result_list) {
  // do something
}

func foo(arg1 int) int {
  return 1
}

func bar() (r1 int, r2 string, r3, err) {
  err = nil
  r3 = "something"
  r2 = "other thing"
  r1 = 2
  return	// better in simple function
  // or
  // return r1, r2, r3, err
  // better in large function
}
```

## 多返回值

## 递归

## 函数变量

## 匿名函数

- 闭包 closures
- **迭代变量捕获**
  - 函数闭包保存的**是指针而不是值**

## 可变参数

```go
func sum(vals...int) int {
  // vals is a slice of int
}

values := []int{1, 2, 3}

r := sum(1, 2, 3)
r = sum(valuse...)		// work well also
```

## defer

## panic

## recover

```go
func foo() {
  defer func() {
    if p := recover(); p != nil {
      // panic occured
      // do something to recover
    }
  }()
  // do something
}
```




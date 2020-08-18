# OOP

[TOC]

```go
type Point struct {
  X			int,
  Y			int
}

/*
	A method.
	p just like self or this in other language
	
	p 传值
*/
func (p Point) foo() {
  // do something
}
// 传引用
func (p *Point) foo() {
  // do something
}

var p := Point{1, 2}
p.foo()		// call foo()
```

## 组合

*Is-a OR __Has-a__*

- 包含另一个类型之后，父类型会拥有子类型的全部方法
- 编译器会自动为嵌入的类型生成包装方法

## struct的method仍然可被认为是函数值

## 封装

## 接口

```go
type InterfaceName interface {
  Foo(arg1 int) (r1 int, err error)
  Bar(arg1 string, arg2 int) (r1 int, r2 int, err error)
}
```

- **如果一个类型实现了一个接口的所有方法，那么类型就实现了该接口**
- 任何类型都实现了interface{}
- ```aInterface != nil```不意味着interface已经被赋值

### 原理

- 一个接口值包括两项
  - 动态类型
  - （实现了接口的对象的）指针
- **意味着，可能出现动态类型不为nil但是指针为nil的情况**

## 例子：error接口

```go
type error interface {
  Error() string
}
err := errors.Errorf("format string", arg1, arg2)
```


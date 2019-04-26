# array, string and slice

[TOC]

## array

- 类型包括内容类型和长度

## string

```go
// reflect/value.go
type StringHeader struct {
  Data uintptr
  Len int
}
```

- bs := []byte(aString) 会复制底层数组
- rs := []rune(aString) 会复制底层数组

## slice

```go
// reflect/value.go
type SliceHeader struct {
  Data uintptr
  Len int
  Cap int
}
```

- memory leak
  - **小心对大数据的slice引用导致gc失效**
  - **如果底层数组引用了指针，从slice移除了引用，但是底层数组仍然持有引用，导致memory leak**
  - 可以通过先设置nil来help gc
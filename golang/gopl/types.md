# types

[TOC]

类型

- 基础类型
- 复合类型
- 引用类型
- 接口类型

## 基础类型

- 数字
- 字符串
- bool

### 数字

int(int8==byte int16 int32==rune int64)

uint(uint8 uint16 uint32 uint64)

float32 **float64** *IEEE754 float64 is better*

complex64 complex128

#### 位操作

&，|，^，&^(按位清零，a &^ b，将b中为1的位，在a中置零) ，<<，>>（算术右移）

#### 算术操作

**%**（结果符号按照a % b中的a的符号）**/**（整数除法向0截尾）

#### pkg math

```go
// type limits
// math.MaxType_name
// e.g. float32
math.MaxFloat32
// check is NaN
math.IsNaN(float64(someNumber))
// check is Infinite
math.IsInf(float64(someNumber, sign))
```

### bool

true false

### 字符串

rune==int32 （utf-8字符）

string is **ONLY-READABLE**

**string is a utf8 sequence**

**len(string) will return count of bytes in string**

**string[idx] will get No.idx byte**

```go
import "unicode/utf8"

// get rune count in string
utf8.RuneCountInString(string(someString))
// one-by-one get rune in string
for i,r := range string(someString) {
  // do something
}

// maybe []rune is better than string
runeArray := []rune(string(someString))
```

#### pkg bytes/strings/strconv/unicode

- bytes.Buffer

### 常量 && iota

- 常量的值将在编译期确定
- complier将为无类型的常量提供至少256bit的运算精度

#### iota

```go
// pkg net
type Flags unit

const (
	FlagUp Flags = 1 << iota
  FlagBroadcast
  FlagLoopback
  FlagPointToPoint
  FlagMulticast
)
/*
FlagUp == 00001b
FlagBroadcast == 00010b
...
FlagMulticast == 10000b
*/
```



## 复合类型

1. **前闭后开**

### array

- 长度固定，类型固定
- 默认初始化类型的0值
- 内置函数len返回数组中元素的个数
- 数组长度是数组类型的一部分
  - TypeOf([3]int) != TypeOf([4]int)
  - a := [必须为编译期常量]int{...}
  - 两个数组可以直接比较相等，当其元素是相同的
- **将一个数组传递给函数，是传值**

### slice

- 变长
- **传递给函数是传值，其中值是指指针、len、cap三个，相当于传递了底层数组的引用**

#### 原理

1. 底层是一个数组
2. 指针 **指向slice第一个元素在对应数组的地址**、长度 **len**、容量 **cap**

#### 操作

1. 切片可以超过**原slice**的len，但是不能超过cap
2. 不能直接比较相等
   1. 对于byte型的slice，可以使用**bytes.Equal**
   2. 其他类型需要自行展开比较
3. ```s = append(s, "item")```
   1. built-in函数append用于向slice追加数据
   2. 追加的数据超过cap时，底层数组将会**新建一个长度为原有数组长度两倍的数组（具体实现可能更复杂，将原数据移入新的数组，并返回一个指向新数组的切片**
   3. **特别重要的，原有的切片都会失效（因为底层数组换新，但既有的切片指针仍指向原有数组），需要重新进行赋值来更新原有切片的指针**
   4. 内置的append可以一次性添加多个值，或着slice

### map

- 哈希map
- key、value必须分别为相同的类型
- **key必须支持使用==进行比较**
  - 所以不应该使用浮点数作为key

#### 操作

1. 初始化

   ```go
   m := make(map[string]int) // map key is string and value is int
   m := map[string]int {
     "key1": 1
     "key2": 2
   }
   m := map[string]int{} // create an empty map
   ```

2. ```delete(map_ref, key) //remove element```

3. 遍历

   ```go
   for k, v := range m {
     // do something
   }
   ```

   1. 遍历的顺序是随机的

4. 读取

   ```go
   val, ok := m[key]
   if ok {
     // if ok== true, key has already in map
   }
   
   // another example
   if val, ok := m[key]; !ok {
     // do something
   }
   ```

### struct

#### 操作

1. 定义

   ```go
   type MyStruct struct {
     Property1		int			`tagKey1:value1; tagKey2:value2`
     Property12	string
   }
   
   ```

2. usage

   ```go
   var ms MyStruct
   ms = MyStruct{1, "hello"}
   ms = MyStruct {
     Property1: 1,
     Property2: "hello"
   }
   ms.Property1 = 100 // assign Property1 100
   pMsP1 := &ms.Property1 // get Property1's pointer
   (&ms).Property1 += 102 // this also work well
   
   ```

3. compare

   如果struct中所有的成员都可用==比较，则struct也可以用==比较

4. 结构体嵌入和匿名成员

   - 匿名成员不被导出
---
title: "go tool compile用法与编译指令"
date: 2022-01-24T21:40:31+08:00
draft: false
tags:
    - go
    - how
    - memo
    - go-command
order: 3
---

记录`go tool compile`的用法和编译指令。

<!--more-->

{{% serial_index go-command %}}

# `go tool compile`的工作

`go tool compile`编译由传入的`file`组成的单独的包。
默认的，产物是一个`.o`的的中间目标文件 *(intermediate object file)* 。
目标文件可以被用来与其他的目标文件组成包集合 *(package archive)* ，
也可以直接传递给链接器使用。

生成的目标文件包含包本身暴露的符号的类型信息，
也包含包引用的其他包的符号的类型信息。
所以在编译调用一个包的包时，
只需要读取被调用的包的目标文件即可。

# 用法

```shell
go tool compile [flags] file...
```

`file`一定需要是一整个package所有的代码文件。

## 参数

- `-D path` 用于本地引用依赖的相对路径。
- `-I dir1 -I dir2` 在查询完`$GOROOT/pkg/$GOOS_$GOARCH`之后，
进一步从dir1和dir2查询需要的依赖包。
- `-L` 在错误信息中展示完整的文件路径。
- `-N` 禁用优化。
- `-S` 将code的汇编输出到标准输出。
- `-S -S` code和data都输出。
- `-V` 输出编译器的版本。
- `-asmhdr file` 将汇编的头写到file中。
- `-buildid id` 将id作为build_id写入输出的元数据中。
- `-blockprofile file` 将编译时的block profile写入到file中。
- `-c int` 编译时并发度，默认是1，表示不并发。
- `-complete` 假定包不含有非go的部分。
- `-cpuprofile file` 将编译时的CPU profile写入到file中。
- `-dynlink` 允许引用在共享库中的go符号。
- `-e` 移除错误数量的上限。
- `-goversion string` 定义使用的go tool版本，
用于runtime的版本和goversion不匹配的情况。？
- `-h` 当第一个错误被发现时停止，并输出堆栈trace。
- `-importcfg file` 从file读取配置。
配置包含importmap/packagefile。？
- `-importmap old=new` 在编译时，将对old的引用更换为new。
这个flag可以有多个来设置多个映射。
- `-installsuffix suffix` 从`$GOROOT/pkg/$GOOS_$GOARCH_suffix`查找包，
而不是`$GOROOT/pkg/$GOOS_$GOARCH`。
- `-l` 禁用内联。
- `-lang version` 使用的语言版本，如`-lang=go1.12`，
默认使用当前版本。
- `-linkobj file` 将面向链接器的目标文件写入到file。
- `-m` 输出优化决定。可以传入整数 (`-m=10`) ，越大的整数输出越详细。
- `-memprofile file` 输出编译时的内存profile到file。
- `-memprofilerate rate` 设置编译时的`runtime.MemProfileRate`为rate。？
- `-msan` 开启内存检查器 *（memory sanitizer）* 。
- `-mutexprofile file` 编译时的mutex profile写入到file。
- `-nolocalimports` 禁用本地引用/相对引用。
- `-o file` 将目标文件写入到file。
- `-p path` 判断如果增加了对于path的引用是否会出现循环引用的问题。
- `-pack` 输出打包过的格式，而不是目标文件。
- `-race` 开启数据竞争检测。
- `-s` 输出对于可简化掉的Composite literals的警告。
- `-shared` 生成可以链接到共享库的代码。
- `-spectre list` 在list中启用减轻幽灵攻击的机制。？
- `-traceprofile file` 将执行trace写入到file中。
- `-trimpath prefix` 移除记录的源文件路径的前缀。？

关于调试信息的flag：

- `-dwarf` 生成DWARF符号。
- `-dwarflocationlists` 在优化模式中，
向DWARF增加位置列表 *（location list）*。
- `-gendwarfinl int` 生成DWARF的内联信息记录。

## 编译器指令

### line指令

line指令用于定义紧接着line指令结束问值得自负在源码中的位置。

编译器会识别形如`//line`或`/*line`作为编译器指令line：

```go
//line :line
//line :line:col
//line filename:line
//line filename:line:col
/*line :line*/
/*line :line:col*/
/*line filename:line*/
/*line filename:line:col*/
```

- 开头是紧贴着`//`或者`/*`的`line`。
- 包含至少一个冒号。

### 其他的编译器指令

其他的编译器指令都是由`//go:name`的形式。

- `//go:noescape`

    noescape后续紧跟着一个没有body的函数声明
    （没有body意味着这个函数是用非go实现的）。
    noescape意味着这个函数不允许接受任何会逃逸到堆上的指针作为参数，
    或者逃逸到该函数的返回值中？。
    noescape会作用于对该函数的调用时的逃逸分析。

- `//go:uintptrescapes`

    uintptrescapes后续需要紧跟一个函数声明。
    指令表明该函数的uintptr可能是由指向被垃圾回收管理的对象的指针转换而来的。

- `//go:noinline`

    表明这个函数不可以被内联。

- `//go:norace`

    这个函数的数据访问应该被数据竞争探测器忽略。

- `//go:nosplit`

    表明这个函数需要忽略通常的堆栈溢出检查。
    用于这个函数被调用时，调用者goroutine不可以被抢占。？

    > This is most commonly used by low-level runtime code invoked 
    > at times when it is unsafe for the calling goroutine to be preempted.

- `//go:linkname localname [importpath.name]`

    编译器在目标文件的符号使用importpath.name替换掉localname。
    如果[importpath.name]被忽略，那么会使用默认的符号名，
    但产生副作用：使localname可以被其他包访问。



# 参考

- [cmd/compile的文档](https://pkg.go.dev/cmd/compile@go1.17.6)

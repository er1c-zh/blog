---
title: "go开发套件"
date: 2022-01-03T11:48:10+08:00
draft: true
tags:
    - go
    - how
    - memo
---

# Intro

go开发套件包含一组编译和构建源码的指令和对应的程序。

通常有三种方式来运行这些程序：

1. 最普遍的，通过`go subcommand`的形式，如`go fmt`。

    `go`在调用这些子命令对应的程序时，
    会传递用于处理`package`层次的参数，
    来使命令运行在所有的`package`上。

1. 另一种形式是`go tool subcommand`，被称为 **独立运行** *(stand-alone)* 。

    对于大多数指令来说，这种形式仅用来调试；
    对于`pprof`等某些指令，只能用这种形式来运行。

1. 最后，因为`gofmt`和`godoc`经常被使用，单独构建了二进制包。

## 指令

- `go` 管理go源码和运行其他的指令
- `cgo` 支持cgo特性
- `cover` 用于生成和分析单元测试的覆盖率文件
- `fix` 用于将使用了语言或lib的旧特性的程序，改写成新特性。
- `fmt` 格式化源码
- `godoc` 导出并生成go代码中的文档
- `vet` 检查代码中的可疑之处，比如`Printf`中格式化字符串和参数是否对应。

更加完整的命令集在[标准库/cmd包的文档](https://pkg.go.dev/cmd)中。

# `go`

```shell
go <command> [arguments]
```

command包含：

```plain
bug         start a bug report
build       compile packages and dependencies
clean       remove object files and cached files
doc         show documentation for package or symbol
env         print Go environment information
fix         update packages to use new APIs
fmt         gofmt (reformat) package sources
generate    generate Go files by processing source
get         add dependencies to current module and install them
install     compile and install packages and dependencies
list        list packages or modules
mod         module maintenance
run         compile and run Go program
test        test packages
tool        run specified go tool
version     print Go version
vet         report likely mistakes in packages
```

`go help <command>`可以获得每个指令更加详细的信息。

除了指令，也可以用`go help`来查询下列的主题的帮助信息。

```plain
buildconstraint build constraints
buildmode       build modes
c               calling between Go and C
cache           build and test caching
environment     environment variables
filetype        file types
go.mod          the go.mod file
gopath          GOPATH environment variable
gopath-get      legacy GOPATH go get
goproxy         module proxy protocol
importpath      import path syntax
modules         modules, module versions, and more
module-get      module-aware go get
module-auth     module authentication using go.sum
packages        package lists and patterns
private         configuration for downloading non-public code
testflag        testing flags
testfunc        testing functions
vcs             controlling version control with GOVCS
```

## `go build` 编译包和依赖

`go build [-o output] [build flags] [packages]`

`build`指令编译`packages`和对应的依赖，但是不会**安装**编译结果。

### `packages`

- 如果传入了 **在一个文件夹下的文件列表** ，指令会将这些文件当作一个（只包含这些传入的文件）package来处理。
- 忽略`_test.go`结尾的文件。
- 如果在编译单独的`main`包，`build`指令会生成一个可执行文件；反之，只编译，丢弃编译结果，用来检查编译是否可以通过。
    - 可执行文件的名称取决于：
        1. （通过`-o`指定了路径或文件名）写入到指定的文件夹和指定的文件名
        1. （如果传入了文件列表）第一个文件的名字
        1. （没有指定文件列表）源码目录的名字

### flags 参数

- `-a` 强制重新编译（缓存的编译结果）已经是最新的包。
- `-p n` 指定可以并行的数量，覆盖默认值`GOMAXPROCS`。
- `-race` 开启数据竞争的检测。
- `-msan` enable interoperation with memory sanitizer.
- `-v` 在编译一个包时，打印包名。
- `-work` 打印临时工作目录（temporary work directory），并且退出时不删除。
- `-x` 打印指令？
- `-asmfalgs '[pattern=]arg list'` 会传递给所有的`go tool asm`调用。
- `buildmode mode` 编译模式，表明需要的编译结果，
更多的可以通过`go help buildmode`查看。
    - `default`模式：输入的main包编译成可执行文件；
    其他包编译成`.a`文件。
    - `plugin`模式：编译为go插件。
- `-compiler name` 使用的编译器，gccgo/gc。
- `-gccgoflags '[pattern=]arg list'` 传递给gccgo编译器和链接器的参数。
- `-gcflags '[pattern=]arg list'` 传递给所有的`go tool compile`的参数。
- `-installsuffix suffix`
- `-ldflags '[pattern=]arg list'` 传递给`go tool link`的参数。
- `-linkshared` 将代码按照“会被与以buildmode=shared形式编译的库进行链接”的形式来编译。
- `-mod mode` 用于控制“编译指令”是否要更新`go.mod`文件和是否使用`vendor`文件夹。
默认的，如果`go.mod`中的版本大于等于1.14且存在`vendor`文件夹，那么会使用`-mod=vendor`；反之，使用`-mod=readonly`。
    - `mod` 忽略`vendor`并更新`go.mod`，比如代码中增加了新的依赖，使用`go build -mod=mod`就会自动将新的依赖添加到`go.mod`文件。
    - `readonly`忽略`vendor`；如果`go.mod`文件需要跟新，那么报告错误。
    - `vendor` 使用`vendor`机制。（go指令不会使用网络和缓存的module）
- `-modcacherw` mod cache read-write
- `-modfile file` [module-aware模式]({{< relref "post/language/go/module.md#详细一点点的解释" >}})下，
指定要使用的`go.mod`文件。
- `-overlay file` `overlay`文件提供了将编译中需要使用的文件映射到需要的路径的能力。
- `-pkgdir dir` ？替换所有的安装和加载所有的包的路径。
- `-tags tag,list` 逗号分割的编译tag列表。
- `-trimpath` （用处？）将输出的可执行文件中的所有文件系统路径（前缀）移除，分别替代为：
    - `go` 标准库
    - `module_path@version` 使用module机制
    - `plain import path` 使用GOPATH机制
- `-toolexec 'cmd args'`

对于`-asmflags/-gccgoflags/-gcflags/-ldflags`，每个都接受一组参数`arg list`用于传递给底层的程序。

- `arg list`需要包含一个pakcage pattern用于指明这些参数在处理哪些包时要被传递给的底层的程序。
- 默认的，只有在`go build packages`的`packages`中出现的包。
- 这提供了一种为不同的包使用不同的参数的能力。
- 如果需要传递给所有的包，可以使用如`go build -gcflags='all=-l -N' packages`的形式。

## Gofmt / `go fmt` 格式化源码

todo

## `go generate`

todo

## `go get`

todo

# 参考

- [Golang official document - Command Documentation](https://go.dev/doc/cmd)
- [标准库/cmd包的文档](https://pkg.go.dev/cmd)
- [关于`-mod`的文档](https://go.dev/ref/mod#build-commands)

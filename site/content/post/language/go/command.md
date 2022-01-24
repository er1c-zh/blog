---
title: "go开发套件"
date: 2022-01-03T11:48:10+08:00
draft: false
tags:
    - go
    - how
    - memo
    - go-command
order: 1
---

记录go的指令的用法。

<!--more-->

{{% serial_index go-command %}}

组织的思路是：

1. 首先按照**使用场景**来介绍`go`中一些常用的指令的使用方式。
    1. 编译
    1. 格式化源码
    1. 生成代码
    1. 管理包的依赖
    1. 维护包的依赖
    1. 运行一个项目
    1. 单元测试
    1. 执行具体实现功能的指令
    1. 扫描可能的错误
1. 然后介绍更加细节的实现方式，有的会在这一篇中，有的会链接到其他的上。
    1. 编译相关
    1. 环境变量
    1. gofmt

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
 
本文主要记录了`go`的使用方法。

更加完整的命令集在[标准库/cmd包的文档](https://pkg.go.dev/cmd)中。

# 通用的

## 指令中的`packages`参数如何被使用？

> `go help packages`

很多指令需要一个`packages`参数表示哪些包需要处理。

package通常利用 **引用路径** *(import path)* 来定义。
**引用路径**有两种形式，一个是绝对或相对路径指向的文件夹中的包；
另一个是基于`GOPATH/src/`指向的包。

如果没有传入`packages`，那么表示当前文件夹下的包。

有四个**引用路径**是保留的：

- `main` 表示独立可执行文件的最外的包。
- `all` 表示在GOPATH中的所有的包。
- `std` 表示所有的go标准库。
- `cmd` 表示所有的go command和相关的内部包。

对于一个**引用路径**，如果包含一个或多个的`...`通配符，
表示是一个“模式”。`...`可以匹配任何字符串，包括空字符串和包含`/`的字符串。
特别的，

1. `/...`结尾时，可以匹配空字符串。
比如`net/...`可以匹配到`net`。
1. 包含斜杠的模式，如果没有`vendor`，是不会匹配`vendor`路径中的包。

对于由`.`和`_`开头的或`testdata`，
会被忽略。

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

## 编译包和依赖 `go build`

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

### build flags 编译参数

- `-a` 强制重新编译（缓存的编译结果）已经是最新的包。
- `-p n` 指定可以并行的数量，覆盖默认值`GOMAXPROCS`。
- `-race` 开启数据竞争的检测。
- `-msan` enable interoperation with memory sanitizer.
- `-v` 在编译一个包时，打印包名。
- `-work` 打印临时工作目录（temporary work directory），并且退出时不删除。
- `-x` 打印指令？
- `-asmfalgs '[pattern=]arg list'` 会传递给所有的`go tool asm`调用。
- `buildmode mode` 编译模式，表明需要的编译结果，
更多的可以通过`go help buildmode`查看，[也可以看这里](#编译模式-build-mode)。
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
- `-tags tag,list` 逗号分割的编译tag列表，[详细的用法在这里](#编译约束-build-constraints)。
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

## 格式化源码 Gofmt / `go fmt`

```shell
go fmt [-n] [-x] [packages]
```

`go fmt`相当于在`packages`上执行`gofmt -l -w`，
效果是将格式化之后的代码写回文件，并在标准输出列出修改过的文件。

更多的关于`gofmt`参见[这里](#gofmt指令)。

- `-n` 

## 按照文件中的指令进行以生成、更新go代码的工作 `go generate`

`go generate`会扫描文件中形如：

```go
//go:generate command argument list
```

的指令。

其中`command`指明需要运行的程序，可以是如下的形式：

- 在shell path中
- 绝对路径
- 指令的别名

**特别的，`go generate`不会解析文件，
所以在注释或者多行字符串中符合形式的字符串也会被认为是需要运行的指令。**

`argument list`是传递给`command`的参数，
空格分隔的token或者双引号表示字符串参数。

为了表明某些代码是生成的，
需要在这些文件的开头注明符合以下格式的行作为标记：

```plain
^// Code generated .* DO NOT EDIT\.$
```

`go generate`会在调用`command`时设置一些环境变量，
如`$GOARCH`执行的cpu架构等。

`//go:generate -command foo go tool foo` 可以将`foo`定义为`go tool foo`的别名，用于上述的`command`中。

`go generate`运行时会表明使用tag `generate` ，
这样可以使得一些文件只被`go generate`识别，但在`build`时被忽略。

`go generate`接受`-run=""`参数，指明要执行的指令的正则表达式。

## 增加当前模块的依赖并且安装这些依赖 `go get` 和 `go install`

具体的指令使用[看这里]({{< relref "post/language/go/module.md#指令" >}})。

`go get`与`go install`的职责分别是“管理依赖”和“安装模块”。
这一变化也是随着go版本的迭代会逐渐明确。

### `go install`

`go install`也会将可执行文件安装到`GOBIN`目录下。

```shell
go install [build flags] [packages]
```

出于消除歧义的目的，参数需要满足：

- packages需要是一下情况的一种：
    - 包路径
    - 描述包的pattern
    - 不能是标准库的包，meta-patterns(std, cmd, all)，文件的相对路径或绝对路径。
- 所有的参数需要有相同的版本后缀。
- 所有关联的在同一个模块的包需要使用同样版本的模块。
- 所有关联的模块，在被认为是main与否时，不能有不同的表现。
    - 如当一个模块是main时，go.mod文件中的`replace`指令会生效，导致这种情况。
- 使用“包路径”来描述时，一定要指向main包。

## 列出包或模块

`go list`

todo

## 模块维护

`go mod`指令提供维护模块的能力。

具体的指令使用[看这里]({{< relref "post/language/go/module.md#指令" >}})。

## 编译、运行程序

`go run`编译并运行传入的main包。

```go run [build flags] [-exec xprog] package [arguments...]```

用法：

- `build flags` 与[之前]("#build flags 编译参数")相同
- `-exec` 运行产物的方式：
    - 默认的会直接执行。
    - 如果传入了`-exec {{executor}}`参数，会使用`executor a.out arguments`来执行。
    - 如果没有设置`-exec`，而且编译产物不是 **GOOS** 和 **GOARCH** ，
    而且`go_$GOOS_$GOARCH_exec`在path中存在，那么会使用
    `go_$GOOS_$GOARCH_exec a.out arguments`来执行。使用的场景是交叉编译。

## 测试

`go test [build/test flags] [packages] [build/test flags & test binary flags]`

`go test`会重新编译每个包中的`*_test.go`文件。
每个列出的包会生成一个单独的测试可执行文件。

定义了`*_test`包的测试文件会被编译成单独的包，然后和主test包链接、执行。

`go tool`会忽略`testdata`文件夹。

### `go vet`检测

默认的，`go test`会执行`go vet`来检测明显的错误，
如果`go vet`发现了错误，那么测试就会被提前中断，实际代码不会被执行。

**如果想要关闭vet检测，可以使用参数`-vet=off`。**

### 两种执行模式

第一种，本地目录模式。当没有设置`packages`参数时，会使用这种模式。

- `go test`处理在当前目录下的测试
- 不使用缓存
- 输出测试结果的总结

第二种，传递了包的模式。
`go test`测试列出的包。

- 如果没有问题，只显示ok；有问题或者传入了`-v`或者`-bench`，会打印出所有的输出。
- 缓存成功的测试结果，来避免重复的测试。
    - 判定是否命中缓存，由是否是同一个二进制测试文件和传入的参数是否是“可缓存的”来决定。
    - **避免命中缓存，可以通过传入不可缓存的参数来实现，比如`-count=1`。**

### `go test`接受的参数

这里列出的是只传递给`go test`的参数。

- `-args` 该参数会被原封不动 (uninterpreted and unchanged) 的传递给测试二进制文件。
    - 可以用来避免参数被`go test`解析和重写。
- `-c` 只编译不执行。
- `-exec xprog` 使用xprog来执行测试二进制程序。
- `-i` 安装测试依赖的包，不执行测试。
- `-json` 将输出格式化为json格式，用于自动化处理结果。
- `-o` 测试二进制文件的名字。

### 测试二进制文件接受的参数

`go test`与生成的测试可执行文件都接受一组参数，
用来控制测试的运行方式。

需要注意的：

1. 所有的下列的参数接受一个`test.`前缀，例如`-test.v`。**如果直接执行生成的可执行文件，那么这个前缀是必须的。**

常用的：

- `-run regexp` 运行符合正则的测试。
- `-bench regexp` 只运行匹配该正则的benchmark，`-bench .`或者`-bench=.`可以用来运行所有的benchmark。
- `-vet list` 逗号分割的vet检测列表,`-vet off`关闭检测。
- `-failfast` 快速失败，发现了错误就不再执行了。
- `-count n` 运行每个test和benchmark n次。
- `-cover` 开启覆盖分析。 *（因为会插入代码来计算覆盖率，所以出错时展示的行号可能不可信）*
- `-v` verbose模式。

其他：

- `-benchtime 1h30s` `-benchtime 100x` 使得每一个benchmark运行到1h30s或者100次。
- `-covermode set,count,atomic` 覆盖分析的模式。
    - `set` 每个语句是否运行。
    - `count` 每个语句运行的次数。
    - `atomic` 并发安全的`count`。
- `-coverpkg pattern1,pattern2` 为匹配模式的包运行覆盖分析。
- `-cpu 1,2,4` 分别在`GOMAXPROCS`为1、2、4的情况下运行测试。
- `-list regexp` 不执行，只列出符合正则的test/benchmark/example。
- `-parallel n` 允许调用`t.Parallel`的测试执行。
- `-short` ？
- `-timeout d` 测试可执行文件如果运行超过d时，抛出panic。

性能检测：

- `-benchmem` 打印benchmark的内存分配统计信息。
- `-blockprofile block.out` 将goroutine阻塞情况写到`block.out`中。
- `-blockprofilerate n` 调整阻塞情况的采样率。
- `-coverprofile cover.out` 将覆盖分析写入到`cover.out`。
- `-cpuprofile cpu.out` 将CPU分析写入到`cpu.out`中。
- `-memprofile mem.out` 内存分配分析。
- `-memprofilerate n` 更精准的内存分配采样。
- `-mutexprofile mutex.out` 互斥量竞争情况。
- `-mutexprofilefraction n` 每n个采样一个获得发生竞争的互斥量的goroutine的堆栈信息。
- `-outputdir directory` 所有的信息的输出目录。
- `-trace trace.out` 将执行trace写入到`trace.out`。

## 执行指定的指令 `go tool command`

`go tool [-n] command [args...]`

- `-n` 只打印指令，不执行。

## 扫描可能的错误

```shell
go vet [-n] [-x] [-vettool prog] [build flags] [vet flags] [packages]
```

# 更加详细的细节

## 编译

### 编译约束 build constraints

aka build tag

用于描述一个文件在什么情况下需要被包含在package中。

例子：

```go
//go:build (linux && 386) || (darwin && !cgo)
```

约束了编译该文件的操作系统、指令集、cgo。

要求：

- 可以出现在所有的源码文件中（不仅是go）。
- 需要在文件的靠前的位置，再之前只能有空白行或其他注释。
- 为了区分编译约束和package文档，编译约束需要接一个空行。
- 一个文件只能有一个编译约束。

编译约束支持以下的类型：

- 目标操作系统
- 目标指令集
- 使用的编译器，比如gc或者gccgo。
- 是否支持cgo。
- go的版本，如`go.1.12`。
- 通过`-tags`给出的其他tag。

如果一个源码文件在去除扩展名和可能的`_test`后缀后，
匹配到类似：`*_GOOS/*_GOARH/*_GOOS_GOARCH`的文件名时，
等价于追加了操作系统或指令集的编译约束。

对于1.16之前的版本，使用的是`//+build`作为编译约束的前缀。

### 编译模式 build mode

`go build`和`go install`接受一个编译模式，
用来控制输出产物的类型。

- `archive` 构建给出的所有非main包为`.a`文件。
- `c-archive` 构建给出的main包和所有它引用的包，产出C archive文件。
导出的可调用符号 *(callable symbols)* 只有使用cgo的`//export`标记的函数。
- `c-shared` 构建给出的main包和所有它引用的包，产出C共享lib。
导出的可调用符号 *(callable symbols)* 只有使用cgo的`//export`标记的函数。
- `default` 默认的参数。
构建给出的main包，产出可执行文件。
构建所有的非main包，产出`.a`文件。
- `shared` 构建所有的非main包，产出后续用于`-linkshared`的单个共享lib。
- `exe` 构建给出的main包，产出可执行文件。
- `pie` 构建给出的main包和它引用的包，产出位置无关可执行文件(Position indenpendent executable)。
- `plugin` 构建给出的main包和它引用的包，产出Go plugin。

### 编译和测试缓存

go指令会缓存编译的输出并且尝试在未来的构建中使用。
默认的位置在当前操作系统的用户缓存目录下的go-build子文件夹中。
位置可以通过`GOCACHE`控制。

缓存会周期性的清理，或者使用`go clean -cache`来清理。
通常不会需要手动清理编译缓存，
go指令会探知源文件的修改、编译选项的变更等会影响编译结果的变化，
**但C库的变化不会被感知**，所以如果使用了C库，且发生了变动，
需要手动的清理编译缓存。

## 环境变量

go指令会通过环境变量来进行一些设置，
如果相关的环境变量没有设置，
那么会使用默认值。

操作：

- 查看环境变量NAME `go env <NAME>`
- 修改环境变量NAME `go env -w <NAME>=<VALUE>`
    - 默认的，保存到`GOENV`指向的文件中。

- 通用的
    - `GO111MODULE` 控制是否运行在`module-aware`模式下。
    - `GCCGO` `go build -compiler=gccgo`使用的指令。？
    - `GOARCH` 目标指令集。
    - `GOBIN` `go install`的目标路径。
    - `GOCACHE` 构建缓存存放的文件夹。
    - `GOMODCACHE` 下载的模块的缓存的文件夹。
    - `GODEBUG` 激活多种调试机制。？
    - `GOENV` 存储go环境变量的文件。
    - `GOFLAGS` 空格分割的`-flag=value`列表，
    当要执行的go指令支持这些flag时，会传递给go指令。
    优先级低于直接在命令中给出的flag。
    - `GOINSECURE` 一组逗号分割的模块通配符，
    符合的模块会被使用不安全的方法来获取。
    只在直接获取的模块上生效。
    - `GOOS` 编译目标的操作系统。
    - `GOPATH` 指明需要从哪些地方来获取go代码，
    在使用模块时不再用于解析引入的包。
        - unix下，是冒号分割的字符串。
        - windows下，是分号分割的字符串。
    - `GOPROXY` go模块代理的地址。
    - `GOPRIVAGE,GONOPROXY,GONOSUMDB` 逗号分割的模块前缀通配模式，
    符合模式的模块会直接获取 *(be fetched directly)* ，
    也不会进行校验和校验。
    - `GOROOT` go树的根。？
    - `GOSUMDB` 需要使用的校验和数据库。
    - `GOTMPDIR` go指令使用的临时文件夹。
    - `GOVCS` 会用来尝试匹配服务器？的版本控制指令。
- 用于`cgo`的 todo
    - AR The command to use to manipulate library archives when building with the gccgo compiler. The default is 'ar'.
    - CC The command to use to compile C code.
    - CGO_ENABLED Whether the cgo command is supported. Either 0 or 1.
    - CGO_CFLAGS
        - Flags that cgo will pass to the compiler when compiling
        - C code.
    - CGO_CFLAGS_ALLOW
        - A regular expression specifying additional flags to allow
        - to appear in #cgo CFLAGS source code directives.
        - Does not apply to the CGO_CFLAGS environment variable.
    - CGO_CFLAGS_DISALLOW
        - A regular expression specifying flags that must be disallowed
        - from appearing in #cgo CFLAGS source code directives.
        - Does not apply to the CGO_CFLAGS environment variable.
    - CGO_CPPFLAGS, CGO_CPPFLAGS_ALLOW, CGO_CPPFLAGS_DISALLOW
        - Like CGO_CFLAGS, CGO_CFLAGS_ALLOW, and CGO_CFLAGS_DISALLOW,
        - but for the C preprocessor.
    - CGO_CXXFLAGS, CGO_CXXFLAGS_ALLOW, CGO_CXXFLAGS_DISALLOW
        - Like CGO_CFLAGS, CGO_CFLAGS_ALLOW, and CGO_CFLAGS_DISALLOW,
        - but for the C++ compiler.
    - CGO_FFLAGS, CGO_FFLAGS_ALLOW, CGO_FFLAGS_DISALLOW
        - Like CGO_CFLAGS, CGO_CFLAGS_ALLOW, and CGO_CFLAGS_DISALLOW,
        - but for the Fortran compiler.
    - CGO_LDFLAGS, CGO_LDFLAGS_ALLOW, CGO_LDFLAGS_DISALLOW
        - Like CGO_CFLAGS, CGO_CFLAGS_ALLOW, and CGO_CFLAGS_DISALLOW,
        - but for the linker.
    - CXX
        - The command to use to compile C++ code.
    - FC
        - The command to use to compile Fortran code.
    - PKG_CONFIG
        - Path to pkg-config tool.

- 用于特定的指令集的 todo
- 特殊目的的 todo
    - GCCGOTOOLDIR
    - GOEXPERIMENT
    - GOROOT_FINAL
    - GO_EXTLINK_ENABLED
    - GIT_ALLOW_PROTOCOL
- 其他的`go env`提供的信息，但不是从环境变量读取的。
    - GOEXE
    - GOGCCFLAGS
    - GOHOSTARCH
    - GOHOSTOS
    - GOMOD
    - GOTOOLDIR
    - GOVERSION


## gofmt指令

Gofmt格式化go源码，使用tab缩进，空格对齐（注释、结构的字段类型等）。

如果没有指定路径，那么gofmt处理标准输入。
如果指定了一个文件，那么处理文件。
如果指定了文件夹，那么递归的处理文件夹中的所有.go文件。

默认的，gofmt将格式化之后的源码输出到标准输出（而不是直接写回）。

```shell
gofmt [flags] [path ...]
```

参数：

- `-d` 不输出格式化之后的源码，而是将格式化前后的diff输出到标准输出。
- `-e` ？输出所有的异常。
- `-l` 不输出格式化之后的源码，而是输出格式化前后不同的的文件名。
- `-r rule` 在格式化之前将重写规则 *（rewrite rule）* 应用到源码上。
- `-s` 尝试简化代码。
- `-w` 不输出格式化之后的源码，而是将格式化之后的代码写回到文件。

# 参考

- [Golang official document - Command Documentation](https://go.dev/doc/cmd)
- [标准库/cmd包的文档](https://pkg.go.dev/cmd)
- [关于`-mod`的文档](https://go.dev/ref/mod#build-commands)
- [gofmt的文档]](https://pkg.go.dev/cmd/gofmt@go1.17.5)

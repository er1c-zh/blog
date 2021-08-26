---
title: "有关go module的一点点记录"
date: 2021-08-25T20:59:19+08:00
draft: true
tags:
    - go
    - how
---

这是关于go的包管理的入门学习记录。

<!--more-->

# 模块相关的指令

todo

# 根据包路径查找依赖的模块

`go`指令会首先从**build list**来寻找符合包路径的前缀的模块。
如果有且仅有一个**build list**中的模块能够提供符合要求的包，
工作会正常继续；反之，会产生一个错误。
可以向`go`指令增加参数`-mod=mod`来指示`go`指令尝试从可能的地方来获取能够提供缺失的包的模块，
并更新`go.mod`和`go.sum`文件。
特别的，对于`go get`和`go mod tidy`指令，会自动的进行这个操作。

## 具体实现

当提供了`-mod=mod`参数时，`go`指令会尝试去寻找能够提供缺失的包的模块。

1. 首先，检查环境变量`GOPROXY`

    `GOPROXY`的值是由逗号分割的字符串列表，
    可能的值有模块代理的地址、
    表明直接与版本控制系统交互的`direct`和不需要做任何操作的`off`。
    除此之外，环境变量`GOPRIVATE`和`GONOPROXY`也会影响查询模块的行为。

1. 遍历`GOPROXY`的所有项

    根据缺失的包的路径，尝试所有可能的模块路径来获取最近版本的模块代码。

    如果有多个模块能够提供需要的包，会按照最长匹配原则来选取模块；
    如果查找到了至少一个模块，但没有模块能够提供需要的包，返回一个错误；
    如果没有请求到模块，会尝试`GOPROXY`中下一项，如果没有下一项，返回错误。

1. 如果找到了符合要求的模块，会向`go.mod`文件中增加新的`require`记录

# `go.mod`文件

`go.mod`文件保存了当前模块依赖的包和版本，并定义自身所在的文件夹的模块。

文件以行为单位组织，
每一行由指示的关键字和参数构成。

## 从例子出发了解`go.mod`的格式

```gomod
/*
cstyle的块注释
*/
// cstyle的行内注释
module example.com/my/thing // module关键字定义了一个模块，第一个参数是模块路径

go 1.12 // go关键字指示模块使用的go版本

require example.com/other/thing v1.0.2 // 指明主模块依赖的模块和最小的版本
/*
indirect表明这个依赖没有被主模块的代码直接引用
*/
require example.com/new/thing/v2 v2.3.4 // indirect
exclude example.com/old/thing v1.2.3 // exclude 排除了这个版本 只有本文件是主模块的go.mod时生效
require (
    example.com/new/thing/v2 v2.3.4
    example.com/old/thing v1.2.3
)
replace example.com/bad/thing => example.com/good/thing v1.4.5 // replace 利用后面的版本替换前面的 只有本文件是主模块的go.mod时生效
replace example.com/bad/thing v1.4.5 => example.com/good/thing v1.4.5 // replace 利用后面的版本替换前面的这个版本的依赖
retract [v1.9.0, v1.9.5] // retract 指明该版本的模块不应该被依赖，用于处理有问题的版本
```

# **最小版本选择** *(minimal version selection)* 算法

go使用最小版本选择算法计算出每个依赖的模块的版本用于编译。

最小版本选择算法从表示模块依赖关系的有向图中构建出编译需要的**build list**。
图中的顶点表示一个模块与版本的组合。
边表示`go.mod`文件中`require`关键字表明的对一个模块的依赖，以及依赖的最小版本。
图会被**主模块**的`go.mod`文件中的`replace`和`exclude`关键字影响。

最小版本选择算法从**主模块**顶点开始，
遍历整个图，追踪所有依赖的模块的最大版本，
最后输出这些版本作为**build list**。
至于为什么叫做最小版本选择算法，
因为输出的版本是满足**主模块**的“**最小版本**”。

## replace

可以在主模块的`go.mod`文件中增加`replace`关键字来替换某个模块的某个版本或某个模块的所有版本为指定的版本。

## exclusion

可以在主模块的`go.mod`文件中增加`exclude`关键字来排除某个模块的某个版本。

# GOPROXY与访问相关

todo

## goproxy协议

todo

## 与版本管理系统的交互

todo

## 私有库与鉴权

todo

# GOSUM

todo

# 环境变量参数速查

# 名词表

- **build list** 编译指令使用的模块及版本列表，
由主模块的`go.mod`文件和依赖的模块的`go.mod`文件通过**最小版本选择** *(minimal version selection)* 算法合并构成。

- **package** 包，一组在相同目录的一同编译的源码文件集合。

- **package path** 包路径，可以唯一标识一个包的路径，
由**模块路径**拼接上包在模块的路径来构成。

- **module** 模块，一组一同发版、分发的包的集合。

- **module path** 模块路径，用于标识模块的路径，
还被用来当作模块中的包的 **引用路径** *(import paths)* 的前缀。

- **module proxy** 模块代理，实现了GOPROXY协议的网络服务器，
`go`指令可以从模块代理下载模块的版本信息、`go.mod`文件和模块的压缩文件。

# 参考

- [golang.org/ref/mod](https://golang.org/ref/mod)

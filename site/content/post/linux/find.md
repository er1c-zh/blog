---
title: "find"
date: 2023-06-29T16:43:06+08:00
draft: false
tags:
    - memo
    - linux
    - command
---

search for files in a directory hierarchy

<!--more-->

# TL;DR

```shell
# 1. 搜索 .md 文件
find . -name '*.md' -print

# ...
# ./readme.md

# 2. 以 egrep 正则来搜索
find . -regextype posix-egrep -regex '.*/java.*.md' -print
```

# Intro

GNU `find` 遍历传入的文件树，
在每个文件名上执行从左到右的匹配，
直到本次匹配结束（短路匹配）。

```shell
find [-H] [-L] [-P] [-D debugopts] [-OLevel] [path...] [expression]
```

其中 `-H -L -P` 控制如何处理链接。
`path`是文件名或文件夹名，用于指定执行匹配的目标。
直到遇到`- ( ~`之前，都作为`path`处理。
剩下的都会作为`expression`处理。

`path`默认值是当前目录，
`expression`默认值是 `-print` 。

# 选项

## 如何处理链接

`-P` 默认值，不追踪链接，使用链接的名称、属性匹配。

`-L` 追踪链接，使用链接指向的文件的名称、属性来匹配。

`-H` 除了处理参数时，不跟踪链接。

## 测试选项 `-D debugoptions`

`tree` 输出表达式树原始和优化后的情况。

`stat` 打印使用`stat`和`lstat`系统调用的情况。

`opt` 打印表达式树的诊断信息。

`rates` 打印每个谓词（predicate）成功或失败的情况。

## -Olevel 开启查询优化

# 表达式

表达式由 options、tests 和 actions 的操作组成。

如果除了 `-prune` 没有其他的 action，
`-print` 会被应用于所有表达式为 true 的 file。

## OPTIONS

所有的options都返回true。
除了`-daystart` `-follow` `-regextype`，
options都影响所有的tests，尽管options在tests后面给出。
`-daystart` `-follow` `-regextype`只会影响后面的tests。

`-d` `-depth` 先处理目录的内容，然后处理目录本身。`-delete` 的表现也是如此。

`-daystart` 从今天的开始计算时间，而不是24小时之前。

`-follow` 被`-L`替代。

`-ignore_readdir_race` 关闭遍历文件夹时文件夹被删除的错误提醒。

`-maxdepth levels` 最多深入 *levels* 层。

`-mindepth levels` 小于 *levels* 时不执行任何 tests 和 actions。

`-mount` `-xdev` 不要进入其他文件系统。

`-noignore_readdir_race` 关闭`-ignore_readdir_race`。

`-noleaf` 关闭基于一个文件夹的子文件夹数量 = 文件夹硬链接 - 2 的优化。
 
    > 通常，Unix 文件系统拥有至少两个硬链接： `.` 和本身的名字。

`-regextype type` 修改使用的 regex 类型，目前支持 emacs (default) / posix-awk / posix-basic / posix-egrep / posix-extended 。

`-xautofs` 不进入在 autofs 文件系统上的文件夹。

## TESTS

某些 tests 支持比较被遍历的文件和命令行指定的文件。
其中，双方都会按照 `-H -L -P` 来获取信息，命令行指定的文件的信息只会获取一次。

数值类型的参数可以以如下的格式给出：

`+n` 大于 n ， `-n` 小于 n ， `n` 等于 n。

`-amin n` 文件最后的访问时间位于 n 分钟之前。

`-anewer file` 文件被访问的时间比 file 更改的时间更近。

`-atime n` 文件最后访问时间 n*24 小时之前。需要注意的是，小数部分会被丢弃；换句话说， `-atime +1` 只匹配至少 **2** 天内没访问过的文件。

`-cmin n` `-cnewer file` `-ctime n` 文件的状态，类似 文件的最后访问时间。

`-empty` 文件是空的且是普通文件 (regular file) 或文件夹。

`-executable` 匹配可以执行的文件和可以被搜索的文件夹。

`- false` false 。

`-fstype type` 在 type 文件系统上的文件。

`-gid n` group ID 是 n 的文件。

`-group gname` 属于 gname 组的文件。

`-ilname pattern` 大小写敏感的 `-lname` 。

`-iname pattern` 大小写敏感的 `-name` 。

`-inum n` 文件的 inode 号码是 n 。通常可以使用 `-samefile` 来替代。

`-iregex pattern` 大小写敏感的 `-regex` 。

`-iwholename pattern` 大小写敏感的 `-wholename` 。

`-links n` 拥有 n 个链接的文件。

`-lname pattern` 匹配内容符合 pattern 的符号链接。

`-mmin n` 文件修改时间在 n 分钟之前。

`-mtime n` 文件修改时间在 n * 24 小时之前。类似 `-atime n`的处理逻辑。

`-name pattern` 文件名符合 pattern 的文件，不包括路径。

`-newer file` 文件修改时间比 file 更近。

`-newerXY reference` 比较当前文件的 timestamp 和 reference 。 
`reference` 可以是一个文件，也可以是一个描述绝对时间的字符串。
`X` 和 `Y` 是一些字符的占位符，
这些字符用来选择用于比较的时间。

`-nogroup` 用于筛选没有合法 group 的文件。

`-nouser` 没有合法的 user 。

`-path pattern` 文件名称符合 pattern 。

`-perm mode` 精准匹配权限 mode 。

`-perm -mode` 匹配有 mode 给出的所有权限的文件。

`-perm /mode` 匹配有 mode 给出的任意权限的文件。

`-readable` 匹配可读文件。

`-regex pattern` 完整路径匹配 pattern 的文件。

`-samefile name` 匹配指向 name 文件的所有引用。

`-size n[cwbkMG]` 占用了 n 个单元的文件。

`-true` true.

`-type C` 文件类型是 c 的文件。

    - b 块文件
    - c 字符文件
    - d 文件夹
    - p 命名管道
    - f 普通文件
    - l 符号链接
    - s 套接字
    - D 门

`-uid n` 文件 user ID 是 n 。

`-used n` 自从文件状态被修改之后 n 天被访问过。

`-user uname` 文件用拥有者是 uname 。

`-writable` 可修改的文件。

`-xtype c` 除了链接与 `-type` 相同。

# ACTIONS

`-delete` 删除。

`-exec command ;` 执行命令，所有分号之前的字符串都作为 command 的参数。

`-exec command {} +` 类似 xargs 的 `-exec command ;`

`-execdir command ;` `-execdix command {} +` 文件夹版本的 `-exec` 。

`-fls file` 将 `-ls` 输出到 `file` 中。

`-fprint file` 输出文件的全路径到 file 中。

`-fprint0 file` 输出到 file 的 `-print0` 。

`-fprintf file format` 输出到 file 的 `-fprint` 。

`-ls` 输出当前文件的 `ls -dils` 。

`-ok commond ;` `-okdir command ;` 需要确认的 `-exec` 。

`-print` 输出文件的路径 + 文件名 + 一个换行符。

`-print0` 空行替换为 null 字符的 `-print` 。

`-printf format` 输出 format 。

`-prune` 如果文件是一个文件夹，不进入。

`-quit` 立即退出。

# OPERATORS

按照优先级下降的列出。

`( expr )` 强制优先级。

`! expr` 

`-not expr`

`expr1 expr2` 短路的 and 链接。

`expr1 -a expr2` and.

`expr1 -and expr2` and.

`expr1 -o expr2` or.

`expr1 -or expr2` or.

`expr1 , expr2` 两个表达式都会执行， expr1 的返回值会被丢弃，将 expr2 的返回值作为表达式的返回值。

# Refefrence

- [find(1) manual](https://linux.die.net/man/1/find)

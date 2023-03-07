---
title: "lsof"
date: 2023-03-07T14:28:18+08:00
draft: false
tags:
    - memo
    - linux
    - command
---

list open files.

<!--more-->

# 总览

`lsof`列出打开的文件。

各种option可以用来筛选不同的文件，如普通的文件、socket等等。
默认的，各种条件是“或”组合一起的。
可以通过`-a`flag来标记使用“与”组合。

不同的flag可以用一个前缀，如`-`，组合起来：
`lsof -a -b -C` 与 `lsof -abC` 一致。
但是要注意缩写flag可能产生歧义。

# 输出格式

- `COMMAND` 命令
- `PID` pid
- `TYPE` 类型

# 部分参数

## 文件夹 `+d s` `+D s`

搜索所有在s一层中的文件(`+d`)，或者包含递归的文件(`+D`)。

## 网络

### `-i`

`-i [i]`匹配给出的地址。

格式：`[46][protocol][@hostname|hostaddr][:service|port]`

- 46 表示IP版本
- 协议 TCP/UDP
- hostname 互联网主机名
- hostaddr aaa.bbb.ccc.ddd
- service 一个`/etc/services`中的名称，比如smtp。
- port 端口，也支持list。

> -i6 - IPv6 only
> TCP:25 - TCP and port 25
> @1.2.3.4 - Internet IPv4 host address 1.2.3.4
> @[3ffe:1ebc::1]:1234 - Internet IPv6 host address
>     3ffe:1ebc::1, port 1234
> UDP:who - UDP who service port
> TCP@lsof.itap:513 - TCP, port 513 and host name lsof.itap
> tcp@foo:1-10,smtp,99 - TCP, ports 1 through 10,
>     service name smtp, port 99, host name foo
> tcp@bar:1-smtp - TCP, ports 1 through smtp, host bar
> :time - either TCP, UDP or UDPLITE time service port

### `-U`

只展示Unix domain socket files。

## pid

`lsof -p pid` 根据pid列出打开的文件。

# Reference

- [lsof manual](https://linux.die.net/man/8/lsof)

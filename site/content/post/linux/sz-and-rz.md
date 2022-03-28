---
title: "利用sz与nz发送和接收文件"
date: 2022-03-01T09:35:27+08:00
draft: false
tags:
    - linux
    - command
    - how
---

当访问线上机器有很严格的网络隔离时，
其他的方法如开一个HTTP服务器或者nc开启一个端口等因为网络的隔离而无法使用。

如果可以ssh到目标机器（包括走跳板机），
可以尝试使用sz和nz两个工具来实现发送和接受文件。

# TL;DR

**前提是发送的机器需要安装`sz`，接收端需要安装`rz`。**

## 基于iterm2 trigger机制

参考[一个github的repo](https://github.com/aikuyun/iterm2-zmodem)。

## 基于`zssh`

1. 安装sz和rz和zssh。

    ```shell
    brew install lrzsz zssh
    ```

1. 利用`zssh`连接到远程机器。

    ```shell
    # 使用与ssh相同
    zssh root@192.168.1.1
    ```

1. 传输文件

    - 发送本地到remote

        ```shell
        # 连接到远程机器上之后
        # 按下ctrl-2 是zssh的触发键，会进入文件传输模式
        输入sz <filename-to-send>
        ```

    - 从远程接收文件

        ```shell
        # 在远程机器上执行
        sz <filename-to-receive>
        # 按下ctrl-2
        输入 rz
        ```

![sz-rz-by-zssh.gif](https://dsm01pap001files.storage.live.com/y4mXeS47HglpvMN3iW9JMXtRVEZBid4c9BgsgNEYpkqf6axebTHw9Wu9YgfbKjwcP3u2FkqRVra97dOSEOdL1KJn-5BB2GqyWmVw1p1uFamMYIBtKeVRdHxDXqTbuo2KjjSXDnoD19FOy8M7t0-nSceqeY_DIQDHBiAkJ3pTVF6qgj8Lzv0hhYkI-HxOCLtxF4k?width=1024&height=576&cropmode=none)

# 原理

`sz`和`rz`中的s和r是send和receive的首字母，
z表示的文件传输协议zmodem。

以上的方法的原理都是在发送端运行sz按照协议发送文件内容，
在接收端运行rz按照协议接收文件。

这也是为什么需要在两端都安装好`sz`和`rz`。

## 关于「基于iterm2 trigger机制」

iterm2提供了监听输出并做出相应动作的能力，
在监听到`sz`或`rz`的关键字时，
触发选择文件等流程，
完成传输。

## 关于`zssh`

是对`ssh`的一个包装，
`ctrl-2`是激活文件传输模式的触发键。

# 为什么要选择`sz`和`rz`

- 目标机器网络环境比较严苛，只能ssh上去。
- ssh上去需要经过复杂的跳板机制，不能使用scp或ssh等指令简单的跳板。

# 参考

- [https://github.com/aikuyun/iterm2-zmodem](https://github.com/aikuyun/iterm2-zmodem)
- [zssh](http://zssh.sourceforge.net/)
- [zmodem wiki](https://zh.wikipedia.org/wiki/ZMODEM)


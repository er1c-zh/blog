---
title: "nc and pv"
date: 2020-08-18T22:34:28+08:00
draft: true
tags:
    - linux
    - command
    - how-to
---

# 如何从线上机器拉取数据

最近遇到了一些线上问题，希望能够把线上抓到的信息拉到本地处理。
实现的方式有很多，比如利用python开启一个http server等。

但是我还是偏好使用较为广泛、各种发行版自带的命令来实现。 

辗转中了解到`nc`指令以及可以用来限流的`pv`指令，两者结合，可以满足绝大部分网络拷贝的场景。

## how-to

*作为how-to系列的特色，开篇就会提供一个可用的方案。*

在线上机器执行，会监听1220端口，并把数据发送到该链接上。

```shell
echo data_file | nc -l 1220
```

在本地上执行，连接到线上机器，并把数据写入到local_data_file。

```shell
nc ip_or_hostname 1220 > local_data_file
```

## nc

> arbitrary TCP and UDP connections and listens

手册上，netcat的简介暗示了这是一个很强大的网络向的指令。

`nc`能够

1. 发起TCP链接，发送UDP包
1. 作为服务端，监听任意端口
1. 进行端口扫描
1. 支持IPv4和IPv6

### 用法

```shell
nc [-46bCDdFhklNnrStUuvZz] [-I length] [-i interval] [-M ttl] [-m minttl] [-O length] [-P proxy_username] [-p source_port] [-q seconds] [-s source] [-T keyword] [-V rtable] [-W recvlimit] [-w timeout] [-X proxy_protocol]
[-x proxy_address[:port]] [-Z peercertfile] [destination] [port]
```

#### 基础用法：建立TCP连接与传输数据

1. 指定是客户端还是服务端

使用 `-l` 参数来指定本次运行是作为客户端，还是作为服务端监听端口。

```shell
nc -l 1220 # 服务端 监听1220端口
nc 127.0.0.1 1220 # 连接 127.0.0.1:1220端口
```

建立链接之后，会进入一个交互的cli，链接的两端都可以自由的发送信息。

1. 传输数据

nc指令接受管道传输数据。

```shell
cat data_to_transfer.dat | nc host.to.transfer 1220 # 传输数据到 host.to.transfer 1220
printf "GET / HTTP/1.0\r\n\r\n" | nc blog.er1c.dev 80 # say hello to my blog
```

#### 端口扫描

手册已经很清晰了，这里贴一个手册上的例子。

```shell
nc -zv host.to.scan 20-30 # 扫描 host.to.scan 端口20-30
```

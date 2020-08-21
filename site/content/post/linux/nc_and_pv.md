---
title: "如何从线上机器拉取数据"
date: 2020-08-18T22:34:28+08:00
draft: false
tags:
    - linux
    - command
    - how-to
---

# 如何从线上机器拉取数据

最近遇到了一些线上问题，希望能够把线上抓到的信息拉到本地处理。
实现的方式有很多，比如利用python开启一个http server等。

但是我还是偏好使用更原生<del>装逼</del>的方法。

辗转中了解到`nc`指令以及可以用来限流的`pv`指令，两者结合，可以满足绝大部分网络拷贝的场景。

## how-to

*作为how-to系列的特色，开篇就会提供一个可用的方案。*

在线上机器执行，会监听1220端口，并把数据发送到该链接上。

```shell
cat data_file | nc -N -l 1220
```

在本地上执行，连接到线上机器，并把数据写入到local_data_file。

```shell
nc -d ip_or_hostname 1220 > local_data_file
```

### 1. 如何让双方在数据发送完成后都退出？

1. 发送方使用`-N`选项，表示如果读入`EOF`就关闭链接，这样就可以在文件全写入之后退出。
1. 接受方使用`-d`选项，表示忽略`stdin`，所以如果链接关闭，就会退出。

有其他实现方式的讨论，参见`参考`。

### 2. 如何控制流量，避免线上机器的网卡、带宽被打满？

这个问题可以转化为“如何控制管道中的数据传输速度？”，那么考虑使用`pv`指令。

*`pv`可能需要安装,如 `apt install pv`*

```shell
pv -L 1k data_file | nc -N -l 1220 # 限制传输速度1KB/s
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

## pv

指令用于管理流经管道的数据。可以展示消耗的时间、传输进度（进度条）、当前传输速率、传输的数据总量和预计完成时间(ETA estimated time of arrival).

### usage

```shell
# 1. 可以直接读文件
pv [option] file
# 2. 桥接管道
... | pv [option] | ...
```

### 选项

选项主要分为三个部分：展示、输出、数据控制。这里记录下数据控制部分的一些参数。

1. `-L`用来控制传输速率，byte/s为单位，接受`K` `M` `G` `T`来修饰。

# 参考

1. man手册
1. [BSD nc (netcat) does not terminate on EOF](https://serverfault.com/questions/783169/bsd-nc-netcat-does-not-terminate-on-eof)
1. [How to rate-limit a pipe under linux?](https://superuser.com/questions/239893/how-to-rate-limit-a-pipe-under-linux)

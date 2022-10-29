---
title: "RFC9113笔记-HTTP/2"
date: 2022-10-28T19:32:20+08:00
draft: true
tags:
    - http
    - rfc
    - memo
    - http-2-rfc
---

对于HTTP/2 RFC的笔记。

<!--more-->

HTTP/2支持更加高效的使用网络资源，
通过引入header field压缩和多路复用降低了延迟。

# 1 简介

旧版本的http有若干性能问题：

1. HTTP/1.0只允许同时发送一个请求；HTTP/1.1引入了pipeline机制，但仍然有head-of-line问题。
1. HTTP field有非常多的重复，这导致了不必要的流量和延迟。

HTTP/2在这两个方面作出了改进。

除此之外，HTTP/2还使用了二进制帧来传输，也有收益。

# 2 HTTP/2协议简介

HTTP/2是一个基于TCP面向链接的应用层协议。
client是TCP链接的发起方。

最基础的协议单元是帧。
每种帧用于不同的目的。

引入了流的概念来支持多路复用。
一个链接上可以有很多流，流之间非常的独立，不会互相影响。
为了更高效的利用多路复用，引入了流控和优先级（已经废除）。

frame中对一个链接中的HTTP的字段进行了压缩。

HTTP/2引入了服务端推送机制。

# 3 开始HTTP/2

首先需要探测服务器是否支持HTTP/2。

## 3.1 HTTP/2 版本标识符

- 基于TLS的时候使用"h2"。
- HTTP升级机制使用"h2c"。

## 3.2 https的HTTP/2

借助ALPN来实现。

## 3.3 通过预先设定来确认可以使用HTTP/2

## 3.4 HTTP/2 链接preface

每个终端都需要发送connection preface作为
使用HTTP/2的最终确认和
建立链接所需的初始化设置。

client和server使用不同的connection preface。

# 4 HTTP帧

当链接建立后，双方可以开始交换帧。

## 4.1 帧格式

所有的帧都有一个9-octet的头，然后是变长的帧payload。

```
HTTP Frame {
    Length (24),
    Type (8),

    Flags (8),

    Reserved (1),
    Stream Identifier (31),

    Frame Payload (..),
}
```

- **Length** 除了帧header之外的数据的长度，byte为单位。
- **Type** 帧的type。MUST丢弃不支持的帧。
- **Flags** 为特别的type保留的flag。
- **Reserved** 
- **Stream Identifier** 所属的流的标识符。

## 4.2 帧长度

有最大限制。

所有的实现MUST最少能够支持接受和处理2^14byte的帧。

## 4.3 Field Section 压缩和解压缩

Field Section压缩是指
将一组field line压缩成一个field block。

一个field block包含了一个单独的field section的所有field line。

field block包含了各种包的控制信息和header section。

### 4.3.1 压缩状态

# 5 流与多路复用

流是一个独立的、双向的、基于HTTP/2链接的、在client和server之间交换的帧序列。

- 一个链接可以有多个并发的流。
- 流可以被单方面的建立以及使用以及分享。
- 流可以被两侧关闭。
- 帧的发送顺序是敏感的。
- 流有一个唯一的整数来标识，在初始化时授予。

## 5.1 流状态

```
                                +--------+
                        send PP |        | recv PP
                       ,--------+  idle  +--------.
                      /         |        |         \
                     v          +--------+          v
              +----------+          |           +----------+
              |          |          | send H /  |          |
       ,------+ reserved |          | recv H    | reserved +------.
       |      | (local)  |          |           | (remote) |      |
       |      +---+------+          v           +------+---+      |
       |          |             +--------+             |          |
       |          |     recv ES |        | send ES     |          |
       |   send H |     ,-------+  open  +-------.     | recv H   |
       |          |    /        |        |        \    |          |
       |          v   v         +---+----+         v   v          |
       |      +----------+          |           +----------+      |
       |      |   half-  |          |           |   half-  |      |
       |      |  closed  |          | send R /  |  closed  |      |
       |      | (remote) |          | recv R    | (local)  |      |
       |      +----+-----+          |           +-----+----+      |
       |           |                |                 |           |
       |           | send ES /      |       recv ES / |           |
       |           |  send R /      v        send R / |           |
       |           |  recv R    +--------+   recv R   |           |
       | send R /  `----------->|        |<-----------'  send R / |
       | recv R                 | closed |               recv R   |
       `----------------------->|        |<-----------------------'
                                +--------+
```

**名词解释**

- `send` 发送一个帧
- `recv` 接受一个帧
- `H` HEADERS frame
- `ES` END_STREAM flag
- `R` RST_STREAM frame
- `PP` PUSH_PROMISE frame

特别的，上述的列出来的帧或者flag都是影响状态的。

# Reference

- [rfc9113](https://datatracker.ietf.org/doc/html/rfc9113)

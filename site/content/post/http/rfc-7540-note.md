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

## 3.1 

# Reference

- [rfc9113](https://datatracker.ietf.org/doc/html/rfc9113)

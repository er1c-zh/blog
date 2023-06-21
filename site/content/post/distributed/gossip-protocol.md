---
title: "Gossip Protocol"
date: 2023-06-20T11:38:35+08:00
draft: false
tags:
    - what
    - network
---

Gossip 协议是一种病毒式传播的P2P协议。
一些分布式系统用 p2p gossip 来保证数据分发到系统中的所有成员。

<!--more-->

# 交互的过程

类似流言传播的过程：“一传十，十传百。”

一轮传播中，拥有数据的成员选择一些其他成员（通常是随机选择），并将数据传递给他们。
按照一定的间隔，重复若干次传播，直至传播完成。

# 两种常见类型的 gossip 协议

## 用于传播

用于传播数据时需要考虑传播的时延问题。

## 用于统计数据

需要控制数据大小和存活时间。

也可以依赖节点间传播的顺序来解决一些问题。

# Reference

- [wiki Gossip Protocol](https://en.wikipedia.org/wiki/Gossip_protocol)

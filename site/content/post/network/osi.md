---
title: "OSI7层模型"
date: 2023-06-28T19:50:50+08:00
draft: true
tags:
    - network
    - what
---

**O**pen **S**ystem **I**nterconnection model，
开放式系统互联模型是国际标准化组织为
“协调系统互联类型标准研发提供一个公共的基础”而提出的理论模型。

<!--more-->

OSI分为七个抽象层，描述了网络交互从
比特在交互媒介中传输到数据在分布式应用的展示。
每一层都依赖下一层提供的能力来提供被上一层使用的能力。

Internet protocol suite (TCP/IP) 是一个OSI模型同时代的网络模型。
相比于OSI，TCP/IP模型更加的关注链接的上层、将物理层当作通用的物理链接来处理，有类似的分层结构但更加的粗泛。

# Layer 1: Physical layer

physcial layer负责在一个设备（网卡、以太网接口、网络交换机）和一个传输媒介间传递、接收非结构化的原生数据。

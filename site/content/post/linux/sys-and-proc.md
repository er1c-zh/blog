---
title: "/proc与/sys"
date: 2020-11-10T11:06:32+08:00
draft: true
tags:
    - linux
    - memo
    - what
---

> “任何东西都是文件。”

一如Linux的设计理念，内核将自身状态的展示和配置的展示修改与进程的状态信息和写操作也通过文件来实现。

`/proc`路径下放了两种信息：
- 以pid为名的进程信息
  - `/proc/{pid}/maps`
  - `/proc/{pid}/status`
  - `/proc/{pid}/stack`
  - `/proc/{pid}/io`
  - etc
- 以组件为名的组件信息
  - `/proc/cpuinfo`
  - `/proc/meminfo`
  - `/proc/net`
  - etc

`/sys`

# 


---
title: "go select原理分析"
date: 2021-05-14T11:17:03+08:00
draft: true
tags:
    - go-src
    - go
    - what
order: 2
---

从功能出发，分析select的原理。

<!--more-->

{{% serial_index go-src %}}

*go1.16*

*src/runtime/select.go*




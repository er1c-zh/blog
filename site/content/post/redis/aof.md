---
title: "redis AOF的实现"
date: 2021-03-02T23:39:28+08:00
draft: true
tags:
  - redis
  - how
---

上一篇学习了RDB相关的实现，接下来就轮到AOF了。

AOF是Append Only File的缩写，
很直观的说明了这个持久化的方法是通过追加写的方式来实现的。
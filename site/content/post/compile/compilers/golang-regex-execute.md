---
title: "golang的regex实现 执行匹配"
date: 2021-08-15T19:21:07+08:00
draft: true
tags:
    - compilers-book
    - go
    - how
order: 5
---

本文是golang的正则表达式的学习笔记。

主要介绍了在编译完成后，匹配的实现细节。

<!--more-->

*基于golang 1.16*

{{% serial_index compilers-book %}}

# 最开始的

在编译完成后，用户持有一个`regexp.Regexp`对象。

经过层层包装，
最后执行的函数是`doExecute`方法。
首先是检查是否可以一趟匹配来完成，
如果可以会调用`Regexp.doOnePass`来进行一趟匹配。
否则，会尝试利用`Regexp.backtrack`进行回溯匹配。
如果以上都不可以，最后，会尝试利用自动机形式进行匹配。

## 一趟匹配

编译好的正则表达式是否支持一趟匹配由`Regexp.onepass`为空来控制。

---
title: "eBPF总览"
date: 2023-03-30T19:01:52+08:00
draft: false
tags:
    - ebpf
    - linux
    - what
---

eBPF是一种Linux内核提供的，可以在操作系统内核这种需要权限的上下文中执行沙盒程序的技术。

<!--more-->

# Hook Overview

eBPF程序是事件驱动来执行的。
当系统运行到某个特定的节点时，注册的eBPF程序会被调用。

预定义的节点(hooks)包含系统调用，
函数进出、内核tracepoint，网络事件和其他的一些节点。

可以通过[kernel probe](https://docs.kernel.org/trace/kprobes.html)
或者user probe来将eBPF程序链接在任意地方而不限于预定义的节点。

# 如何开发eBPF程序

通常情况下，可以通过 **Cilium** , **bcc** 或者 **bpftrace** 等工具来开发。

如果这种高层抽象无法满足需要，
可以通过LLVM这种编译套件将c源码编译成eBPF需要的字节码(bytecode)。

# 加载与验证

eBPF程序可以通过系统调用加载到内核中，
通常这个过程由eBPF库来完成。

![eBPF-loader-and-verification-architecture](https://ebpf.io/static/1a1bb6f1e64b1ad5597f57dc17cf1350/6515f/go.png)

加载过程主要分为两步：

1. **验证**

    首先会验证eBPF程序是否安全：

    - 是否有权限加载。
    - 程序不会crash或者影响整个系统。
    - 程序会执行完成，而不是无限循环或者阻止系统继续运行。

1. **JIT 编译**

    将字节码编译成机器指令，使eBPF程序能够以编译过的程序的性能运行。

# Maps

eBPF maps提供存取各种结构的数据的能力。
其中的数据可以从多种地方读取，比如通过用户空间发起系统调用。

# Helper Calls

eBPF程序不能调用内核函数。

因为这个会导致程序与特定的内核版本绑定，
也会增加程序的复杂度。

作为替代，可以调用helper functions，
内核提供的稳定、众所周知的API。

# Tail & Function Calls

eBPF程序通过尾调用和函数调用组织在一起。
函数调用允许在eBPF程序中定义并调用函数。
尾调用可以调用另一个eBPF程序并且替换执行的上下文，
类似`execve`在普通进程中的效果。

# eBPF Safety

> With greate power there must also come great responsibility.

eBPF程序的安全性通过如下几个方法来实现：

## Verifier

eBPF验证器保证程序本身的安全性：

- 会完成而不是无限循环或者阻塞。
- 不会使用未初始化的变量或者越界读取。
- 程序不会过大。
- 程序的复杂度可控。验证器会生成所有的执行路径，所以过于复杂的程序会导致分析失败。

## Hardening

> However, in the context of technology or computer security, hardening specifically refers to the process of making a system or application more secure and resistant to potential cyberattacks.

*From ChatGPT*

在验证后，eBPF程序会根据从授权或未授权的进程中加载来进行加固。

- 执行保护：eBPF程序所在的内存只读。 
- Mitigation against spectre
- constant blinding

## 抽象运行时

程序不能直接访问任意内存，只能通过helper来访问进程上下文之外的内存。

# Development Toolchains

bcc/bpftrace/eBPF Go Library/libbpf C/C++ Library

# Reference

- [What is eBPF](https://ebpf.io/what-is-ebpf/)

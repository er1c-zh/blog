---
title: "ruby入门笔记"
date: 2021-04-27T17:48:52+08:00
draft: true
tag:
    - what
    - how
    - ruby
---

# 最开始的

**面向对象**


- 大小写敏感
- **statement**是一条**指令**_(instruction)_或一组指令，接受换行符和分号作为切分
- `#`作为行内注释的开头，`=begin`和`=end`标记多行注释
- **代码块** _(code block)_ 是若干行代码的组合。有两种标记方式：
    - 用大括号包裹
    - 用任意方式标记开始，由`end`关键字标记结束
- **标记符** _(identifier)_ 是变量､类､函数等的名称，有如下要求：
    - 数字､字母_(alphanumeric character)_､下划线组成
    - 非数字开头 
- ruby自然有自己的**保留字**，这里不再赘述

# 变量 _variable_

> 变量是一个用于存储数据的内存区域的名称
>
> A variable is a name given to a memory location which is used to store data. 

对于**基本类型**变量存储了值，对于其他类型，变量存储了引用。


# 参考

- [ruby-new](http:/saito.im/slide/ruby-new.html)


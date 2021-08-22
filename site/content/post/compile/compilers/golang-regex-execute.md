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

考虑到最后的匹配更为通用，
首先分析的是普通的匹配，
然后分析一趟匹配与回溯匹配。

## 普通匹配

如果以上的两种特别情况都不满足，
就会进行普通匹配。

实现的思路是构建一个自动机来执行编译好的匹配程序。
自动机维护若干“进程”，每次从输入消费一个字符，迭代进程，
并创建从当前位置开始匹配的一个“进程”。

首先通过`Regexp.get`来获得一个自动机`machine`，
随后初始化。

通过`machine.match`来执行匹配。

### `machine.match`

首先分析运行的状态、环境与变量。
自动机拥有两个队列，
用来交换作为当前处理的队列`runq`和下次要执行的队列`nextq`。
`pos`表示下一个处理的字符的下标。
`width`表示本轮处理的字符的宽度。
`r`表示本轮处理的字符。

`match`方法逻辑比较清晰，
首先初始化了若干变量，
执行一些fast-path的处理，如需要从开头匹配，但`pos`已经不再开头这种情况。

随后在循环中分别调用
`machine.add`创建从`pos`开始匹配的“进程”
和`machine.step`来推进队列中的进程处理当前字符`r`。

检查如果匹配到了输入的结尾，跳出循环。
否则迭代入下一轮处理，更新下标、待处理的字符`r`、宽度`width`，以及交换两个队列。

最后返回`machine.matched`作为匹配的结果。

### `machine.add`

`add`会尝试向传入的任务队列中增加一组在不额外匹配字符的情况下可以执行到的指令。

`func (m *machine) add(q *queue, pc uint32, pos int, cap []int, cond *lazyFlag, t *thread) *thread`

函数接受很多参数：

- 要加入的队列`q`。
- 指令下标`pc`。
- 当前匹配到的位置`pos`，用于递归增加进程时检查到捕获组等情况来标记位置。
- `cap`
- `cond`
- `t` 如果有已经创建好的`thread`对象，会传入来复用。

返回值的`thread`对象如果不为nil，表示没有被复用，
比如在`step`调用`add`时，如果`thread`没有被复用，会在该轮执行结束后被回收。

首先检查任务队列中是否已经拥有了指向当前指令的任务，
如果有，直接返回。

然后向队列中增加一个新建的、表示任务的`entry`对象。
根据指令的类型：

- 如果遇到了`InstAlt`/`InstEmptyWidth`等不需要额外消费字符的指令，
更新指令下标`pc`重复这个过程，直到全部遇到了会消费字符的指令。

- 如果遇到了如`InstRune`等会匹配字符的指令，
需要时（比如还未新建一个`thread`）会新建一个`thread`对象，
设置到`entry`中。

在一个任务队列`regexp.queue`对象中，
可以通过`queue.dense[queue.sparse[pc]].t`
来查找到当前执行到pc指向的指令的`thread`对象。

### `machine.step`

`step`遍历`runq`队列中的所有任务，
执行匹配，并将可能的任务放入到下一轮的任务队列`nextq`中。

`func (m *machine) step(runq, nextq *queue, pos, nextPos int, c rune, nextCond *lazyFlag)`

- 当前任务队列`runq`和下一轮任务队列`nextq`。
- 本轮处理的字符`c`，在输入中的下标`pos`以及下一轮处理的字符的下标`nextPos`。
- `nextCond`

`step`遍历`runq.dense`中的每个`entry`。
根据该任务的指令的类型，
利用不同的匹配方式，
如`InstRune`会调用`syntax.Inst.MatchRune`而
`InstRune1`会直接`c == i.Rune[0]`比较，
检查该任务是否能匹配该轮要处理的字符，
如果可以，那么调用`machine.add`来将下一轮需要执行的指令推入`nextq`中。

如果遇到了`InstMatch`指令，
表示匹配完成，
存储匹配的结果，
并将`machine.matched`改为`true`。

## 一趟匹配

编译好的正则表达式是否支持一趟匹配由`Regexp.onepass`为空来控制。

编译过程上一篇文章已经简单分析了，
匹配过程由`Regexp.doOnePass`来完成。

根据上篇文章的分析，
我们知道重新改写之后的`Regexp.onepass`中可以进行较为 **“固定的”** 的匹配，
所以实现上比较简单。

`Regexp.doOnePass`核心逻辑由一个循环来实现，
因为“一趟”的特性，
所以可以简单的顺序的匹配即可。

与普通匹配相比，匹配字符、捕获、匹配完成等十分类似，
区别主要是`InstAlt`的分支指令的匹配。

`pc = int(onePassNext(&inst, r))`会根据`lookahead`的`r`来判断下一步要执行的指令。

根据生成的表示要匹配的备选字符参数结构，
首先会进行一些快速检测，
随后二分查找是否匹配与对应的下一步指令。

## 回溯匹配

todo

# 参考

- [Regular Expression Matching: the Virtual Machine Approach
](https://swtch.com/~rsc/regexp/regexp2.html)
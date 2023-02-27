---
title: "[吃太饱]diff3"
date: 2021-04-16T17:08:35+08:00
draft: false
tags:
    - eat-too-much
    - how
---

最近在尝试看一些理论的材料然后落地其中的设计。

这种事情属实费力（因为能力太差）还没有收益（已经有完美的落地方案了），
所以打算开启一个新的系列——《吃太多》来记录学习经过。

通过这一篇可以了解到：

- diff3是什么
- 提供的保证和一些性质
- 一个简单的diff3实现

<!--more-->

{{% serial_index eat-too-much %}}

# 是什么与三向合并

diff3是一个三向合并的工具，用来合并两个基于同一个文件的修改而来的文本文件。

## 背景

实际工作中经常有多个人同时修改一个文本文件，随后合并到一起的场景，
用来解决这种问题有两种思路：基于操作的合并和基于状态的合并。

基于操作的合并通过追踪全部的修改操作，然后尝试构造出一个统一的视图。

相反地，基于状态的合并，只需要关注文本的当前状态，然后得出一个统一的状态。
一个关键的难点是如何将不同版本之间的数据“对齐(align)”。
如果数据是结构化的、有key的，“对齐”的规则（用key来对齐）是清晰的；
而对于更富变化性的数据，比如文本，或者抽象的讲，列表的内容，
因为没有天然的对齐锚点，进行“对齐”更加困难。

# diff3

`diff3`是一个基于状态的合并工具，
支持输入两个需要合并的版本和一个旧版本来输出合并的结果。

## 定义

首先给出原子对象$\mathcal{A}$的集合。
从实现上来看，原子对象是合并操作的不可分的对象，通常是文本的一行。

有$\mathcal{A}^\*$来表示$\mathcal{A}$中的元素组成的列表的集合，
用$J,K,L,O,A,B,C$来表示$\mathcal{A}^\*$中的元素。
下文称之为**序列**。

对于列表$J$，有$J[k]$表示$J$中的第$k$个元素，
有$J[i..j]$表示$J$中的一个**片段(_span_)**，是闭区间。

定义 **配置(_configuration_)** 是
三元组$(A,O,B)\in\mathcal{A}^\* \times \mathcal{A}^\* \times \mathcal{A}^\*$。
特别的，$\mathcal{A},\mathcal{B}$是从$\mathcal{O}$派生出来的。

定义 **同步器(_synchronizer_)** 是一个函数，
接受 一个**配置**，输出另一个**配置**。
**同步器**的一次 **运行(_run_)** 表示为$(A,O,B)\rightarrow(A',O',B')$。
特别的，如果输出的三个相等，那么称这次 **运行** 为 **无冲突的(_conflict-free_)**，
表示为 $(A,O,B)\rightarrow C$ 。

**不交叉匹配**是一个布尔函数，接受两个参数分别是两个序列的下标，
如果有$M_A(i,j)=true$，那么：

1. $O[i]=A[j]$
1. 对于任意的$i' \neq i$ 和 $j' \neq j$ ，有 $M_A[i',j] = M_A[i,j'] = false$
1. 如果有 $i' > i  \land j' < j$ 或 $i' < i  \land j' > j$，则$M_A[i',j'] = false$

直觉的看，如果两个序列有两个元素匹配，那么两个元素唯一匹配且两个元素的前后不会匹配。
这样可以简单的通过递归实现。

定义 **chunk** 是分别来自$\mathcal{A},\mathcal{O},\mathcal{B}$的三个**片段**的三元组，
三个**片段**至少有一个不为空，
表示为$H=([a_i..a_j],[o_i..o_j],[b_i..b_j])$。
定义表示$A[H]=[a_i..a_j]$。
**chunk**的**长度(_size_)**是三个**片段**之和。
特别的，**稳定chunk(_stable chunk_)**指的一个**chunk**的三个片段长度相同且满足**不交叉匹配**；
反之，如果有不匹配的片段，则称为**不稳定chunk(_unstable chunk_)**。
针对**不稳定chunk**，有如下分类：

- $H$在$A$中**改变(_changed_)** $O[H]=B[H]$但不与$A[H]$相同
- $H$在$B$中**改变(_changed_)** $O[H]=A[H]$但不与$B[H]$相同
- $H$是**假冲突的(_falsely conflicting_)** $A[H]$与$B[H]$相同但与$O[H]$不同
- $H$是**真冲突的(_(truly) conflicting_)** $A[H],B[H],O[H]$三者不同

## 根据直觉的描述

diff3接受三个输入，输出一个由稳定chunk和不稳定chunk组成的序列。
这个序列满足：

1. 任意满足$M_A[o,a]=M_B[o,b]=true$的下标$a,o,b$，都会出现在相同的稳定chunk中。
1. 每个稳定chunk都是尽量大的。

下面简单给出diff3的具体运行流程。

diff3首先调用两路比较来比较$(O,A)$和$(O,B)$来生成两个 **不交叉匹配(_non-crossing matching_)** ：$M_A,M_B$。

特别的，在diff3的实现中，认为生成**不交叉匹配**的**两路比较**函数是一个黑盒，具有稳定性和最大匹配性。

然后执行如下流程：

1. 初始化三个下标$l_O=l_A=l_B=0$
2. 查找最小的正数$i$使$M_A[l_O + i, l_A + i]=false$或$M_B[l_O+i, l_B+i]=false$成立，如果不存在，进入最后一步，输出一个**稳定chunk**

    - 如果$i=1$，那么查找最小的整数$o>l_O$，使得存在下标$a,b$令$M_A[o,a] = M_B[o,b] = true$成立。
        如果不存在进入最后一步，输出一个**不稳定chunk**。
        如果存在输出不稳定chunk：

        $$
        C=([l_A + 1 .. a-1],[l_O+1..o-1],[l_B+1..b-1])
        $$
    
        然后令$l_A=a-1, l_O=o-1, l_B=b-1$，重复步骤2
    
    - 如果$i>1$，那么输出**稳定chunk**

        $$
        C=([l_A + 1 .. l_A + i - 1], [l_O + 1 .. l_O + i - 1], [l_B + 1 .. l_B + i - 1])
        $$

        然后令$l_A = l_A + i -1, l_O = l_O + i - 1, l_B = l_B + i - 1$，重复步骤2

3. 如果$A,O,B$有任意剩余的数据，输出最终的chunk

    $$
    C=([l_A + 1 .. |A|], [l_O + 1 .. |O|], [l_B + 1 .. |B|])
    $$

简单的来说，在获得两个匹配之后，
算法通过遍历三个序列查找稳定的chunk或不稳定的chunk并输出，直到结束。

## 证明

对于结果满足的性质1，
可以看到每个不稳定chunk开始于有至少一个不匹配的下标组$a,b,o$，
随后递增$o$，直到找到第一个$o'$满足$M_A[o',a']=M_B[o',b']=true$。
显然，能够得出任意不稳定chunk中都不存在下标组$a,b,o$使得$M_A[o',a']=M_B[o',b']=true$成立。

对于性质2，
可以通过反证法简单的来证明。
稳定chunk的产生有三种，分别是起始的时候、中间的时候和最后一步。
对于任意一种情况，如果要产生“不是最大”的稳定chunk，
就需要在开始或结束的时候有满足$M_A[o',a']=M_B[o',b']=true$的元素被包含在不稳定chunk中
或在开始结束的时候错误的中断。
对于第一种情况，由性质1可以得出是不可能的。
对于第二种情况，第一个稳定chunk的结束是遇到$M_A[o',a'] \neq true \lor M_B[o',b'] \neq true$，
而最后一个稳定chunk的结束是三个序列的结尾，所以也是不可能的，
故性质2也是成立的。

# diff3的一些性质

### 一个能输出唯一无冲突结果的充分条件

首先给出结论：

> 对于一个**配置** $A,O,B$，如果满足：
> 1. 对于$A,O,B$三个序列拆分为$A=A_1A_2A_3,O=O_1O_2O_3,B=B_1B_2B_3$，
> $A,B$的修改分别仅发生在$A_1,B_3$或$A_3,B_1$中
>
> 1. 存在一个元素 $x \in \mathcal{A}$ 在$A,O,B$各序列中出现且仅出现一次
>
> 则$(A \leftarrow O \rightarrow B)$能产生唯一无冲突结果。

在一些实际的使用环境中，比如代码合并，通常情况下能够找到唯一的一行来满足条件2，
进而在满足情况1的情况下，能够得出唯一的无冲突合并结果。

结论比较符合直觉，
证明过程概括来说，
首先给出了引理：

> 如果存在一个唯一的元素$x$，那么它一定在一个**稳定chunk**中

然后通过反证法的思路，证明了结论，具体的证明这里就不给出了。

### 对于同一个输入，可能有多个输出

简单来说，造成这个现象的原因是因为最长匹配并不是唯一的。

举个例子：

- A: 1 2 4 6 8
- O: 1 2 3 4 5 5 5 6 7 8
- B: 1 4 5 5 5 6 2 3 4 8

有两种可能的最长匹配：

1. 第一种情况

    ```plain
    1 2   4       6       8
    1 2,3 4 5,5,5 6   7   8
    1     4 5,5,5 6 2,3,4 8
    ```

2. 另一种情况

    ```plain
    1     2   4 6   8
    1     2 3 4 6,7 8
    1 4,6 2 3 4     8
    ```

# 一个简单实现

diff3从**Match**生成一系列Chunk的过程已经非常的清晰了。

那实现一个diff3只差从两个序列中生成**Match**。

根据算法的描述，**Match**需要满足不交叉匹配、稳定性与最大匹配的特性。

[不交叉匹配与最大匹配通过递归来实现](https://github.com/er1c-zh/diff/blob/master/diff3/diff3.go#L152)：
1. 首先通过类似[LCS](https://github.com/er1c-zh/diff/blob/master/diff3/diff3.go#L233)的思路来查找两个序列的最长子序列
    1. 如果有多个长度相同的最长子序列，固定取第一个（或最后一个），这样就能满足稳定性
1. 这时两个序列被分为三个部分，分别是匹配的部分、匹配的部分之前的部分和匹配的部分之后的部分。
这时分别在前、后递归的调用该过程，直到匹配结束。这样就能保证不会出现交叉匹配的情况。

如此，我们就能获得一个满足需要的匹配，
只需要按照diff3[生成chunk的过程](https://github.com/er1c-zh/diff/blob/master/diff3/diff3.go#L20)依次执行，
就能完成一个简单的diff3实现.

上述的代码可以在我的[repo](https://github.com/er1c-zh/diff)查看。

![re-impl-diff3.png](https://dsm01pap001files.storage.live.com/y4msYJ9CAQiTK9p37pcfRsVSZtboncVc1zA2MCXNUe7TMCH3CS5OJCuLgmr6AU30x7GxbaW0xUGyfMUjUXV6ep1ErOcPQWAyBT2Anyz530PyCVzzMHRC2E0ybdWhYoy4EvSIj12b9w5hozZswTniSvWv8lprUxSdprqIMpRsZ9Q4gXtjyNpgv05nCEAPmx9JvT1?width=2198&height=1026&cropmode=none)

# 参考

- [diff3 wiki](https://en.wikipedia.org/wiki/Diff3)
- [A formal investigation of Diff3](http://www.cis.upenn.edu/~bcpierce/papers/diff3-short.pdf)

---
title: "正则表达式速查"
date: 2020-08-21T20:31:59+08:00
draft: false
tags:
    - memo
    - regex
---

# 基础

## 元字符

1. `\b` 匹配单词分割处
1. `.` 匹配非换行符所有字符
1. `\w` 匹配文字字符
1. `\s` 空白字符
1. `\d` 数字
1. `^` 开头
1. `$` 结尾

## 转义

使用 `\` 进行转义

## 重复 

均指重复之前一个组或字符类或字符或元字符

1. `{n}` 重复n次
1. `{n,}` 重复n次或更多
1. `{n,m}` 重复n次到m次，闭区间
1. `*` 等价于 `{0,}`
1. `+` 等价于 `{1,}`
1. `?` 等价于 `{0,1}`

## 字符类

表明一个字符集合

1. `[a-z]` 从a到z
1. `[a-zA-Z]` 从a到z和从A到Z
1. `[0-9]` 从0到9
1. `[ace]` a或c或e

## 分支条件

等价于或

1. `110|120` 匹配110或者120
1. `1(1|2)0` 匹配110或者120

## 反义

1. 简单的反义
    1. `\W` `\S` `\D` `\B` 与相应的小写元字符相反
    1. `[^abc]` 不是abc的字符
1. 用负向零宽断言实现反义
    1. `^(?:(?!abc).)*$` 匹配不含abc的字符串

## 分组

用小括号进行分组，构成 *正则表达式匹配单元* ，可以用来作为重复、分支条件、后向引用的单元。

1. `(abc)+` 可以匹配如 `abc` ， `abcabcabc` 等字符串
1. `1(1|2)0` 可以匹配 `110` 和 `120`
1. `(\d*) \1` 可以匹配 `110 110` `120 120` `12341234 12341234` 等

## 后向引用

引用前面捕获的字符串

1. `(a regex) \1` 捕获并自动命名为1
1. `(?<name> aRegex) \k<name>` 捕获并命名为name
1. `(?:aRegex)` 不捕获，不命名，用于重复或分支条件等，即不用于后向引用

## 零宽断言与负向零宽断言

用于匹配位置，可以理解为位置（比如匹配不包含某个字符串的字符串）或是否存在（用于判断不是某些字符串结尾的 `windows(?!(98|xp|XP)` ）

1. `(?=aRegex)` 匹配aRegex前面的位置
1. `(?<=aRegex)` 匹配aRegex后面的位置
1. `(?!aRegex)` 匹配后面跟的不是aRegex的位置
1. `(?<!aRegex)` 匹配前面不是aRegex的位置

## 注释

## 贪婪与懒惰

# 工具

## 在线测试网站

1. 网站支持中文。
1. 有对正则结构的解析。

[regex101.com](https://regex101.com)

## grep

```shell
# 简单的正则处理
grep 'a regex'

# 加强正则
grep -E '1(1|2)0'

# perl
grep -P '^(?:(?!abc).)*$'
```

# How-to

## 如何匹配不包含某个字符串的行？

```shell
grep -P "^(?:(?!abc).)*$" ./file_to_check
```


![regex-do-not-have-a-string](https://dwbjpa.dm.files.1drv.com/y4mhcUtoiMm1qQzxKK-ICdtDGJkTJSdyEgQHS5IRT74YTK2Oeo-sxGw7kcs-fjpodTkIftW4Fuleq8Drc5DEsHYHm-FFfoj_Qy4MT8nt6-zF_faoNo_z-JAskH3zCUZcVp3ugjjO1L4ADyiTx7AaUtSR-U9Z9QoyDidr-rC_tZF2AWEfs718ByXv9K-w2N6H5t0YeG0DloWV7IL1D0A8p_ONQ?width=352&height=64&cropmode=none)

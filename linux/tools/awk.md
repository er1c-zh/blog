---
title: awk简介
---

[TOC]

## 写在前面的

- `awk` 是一门语言
- `awk` 也是一个程序
- 基于 `mawk 1.3.3` 编写 _这个程序和我一样大了 笑_

## SYNOPSIS

```shell
mawk [-W option] [-F value] [-v var=value] [--] 'program text' [file ...]
mawk [-W option] [-F value] [-v var=value] [-f program-file] [--] [file ...]
```

## Description

> An AWK program is a sequence of pattern {action} pairs and function definitions.  Short programs are entered on the command line usually enclosed in ' ' to avoid shell interpretation. Longer programs  can be read in from a file with the -f option.  Data input is read from the list of files on the command line or from standard input when the list is empty.  The input is broken into records as determined by the record separator variable, RS.  Initially, RS = "\n" and records are synonymous with lines.  Each record is compared against each pattern and if it matches, the program text for {action} is executed.

一个 `awk`  程序是一个 `pattern {action}` 和 `函数` 的序列. 较短的程序可以直接在命令行中输入, 使用单引号包裹避免shell误读; 相对的, 较长的程序可以写入文件并被执行器读取并执行. 被处理的数据是从 `文件列表 [file]` 或者 `标准输入(当文件列表是空时)`. 输入将被 `record separator variable RS` 分割. 默认的, `RS` 为 \n. 每条记录将被与 `pattern` 比较, 如果匹配成功, `{action}` 将被执行.

## Options

以下四个选项是 POSIX 规定的接口

- `-F value` 设置 `field separator`
- `-f file` 程序所在的文件. 允许多个文件同时存在
- `-v var=value` 将 变量 `var` 赋值为 `value`
- `--` 表明一个选项的结束

以下是是 `mawk` 提供的额外选项

略

## The AWK Language

### 1. Program structure

> An AWK program is a sequence of pattern {action} pairs and user function definitions

#### pattern

一个 `模式 pattern` 可以是

- __BEGIN__
- __END__
- expression
- expression , expression

#### pattern {action}

`pattern` 和 `action` 中可以省略一个, 但至少需要有一个

- 如果 `action` 省略, 默认为 ```{ print }```

- 如果 `pattern` 省略, 默认为总是匹配成功
- 对于 `BEGIN` 和 `END` , 总是需要 `action`

#### statement

- `statement` 由 `一个pattern {action}` 组成

- `statement` 由换行或分号表示终结
- 多个 `statement` 由大括号包裹成为一个 `block` , `block` 中的最后一个 `statement` 不需要终结符
- 空白 `statement` 由一个分号终结
- 较长的 `statement` 可以由 `\` 进行接续换行
- 一个 `statement` 可以在 `,` `{` `&&` `||` `do` `else` , `if` `while` `for` 的条件之后, `函数定义` 之后进行无 `\`的换行
- 注释由 `#` 开头

##### 具有流控制效果的 `statement`

```awk
if ( expr ) statement
if ( expr ) statement else statement
while ( expr ) statement
do statement while ( expr )
for ( opt_expr ; opt_expr ; opt_expr ) statement
for ( var in array ) statement
continue
break
```

### 2. Data types, conversion and comparison

### 两种基本数据类型

#### numeric

- 分为 `integer(-2)` `decimal(1.08)` `scientific notation(-1.1e4 or .28E-3)`

- > all numbers are represented internally

- 所有的计算都是浮点运算

#### string

- `string` 使用引号 _（区别于单引号）_ 封闭
- 使用 `\` 换行
- 可以使用转义符

#### 特别的，新定义的变量自动初始化的值被认为同时具有0和空字符串的含义

### 类型转换

> The type of an expression is determined by its context and automatic type conversion occurs if needed.

**表达式的类型由上下文决定, 并且在必要的时候进行类型转换**

### 3. Regualr expressions

> Regular expressions are enclosed in slashes.

```awk
expr ~ /r/ # 当 expr match r 的时候, 表达式等于1
/r/ { action } # 当输入 match r 时, 执行 action
$0 ~ /r/ { action } # 与上一个 statement 相同
```

- `AWK` 使用的 `egrep` 作为正则的规则

### 4. Records and fields

## Ref

- mawk manual Version 1.2
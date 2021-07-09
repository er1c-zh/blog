---
title: "golang的regex实现"
date: 2021-07-07T23:05:42+08:00
draft: true
tags:
    - compilers-book
    - go
    - how
order: 4
---

本文是golang的正则表达式的学习笔记。

<!--more-->

{{% serial_index compilers-book %}}

# 最开始的

`regexp`包实现了语法与[`RE2`](https://github.com/google/re2/)相同的正则表达式匹配功能，
能够保证在线性时间中完成匹配。

```go
reg, err := regexp.Compile(`regex`)
if err != nil {
    panic("invalid regex.")
}
found := reg.MatchString("hello regex.")
```

这是golang使用`regexp`库的一个典型模式：

1. 首先`Compile`要使用的正则表达式，得到一个`regexp.Regexp`实例。
1. 利用`regexp.Regexp`进行匹配。

本文首先分析记录关联的数据结构，然后分析编译的过程，最后分析匹配实现。

# 数据结构

## `regexp.Regexp`

`Regexp`表示了编译过的正则表达式，
大部分函数并发安全。

`Regexp`的结构如下：

```go
type Regexp struct {
    expr           string       // as passed to Compile
    prog           *syntax.Prog // compiled program
    onepass        *onePassProg // onepass program or nil
    numSubexp      int
    maxBitStateLen int
    subexpNames    []string
    prefix         string         // required prefix in unanchored matches
    prefixBytes    []byte         // prefix, as a []byte
    prefixRune     rune           // first rune in prefix
    prefixEnd      uint32         // pc for last rune in prefix
    mpool          int            // pool for machines
    matchcap       int            // size of recorded match lengths
    prefixComplete bool           // prefix is the entire regexp
    cond           syntax.EmptyOp // empty-width conditions required at start of match
    minInputLen    int            // minimum length of the input in bytes

    // This field can be modified by the Longest method,
    // but it is otherwise read-only.
    longest bool // whether regexp prefers leftmost-longest match
}
```

## `syntax.Regexp`

`syntax.Regexp`表示正则表达式的语法树的一个节点。

```go
type Regexp struct {
    Op       Op // operator 节点类型
    Flags    Flags
    Sub      []*Regexp  // subexpressions, if any 子节点，用于并、连接等操作
    Sub0     [1]*Regexp // storage for short Sub
    Rune     []rune     // matched runes, for OpLiteral, OpCharClass
    Rune0    [2]rune    // storage for short Rune
    Min, Max int        // min, max for OpRepeat
    Cap      int        // capturing index, for OpCapture
    Name     string     // capturing name, for OpCapture
}
```

### 节点类型

```go
const (
    OpNoMatch        Op = 1 + iota // matches no strings
    OpEmptyMatch                   // matches empty string
    OpLiteral                      // 匹配字符序列
    OpCharClass                    // matches Runes interpreted as range pair list
    OpAnyCharNotNL                 // 除换行符之外所有的字符
    OpAnyChar                      // 所有字符
    OpBeginLine                    // 从一行开头开始匹配
    OpEndLine                      // 匹配一行的结尾
    OpBeginText                    // matches empty string at beginning of text
    OpEndText                      // matches empty string at end of text
    OpWordBoundary                 // matches word boundary `\b`
    OpNoWordBoundary               // matches word non-boundary `\B`
    OpCapture                      // capturing subexpression with index Cap, optional name Name
    OpStar                         // matches Sub[0] zero or more times
    OpPlus                         // matches Sub[0] one or more times
    OpQuest                        // matches Sub[0] zero or one times
    OpRepeat                       // matches Sub[0] at least Min times, at most Max (Max == -1 is no limit)
    OpConcat                       // matches concatenation of Subs 匹配Subs的连接
    OpAlternate                    // matches alternation of Subs 匹配Subs的并集
)
```

# 编译正则表达式

编译的工作最终由`func compile(expr string, mode syntax.Flags, longest bool) (*Regexp, error)`完成。

函数接受正则表达式字符串`expr`，正则表达式的格式`mode`和匹配模式`longest`。

`compile`首先
1. 通过`syntax.Parse`将正则表达式解析为正则表达式语法树(syntax tree)，
返回表示语法树的对象`syntax.Regexp`，
1. 随后调用`syntax.Regexp.Simplify`简化结构，
1. 然后根据生成的语法树生成执行匹配需要的代码，由`syntax.Compile`完成，
1. 最后构建`regexp.Regexp`对象返回。

## 构建语法树

`syntax.Parse`将一个正则表达式解析为语法树。

理论上来看，Parse没有使用递归下降算法，
这样可以避免指数级别增长的递归深度以及潜在的栈溢出问题。

实现上依赖`syntax.parser`，故首先分析下该对象。

### `syntax.parser`

```go
type parser struct {
    flags       Flags     // parse mode flags
    stack       []*Regexp // stack of parsed expressions
    free        *Regexp
    numCap      int // number of capturing groups seen
    wholeRegexp string
    tmpClass    []rune // temporary char class work space
}
```

todo 分析每个字段的用途

- `func (p *parser) newRegexp(op Op) *Regexp`

    产生一个新的节点对象，

    

- `func (p *parser) op(op Op) *Regexp`

    新建一个语法树节点，

### 解析正则表达式

在解析开始的时候，初始化一个`parser`，存储了解析需要的一些状态。

函数主体是一个大的循环，
循环体是一个`switch`
检查`lookahead`符号来调用相应的处理函数读取、处理相应的单元。

- 左括号 `'('`

    1. 左括号标记了一个捕获组 *(capture group)* 的开始，
    故增加计数器`parser.numCap`。


# 参考

- [Regular Expression Matching Can Be Simple And Fast](https://swtch.com/~rsc/regexp/regexp1.html)
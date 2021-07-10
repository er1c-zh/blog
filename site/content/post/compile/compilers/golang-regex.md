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

*基于golang 1.16*

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

### Flags

```go
const (
    FoldCase      Flags = 1 << iota // case-insensitive match
    Literal                         // treat pattern as literal string
    ClassNL                         // allow character classes like [^a-z] and [[:space:]] to match newline
    DotNL                           // allow . to match newline
    OneLine                         // treat ^ and $ as only matching at beginning and end of text
    NonGreedy                       // make repetition operators default to non-greedy
    PerlX                           // allow Perl extensions
    UnicodeGroups                   // allow \p{Han}, \P{Han} for Unicode group and negation
    WasDollar                       // regexp OpEndText was $, not \z
    Simple                          // regexp contains no counted repetition

    MatchNL = ClassNL | DotNL

    Perl        = ClassNL | OneLine | PerlX | UnicodeGroups // as close to Perl as possible
    POSIX Flags = 0                                         // POSIX syntax
)
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

`compile`
1. 首先通过`syntax.Parse`将正则表达式解析为正则表达式语法树(syntax tree)，
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
    numCap      int // 捕获组的数量
    wholeRegexp string
    tmpClass    []rune // temporary char class work space
}
```

#### 字段

todo 分析每个字段的用途

- `free`

    `free`字段用于存储已经分配但没有使用的`Regexp`的实例，
    存储的结构类似一个链表，`Regexp.Sub0[0]`是next节点。

    `free`字段只有`func (p *parser) reuse(re *Regexp)`方法会修改，
    当函数不再需要一个节点实例时，调用`reuse`，保存到`free`链表上。

    在`newRegexp`会在新分配一个实例时，尝试从`free`链表中获取，
    如果为空，再新建一个。

#### 方法

- `func (p *parser) newRegexp(op Op) *Regexp`

    产生一个新的节点对象，首先从`p.free`中尝试获取，如果失败，就`new`一个。
    详见`parser.free`的字段的解释。

- `func (p *parser) op(op Op) *Regexp`

    `p.newRegexp`新建一个语法树节点，`p.push`到堆栈中，返回这个节点。

- `func (p *parser) push(re *Regexp) *Regexp`

    将一个节点推入`parser`的堆栈中。

    1. 根据节点的类型做一些工作。

        - 普通类型

            `p.maybeConcat(-1, 0)` 尝试合并栈上最近的两个节点。

        - `OpCharClass`

            todo
    
    1. `p.stack = append(p.stack, re)`

- `func (p *parser) maybeConcat(r rune, flags Flags) bool`

    检查并在可能的情况下，合并栈上的两个`OpLiteral`节点，返回`r`是否被推入栈中。

    1. 从栈上获取最上面的两个节点（栈顶`re1`和栈顶下一个`re2`），
    检查是否是`OpLiteral`节点且两者的大小写匹配是否一致。
    符合则继续，否则返回`false`（表示`r`未入栈）退出。
    1. 如果符合条件，将`re1.Rune`推到`re2.Rune`中。
    1. 如果入参传递了`r`（具体的，通过`r >= 0`来判断），则复用`re1`，返回`true`表示已经将`r`入栈。
    1. 因为无法复用，现在pop出`re1`。
    1. 调用`p.reuse(re1)`，等待后续的复用。
    1. 返回`false`。

- `func (p *parser) literal(r rune)`

    1. 新建一个`OpLiteral`节点。
    1. 如果是大小写不敏感的匹配，调用`minFoldRune`将字符化归到最小编码。
    1. 将`r`保存到节点的`Rune0[0]`和`Rune`中。
    1. 调用`p.push`将节点推入堆栈。

### 解析正则表达式

在解析开始的时候，初始化一个`parser`，存储了解析需要的一些状态。

函数主体是一个大的循环，
循环体是一个`switch`
检查`lookahead`符号来调用相应的处理函数读取、处理相应的单元。

- 默认分支

    默认分支用于处理普通的字符匹配。

    1. `nextRune`读取一个字符。
    1. 调用`parser.literal`，

- 左括号 `'('`

    1. 左括号标记了一个捕获组 *(capture group)* 的开始，
    故增加计数器`parser.numCap`。

# 一些收获

## 处理大小写不敏感的匹配

可以通过同样的化归方式来处理输入和模式。

# 参考

- [Regular Expression Matching Can Be Simple And Fast](https://swtch.com/~rsc/regexp/regexp1.html)
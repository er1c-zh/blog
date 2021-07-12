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
    Min, Max int        // min, max for OpRepeat 用于 x{Min,Max} 类型的节点
    Cap      int        // capturing index, for OpCapture 捕获组的index
    Name     string     // capturing name, for OpCapture 捕获组的名字
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

- `func (p *parser) concat() *Regexp`

    从上搜索堆栈，使用`"|"`或`"("`之后的节点构建一个`OpConcat`节点（用来匹配节点的`Subs`部分的连接体）。

    1. 首先调用`maybeConcat`合并堆栈上的字符序列的匹配。
    1. 从栈顶开始搜索`opLeftParen`和`opVerticalBar`或直到栈底。
    1. 当上一步搜索结束之后，遍历的下标指向上一个`opLeftParen`或`opVerticalBar`的前一个栈帧或者栈底，
    从下标开始切分成两个部分：下标之前的保留在栈中；下标及之后的节点截取出来继续处理。
        - 如果截取出来的部分为空，向栈中推入一个新的`OpEmptyMatch`节点。
        - 否则调用`p.collapse(subs, OpConcat)`创建一个`OpConcat`节点，推入栈中。

- `func (p *parser) collapse(subs []*Regexp, op Op) *Regexp`

    构建一个`Subs`是`subs`的`op`类型的节点，并返回。

    1. 快速路径，如果`subs`长度为一，直接返回。
    1. `newRegexp`创建一个新的`op`类型的节点。
    1. 遍历subs

        - 如果节点类型与`op`相同，那么将节点的`Sub`追加到新节点的`Sub`中，
        ——这样避免出现「连接」的「连接」这种情况，然后`reuse`节点。
        - 其他情况就简单的推入新节点的`Sub`。

    1. 如果要新建的节点是`OpAlternate`，调用`p.factor`来简化，
    如果简化后的节点的`Sub`中只有一个节点，那么简化成一个，`reuse`另一个。
    1. 返回新的节点。

- `func (p *parser) alternate() *Regexp`

    将堆栈中第一个**左括号**之上的所有节点替换为一个`OpAlternate`节点。

    1. 定位到第一个`opLeftParen`的位置，截取出不包括`opLeftParen`的所有节点作为`subs`；
    pop出堆栈中不包括`opLeftParen`的所有节点。
    1. 将第一步中获得的`subs`节点的最后一个调用`cleanAlt`清理。
    1. 如果`subs`为空，推入一个`OpNoMatch`节点；
    否则调用`p.collapse`构建一个包含`subs`的`OpLaternate`节点。

- `func (p *parser) factor(sub []*Regexp) []*Regexp`

    对于**选择**类型的节点，尝试将`Sub`合并。

    todo

- `func (p *parser) swapVerticalBar() bool`

    对于栈中第二个元素是`opVerticalBar`节点的情况，
    `swapVerticalBar`会交换两个节点，返回`true`。

    处理分为两种case，第一种情况：

    ```ditaa
    ditaa
    +-------------+
    | isCharClass | <----- stack top
    +-------------+
    |opVerticalBar|
    +-------------+
    | isCharClass |
    +-------------+
    |          {d}|
    +-------------+
    ```

    其中`isCharClass`包含：

    - 一个字符的`OpLiteral`
    - `OpCharClass`
    - `OpAnyCharNotNL`
    - `OpAnyChar`

    将两个`isCharClass`节点中优先级较低的通过`mergeCharClass`合并到较高的节点中，
    保留优先级较高的节点到栈顶，优先级较低的节点`reuse`。

    第二种情况：

    ```ditaa
    ditaa
    +-------------+
    | isCharClass | <----- stack top
    +-------------+
    |opVerticalBar|
    +-------------+
    |          {d}|
    +-------------+
    ```

    这种情况简单的交换两者的位置。

    特别的，代码中会对`stack[n-3]`调用`cleanAlt`来做一些优化，这里略过不提。

- `func (p *parser) literal(r rune)`

    1. 新建一个`OpLiteral`节点。
    1. 如果是大小写不敏感的匹配，调用`minFoldRune`将字符化归到最小编码。
    1. 将`r`保存到节点的`Rune0[0]`和`Rune`中。
    1. 调用`p.push`将节点推入堆栈。

#### 其他方法

- `func cleanAlt(re *Regexp)`

    todo

### 解析正则表达式

在解析开始的时候，初始化一个`parser`，存储了解析需要的一些状态。

函数主体是一个大的循环，
循环体是一个`switch`
检查`lookahead`符号来调用相应的处理函数读取、处理相应的单元。
最后在将传入的正则表达式全部读取、处理完成后，
依次调用`p.concat`、`p.swapVerticalBar`、`p.alternate`来
将栈中的节点合并。
最后返回语法树的根节点。

翻译的核心思路是**字符类**连接成**连接**类型，通过`|`构建成**选择**类型，
对于捕获组或单纯的分组的情况，嵌套一个**选择**类型。

下面详细分析各个情况的处理。

- 默认分支

    默认分支用于处理普通的字符匹配。

    1. `nextRune`读取一个字符。
    1. 调用`parser.literal`来新建一个`OpLiteral`节点并推入堆栈。

- 左括号 `'('`

    1. 左括号标记了一个捕获组 *(capture group)* 的开始，
    故增加计数器`parser.numCap`。
    1. 调用`p.op(opLeftParen)`生成一个伪op节点，
    并设置`Regexp.Cap`来保存捕获组的标号。

    特别的，这里的`opLeftParen`是一个伪op，
    类似的还有`opVerticalBar`，用于解析过程中放在堆栈中。

- 选择 `'|'`

    调用`parseVerticalBar`。

    `func (p *parser) parseVerticalBar() error`

    1. 调用`p.concat`重整堆栈上的节点，生成一个**连接**节点。
    1. 尝试`p.swapVerticalBar`将交换一个**选择**节点到栈顶。
    交换会将之前的节点当作**选择**的一个子节点来匹配，不改变语义。
    1. 如果失败，则新建一个`opVerticalBar`推入堆栈。

- 右括号 `')'`

    调用`parseRightParen`来处理。

    1. 首先调用`p.concat`来将栈顶的节点重整成一个`OpConcat`节点。
    1. 调用`p.swapVerticalBar`将可能存在的`opVerticalBar`节点移动到栈顶，并pop掉。
    1. 调用`p.alternate`将栈顶的一系列节点构建为一个`OpAlternate`节点。
    1. 检查堆栈的长度和内容，这个时候栈顶的两个节点应该分别为
    `opLeftParen`和`OpAlternate` *（也可能是`OpNoMatch`）* 节点。
    1. pop出最上面的两个节点。
    1. 根据`opLeftParen`节点的`regexp.Cap`来判断是否需要构建一个`OpCapture`节点。
        - 如果`regexp.Cap`为0，表示不需要捕获，直接将`OpAlternate`节点推入堆栈。
        - 否则将左括号的节点的`Op`改为`OpCapture`，然后将`OpAlternate`放到`Sub`中，
        最后推入新的`OpCaptrue`节点。

- 开始与结束 `'^'`和`'$'`

    根据`p.flags`判断是行匹配还是跨行匹配，分别插入相应的节点。

    `Op(Begin|End)(Text|Line)`

- 任意字符 `'.'`

    根据`.`是否支持匹配换行符推入`OpAnyChar(NotNL)?`节点。

- 自定义的字符类 `'['`

    `p.parseClass`

    1. 首先吃掉`'['`。
    1. 检查是否是 **“去除”** ，标记`sign`。
    1. 初始化`Rune`数组`class`用来存储字符类的内容，
    新建一个`OpCharClass`节点。

        `class`的格式是每两个rune标记一个范围。

    1. 循环读取直到字符类结束。
        - 尝试匹配POSIX和Perl风格的`-`。
        - 尝试匹配`[:alnum:]`。
        - 尝试匹配`[\p{Han}]`。
        - 尝试匹配`\d\w`。
        - 尝试匹配普通字符与普通的字符范围。

            1. `p.parseClassChar`根据是否需要转义，
            使用`p.parseEscape`或`nextRune`来获取字符。
            1. 检查下一个符号是否是`-`来判断是否是一个范围，如果是的话，
            重复上一步的操作读取一个字符作为范围的结束。
            1. 最后根据是否大小写敏感调用`append(Folded)?Range`将字符推入。

    1. `cleanClass`清理合并字符类。

        1. 排序，范围开始升序，范围结束降序。
        1. 合并重复。

    1. 如果是“去除”，`negateClass`获取补集。
    1. 将`class`存储到`OpCharClass`节点的`regexp.Rune`中，将节点推入堆栈。

- 重复 `'*','+','?','{}'`

    对于`'*' '+' '?'`，分别调用`p.repeat`创建`OpStar, OpPlus, OpQuest`节点。

    对于`{min, max}`，会首先利用`parseRepeat`尝试是否能正常的匹配，如果不能，
    将`{`作为一个普通的字符来处理。
    正确读取到`min, max`后，调用`p.repeat`来插入节点。

    `p.repeat`检查栈中的情况，将栈顶饿节点包装到新节点的`Sub`中，推入栈中。

## 简化

递归的检查语法树的每个节点，需要简化的类型是：

- `OpStar`
- `OpPlus`
- `OpQuest`
- `OpRepeat`

### 对于`OpStar` `OpPlus` `OpQuest`

首先递归的应用到子节点，然后调用`simplify1`来进行简化。

`func simplify1(op Op, flags Flags, sub, re *Regexp) *Regexp`
函数接受`OpStar`，`OpPlus`和`OpQuest`三种`op`参数，
来尝试简化`sub`和`re`节点：
1. 当`sub.Op`是`OpEmptyMatch`（空匹配不管重复多少次都是等效的）
或与`re.Op`相同且贪婪标记相同时（重复是幂等的），直接返回`sub`。
1. 检查`re`是否需要修改，不需要的话直接返回，否则构建一个新的节点返回。

### `OpRepeat`

简化的结果是消除`OpRepeat`，利用`OpStar`或`OpPlus`等节点来替换，
具体的实现分为以下情况：

1. 首先检查`x{0}`意味着匹配空串，用`OpEmptyMatch`来替换。
1. 然后递归的调用`Simplify`来简化`Sub[0]`节点。
1. 检查不设`Max`的情况：
    1. `x{0,}`类型与`x*`等价，转换成一个`OpStar`节点。
    1. `x{1,}`类型与`x+`等价，转换成一个`OpPlus`节点。
    1. `x{4,}`这种类型，可以使用`xxxx+`这种情况，
    因此，用三个`x`的节点和一个`OpPlus`的节点组成`OpConcat`替换一下。
1. `x{1}`类型是精准匹配，直接使用`sub`来替换。
1. `x{n,m}`是n个x和(m-n)个x?的组合，
特别的，将`x?x?x?`转换成`(x(x(x)?)?)?`会加快程序执行。
1. 最后没有处理的情况是非法的情况，返回一个`OpNoMatch`节点。

## 将语法树编译为要执行的程序

todo

# 一些收获

## 处理大小写不敏感的匹配

可以通过同样的化归方式来处理输入和模式。

## 运算符优先级

定义运算符常量时，可以将常量的数值的大小与运算符的优先级关联起来。

# 参考

- [Regular Expression Matching Can Be Simple And Fast](https://swtch.com/~rsc/regexp/regexp1.html)
---
title: "golang的regex实现 编译"
date: 2021-07-07T23:05:42+08:00
draft: false
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

本文首先分析记录关联的数据结构，然后分析编译的过程，
最后在后面的文章分析匹配的实现。

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

            如果是只匹配一个字符，尝试合并到之前的节点，或者新建一个`OpLiteral`节点；

            如果匹配类似Aa或者Bb这种如果忽略大小写之后是一个字符的情况，
            类似的，会尝试合并到前一个节点（以大小写不敏感的flag），或者新建一个`OpLiteral`节点。
    
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

    对于**选择**类型的节点，尝试将`Sub`的共同前缀提取出来。

    根据注释，这里会做如下的合并：

    > For example,
    >     ABC|ABD|AEF|BCX|BCY
    > simplifies by literal prefix extraction to
    >     A(B(C|D)|EF)|BC(X|Y)
    > which simplifies by character class introduction to
    >     A(B[CD]|EF)|BC[XY]

    <del>具体怎么实现的，等到有时间在看吧。</del>

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

    主要是将`OpCharClass`类型的节点的一些特殊情况做了处理。

    - 匹配任意字符的会替换成`OpAnyChar`。
    - 匹配除换行符的任意字符的会替换成`OpAnyCharNotNL`。
    - 如果给节点的`Rune`数组分配了太多的未使用的空间，通过重新`append`来清理。

        ```go
        if cap(re.Rune)-len(re.Rune) > 100 {
            // re.Rune will not grow any more.
            // Make a copy or inline to reclaim storage.
            re.Rune = append(re.Rune0[:0], re.Rune...)
        }
        ```

        至于这里为什么能实现这种效果，猜测和`append`的实现有关。

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

    `p.repeat`检查栈中的情况，将栈顶的节点包装到新节点的`Sub`中，推入栈中。

## 简化

递归的检查语法树的每个节点，需要简化的类型是：

- `OpStar`
- `OpPlus`
- `OpQuest`
- `OpRepeat`

**简化**流程会将`{m,n}`类型的语法消除，转换成`*/+/?`类型。

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

```go
prog, err := syntax.Compile(re)
```

利用`syntax.Compile`将简化过的语法树编译成可以可以执行的程序。

实现上，go的正则表达式包的匹配的通过构建一个执行指令的自动机，
来执行“能够匹配的字符串的指令列表”完成匹配。
所以，**编译**的结果是一个“程序”，
程序的内容是一系列指令，执行匹配的自动机可以通过执行指令来完成工作。

### 数据结构

1. `syntax.Prog`

    `Prog`表示一个编译好的程序。
    其中，
    - `Prog.Inst`是一个指令列表，可以通过下标寻址。
    - `Prog.Start`是开始指令的下标。
    - `Prog.NumCap`是该程序中捕获组的数量。

    ```go
    type Prog struct {
        Inst   []Inst
        Start  int // index of start instruction
        NumCap int // number of InstCapture insts in re
    }
    ```
1. `syntax.Inst`

    `Inst`表示程序中的一个指令。

    ```go
    type Inst struct {
        Op   InstOp
        Out  uint32 // all but InstMatch, InstFail
        Arg  uint32 // InstAlt, InstAltMatch, InstCapture, InstEmptyWidth
        Rune []rune
    }
    ```

    其中，`Op`表明了该指令的类型，`InstOp`是定义的类型，有若干常量。

    ```go
    InstAlt InstOp = iota // 分支，类似于if-else
    InstAltMatch
    InstCapture // 标记捕获组的开始和结束
    InstEmptyWidth // 匹配位置
    InstMatch // 匹配完成
    InstFail // 匹配失败
    InstNop // nop
    InstRune // 匹配Inst.Rune中的字符
    InstRune1 // 匹配一个字符
    InstRuneAny // 匹配任意字符
    InstRuneAnyNotNL // 匹配除换行符之外的任意字符
    ```

    `Out`的值是`Prog.Inst`的下标，是需要执行的下一个指令的下标；
    `Arg`的值与`Inst.Op`有关，
    比如对于`InstAlt`来说是另一个分支的指令的下标；
    `Rune`存储了需要匹配的字符。

1. `syntax.frag`

    `frag`表示一个编译好的程序的片段，`i`表示片段开始的指令的下标，
    `out`是一个`patchList`对象，业务上的含义是整个片段的下一步的指令，
    编译中，有时无法立即确定下一步需要跳转到哪里，所以需要这种用于暂存编译过程中需要补全信息的位置。
    `patchList`的详细解释在下文。

    ```go
    // A frag represents a compiled program fragment.
    type frag struct {
        i   uint32    // index of first instruction
        out patchList // where to record end instruction
    }
    ```

1. `syntax.patchList`

    `patchList`是一系列指令指针的列表。
    因为编译是自顶向下的，
    所以存在生成一个指令实例`syntax.Inst`时，
    无法确定`Inst.Out`或`Inst.Arg`的情况。
    为了解决这种问题，引入了`patchList`。
    特别的，利用需要填充的字段来存储了列表中的其他指针，
    `patchList`只需要存储头尾指针。

    当`head`为0时表示是空列表，
    当`l.head&1==0`时，指向`p.inst[l.head>>1].Out`，
    当`l.head&1==1`时，指向`p.inst[l.head>>1].Arg`。
    
    ```go
    type patchList struct {
        head, tail uint32
    }
    ```

1. `syntax.compiler`

    `compiler`是执行编译的对象，存储了表示编译结果的`syntax.Prog`实例，
    具体的编译工作由`compiler.compile`完成。

    ```go
    type compiler struct {
        p *Prog
    }
    ```

### 编译

编译的入口函数是`syntax.Compile`，
函数首先初始化一个`compiler`对象，
然后通过`compiler.compile`遍历语法树、生成指令表并表示整个程序的片段实例。
最后，将`patchList`补充好，保存入口的下标，返回表示编译好的程序的`Prog`实例。

```go
func Compile(re *Regexp) (*Prog, error) {
    var c compiler
    c.init() 
    f := c.compile(re)
    f.out.patch(c.p, c.inst(InstMatch).i)
    c.p.Start = int(f.i)
    return c.p, nil
}
```

1. 初始化

    增加捕获组的数量，将指令表开头初始化为InstFail

1. `c.compile`

    `c.compile`的工作是将语法树转换成指令表，并串联起来。

    实现上，`c.compile`自顶向下遍历语法树，
    对于每个节点，会在指令表中生成新的指令，
    如果有子节点的通常还会递归的调用`c.compile`编译子节点，
    下面具体的分析编译的实现。

    `c.compile`中是一个大`switch`，
    首先根据是否有子节点分为两类来分析。

    - 需要处理子节点

        - 连接 `OpConcat`

            遍历节点的子节点，
            调用`compiler.compile`递归的生成指令，获得片段；
            从第二个节点开始，会调用`compiler.cat(f1, f2 frag)`将指令连接起来。

            `cat`的内容十分简单，
            如果两个片段有指向指令列表的第一个元素——意味着匹配失败——直接返回表示匹配失败的片段。
            其他的时候，调用第一个片段的`frag.out.patch`方法，
            将第一个片段的所有在`patchList`中的位置（`Inst.Out`和`Inst.Arg`）填充为`f2`表示的指令。
            返回一个新的片段，入口是`f1`的入口，`patchList`是`f2`的`patchList`。

            从本质上来说，`cat`起到类似于“合并”片段的作用。

        - 选择 `OpAlternate`

            遍历子节点，首先用`compiler.compile`生成指令，获得片段，
            然后调用`compiler.alt(f1, f2 frag)`来生成一个分支。

            `alt`生成一个新的`InstAlt`指令，
            `Inst.Out`指向`f1`的入口，`Inst.Arg`指向`f2`的入口。
            
            调用`f1.out.append(c.p, f2.out)`来将`f1`和`f2`的`patchList`合并起来，
            作为新的`InstAlt`指令的片段的`patchList`。

        - 重复 `*`

            第一步生成子节点的指令，
            然后通过调用`compiler.star`实现重复匹配的功能。

            `star`会生成一个`InstAlt`指令，
            根据是否贪婪，分别将重复的指令放在`Inst.Out`或`Inst.Arg`中，
            将另一个字段放置到`patchList`中，等待后续的填充。

            最后，将需要重复的单元的`patchList`用新的`InstAlt`节点填充，
            这样就生成了类似循环的指令循环，

        - 重复 `+`

            `+`的实现比较巧妙，功能由`compiler.plus`实现，可以实现上看到没有创建新的指令。

            通过简单的将返回的片段的开始下标由指向`InstAlt`转换为指向重复单元，就实现了“至少匹配一次”的目的。

            ```go
            func (c *compiler) plus(f1 frag, nongreedy bool) frag {
                return frag{f1.i, c.star(f1, nongreedy).out}
            }
            ```

        - 重复 `?`

            具体的由`compiler.quest`实现，对比`compiler.star`来看也十分巧妙。

            区别于`star`最后会将`InstAlt`指令填充到重复单元的`patchList`上，
            `quest`将重复单元的`patchList`和`InstAlt`的`patchList`合并起来，等待填充。

            这样就类似于实现了“匹配一次（走`InstAlt`的重复单元分支）或零次（直接走延迟填充的节点的分支）”的目标。

        - 捕获 Capture

            捕获功能比较特殊的是引入了新的指令`InstCapture`，
            在遇到`OpCapture`时，使用`compiler.cat`将两个`InstCapture`和子节点编译成的指令连接起来。

            `InstCapture`指令由`compiler.cap`生成。
            过程比较简单：
            
            1. 将传入的参数保存到指令的`Inst.Arg`字段。
                
                `Inst.Arg >> 1`是对应的捕获组的序数，
                `Inst.Arg & 1`标记捕获组的开始（0）或结束（1）。
                
            1. 将`compiler`的捕获组数量增加。

    - 不需要处理子节点

        - 匹配字符的节点

            对于`OpLiteral`/`OpCharClass`/`OpAnyCharNotNL`/`OpAnyChar`四种节点，
            会调用`compiler.rune(runeArrayNeedMatch, flags)`来生成指令。
            
            1. 首先通过`compiler.inst(InstRune)`生成一个`InstRune`类型的指令，
            写入到指令表中。
            1. 随后将传入的需要匹配的rune放置到指令中的`Inst.Rune`字段。
            1. 过滤出标记中的`FlodCase`，处理后存储到`Inst.Arg`
            1. 将新指令的`Inst.out`加入到返回的片段的`patchList`中。
            1. 最后将三种情况下的指令的类型修改为更特化的指令。
                - 如果是匹配任意一个字符的时候，替换为`InstRuneAny`。
                - 如果是匹配除换行符的所有字符，替换为`InstRuneAnyNotNL`。
                - 如果匹配的只有一个字符，那么替换为`InstRune`。

            特别的，对于`OpLiteral`，会遍历需要匹配的字符列表，
            重复执行上述的过程。对于返回的结果，会调用`compiler.cat`将若干匹配的指令连接起来。

        - 匹配空串

            `compiler.nop`会生成一个`InstNop`指令。

        - 匹配失败

            `compiler.fail`生成一个`Inst.Op`为0的指令实例。

        - 匹配位置

            对于`OpBeginLine`/`OpEndLine`/`OpBeginText`/
            `OpEndText`/`OpWordBoundary`/`OpNoWordBoundary`这类匹配位置的节点，
            会调用`compiler.empty`生成一个`InstEmptyWidth`指令，
            将要匹配的位置信息存储到指令的`Inst.Arg`中。
            最后，将`Inst.Out`连接进`patchList`。

1. 增加程序的结尾

    根据实现，
    最后返回的`frag`对象的`frag.i`表示程序的开头，
    `frag.out`表示的是匹配运行的结尾处的“`Inst.Out`或`Inst.Arg`”，
    为了结束匹配，所以生成一个表示匹配完成的`InstMatch`指令并将它填充到最后的`patchList`上。
    这样，在匹配的程序运行到最后的时候，就会执行`InstMatch`指令，标志匹配的完成。

1. 最后，将程序开始的下标，存储到要返回的`Prog`实例中。

## 收尾工作

### 编译一趟匹配

通过`compileOnePass`来检查前面编译产物`syntax.Prog`是否可以支持一趟匹配。
只有当在`InstAlt`指令，可以无歧义的判断下一步分支时，才可能可以转换为一趟匹配的`Prog`。

1. 检查是否可以转换为一趟匹配

    1. 首先检查了`Prog`是否是锚定的。

        如果`Prog`不是仅从匹配文本开头位置开始，那么就无法转换，返回nil。

    1. 依次检查`Prog`指令表中的指令。

        如果发现任何匹配结束不是文本结束位置，那么就无法转换，返回nil。

1. `onePassCopy`复制一份`prog`，在复制过程中，尝试重写指令为可以转换的形式。

1. `makeOnePass`尝试构建一份一趟匹配的`prog`，如果失败，返回nil。

    1. 如果过长，直接返回失败。

    1. 深度优先扫描指令树。

        目标是检查指令树是否可以转换为一趟匹配，并完成转换。

        如果分支指令的两个分支在不再匹配任何字符后，
        均能完成匹配，返回失败，无法转换。
        我暂时还没想出来什么样的表达式会造成这种情况，
        简单猜测可能是捕获、位置匹配等。

        另外，在遇到分支指令时（`InstAlt`/`InstAltMatch`），
        先递归的检查两个分支，构建好他们各自的匹配的字符集合，
        然后利用`mergeRuneSets`来合并，
        该函数发现两个分支匹配的字符有重复时，
        会返回失败，标志该`prog`无法转换成一趟匹配；
        否则返回该指令匹配的字符。

        遇到匹配字符的指令时，会计算好该指令会匹配的字符集合，
        放置到`onePassRunes`中，等待后续的调用。

        另外，`onePassRunes`起到了指令的跳转表的作用。
        可以通过：如果字符匹配到了该指令的`onePassRunes`数组中的下标i的字符，
        到指令的`inst.Next[i/2]`数组查找到下一个需要执行的指令。

1. 清理，并返回。

# 一些收获

## 处理大小写不敏感的匹配

可以通过同样的化归方式来处理输入和模式。

## 运算符优先级

定义运算符常量时，可以将常量的数值的大小与运算符的优先级关联起来。

# 参考

- [Regular Expression Matching Can Be Simple And Fast](https://swtch.com/~rsc/regexp/regexp1.html)
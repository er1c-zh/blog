---
title: "rust宏"
date: 2023-05-19T17:14:28+08:00
draft: false
tags:
    - rust
    - what
    - how
---

rust有`declarative`宏和`procedural`宏。

<!--more-->

# 宏与函数的区别

- 宏可以用于元编程。
- 宏可维护性、可读性较差。

# `declarative`宏

以一个简化的`vec`宏为例子：

```rust
let v: Vec<u32> = vec![1, 2, 3];

#[macro_export] // annotation表明这个宏在所处的crate被引入后可用。
macro_rules! vec { // macro_rules! 表明这是一个declarative宏，名字是vec。
    ( $( $x:expr ),* ) => { // ( $( $x:expr ),* ) 是一个模式
        {
            let mut temp_vec = Vec::new();
            $(
                temp_vec.push($x);
            )*
            temp_vec
        }
    }
}
```

自顶向下的分析：

`macro_rules! vec { ... }`定义了一个`vec`宏，其中`vec`是一个标识符。

宏由若干`Matcher => Transcriber`组成。
执行时，类似`match`，匹配`Matcher`，执行`Transcriber`。

`Matcher`是一个用来匹配rust代码结构的表达式。
`( $( $x:expr ),* )`是一个`Matcher`。

## Metavariables

其中，`$x:expr`类似于正则表达式的capture group，
捕获expr代码片段，存储到变量x中。

- `item` `item`是crate的一个部分 [items](https://doc.rust-lang.org/stable/reference/items.html)。
- `block` 代码块。
- `stmt` Statement。
- `pat_param` PatternNoTopAlt。 Pattern用来匹配结构体的值或绑定变量到结构体的值，若干PatternNoTopAlt组成了Pattern。
- `pat` 至少一个PatternNoTopAlt。
- `expr` Expression。
- `ty` Type。
- `ident` IDENTIFIER_OR_KEYWORD or RAW_IDENTIFIER。
- `path` TypePath。
- `tt` TokenTree，token或者Delimiters与其中的tokens。
- `meta` Attr的内容。
- `lifetime` LIFETIME_TOKEN。
- `vis` VISIBILITY QUALIFIER。
- `literal` 字面量表达式。

## Repetitions

通过形如`$(...){可选的分隔符}{*|+|?}`来表示重复任意次数、至少一次、零或一次。

## 作用域、引入和导出

因为历史原因，
宏有两种作用域：

- 文本作用域 取决于代码出现在文件中的顺序（多文件也如此）
- 基于path的作用域 与rust module的path一致

### `macro_use`

- 应用于module，使该module中的宏扩散出module。
- 从其他crate引入宏。
    - 只能引入`macro_export`的宏。

### 基于path的作用域

`#[macro_export]`修饰的宏被视作定义在crate根作用域中，
也可以通过`#[macro_use(xxx)]`被其他crate引用。


## Hygiene

对宏的引用是只会搜寻已经出现的宏，且原样展开(expanded as-is)。

可以通过`$crate::xxx!()`来引用还未定义的宏。

## 一些避免歧义的限制

考虑到未来的语言特性可能引起的歧义，有一些额外的限制。

# `procedural`宏: 用来根据属性(Attributes)生成代码

Procedural宏接受一些代码作为入参，
产生另一些代码作为输出（而不是匹配模式和替换）。

Procedural macro有三种类型：

- 自定义derive
- 类attribute
- 类function

## 一个例子

```rust
use proc_macro; // procedural macro需要的lib

#[some_attribute] // some_attribute 根据不同类型的macro替换为不同的attribute
pub fn macro_name(input: TokenStream) -> TokenStream {
    // ...
}
```

## custom derive marco

1. 声明crate是一个procedural macro crate。

    > 目前，procedural macro需要放在专属的crate中。

1. 编写macro。

    ```rust
    use proc_macro::TokenStream;
    use quote::quote;
    use syn; // 将TokenStream编译

    // 定义了 HelloMacro 宏
    // usage:
    // #[derive(HelloMacro)]
    // struct XXX;
    #[proc_macro_derive(HelloMacro)]
    pub fn macro_fn(input: TokenStream) -> TokenStream {
        // ...
    }
    ```

## Attribute-like macros

```rust
#[route(GET, "/")]
fn index() {
    // ...
}

#[proc_macro_attribute] // 表明这是一个 Attribute-like macro
pub fn route(attr: TokenStream, item: TokenStream) -> TokenStream {
    // ...
}
```

例子定义了一个`route`宏。
入参`attr`是attribute的内容，即：`GET, "/"`。
`item`是attribute `route` 关联到的对象的Token流，即：`fn index() { ... }`。

## Function-like macros

Function-like宏使用的样子看起来像函数调用。

```rust
let sql = sql!(SELECT * FROM posts WHERE id=1;)

#[proc_macro]
pub fn sql(input: TokenStream) -> TokenStream {
    // ...
}
```

`input`接受到`SELECT * FROM posts WHERE id=1;`。

# Reference

- [rust doc](https://doc.rust-lang.org/stable/book/ch19-06-macros.html)
- [完整的宏定义语法](https://doc.rust-lang.org/stable/reference/macros-by-example.html)

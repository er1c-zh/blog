---
title: "输入方式提升计划"
date: 2023-05-16T20:13:42+08:00
draft: false
tags:
    - other
    - how
---


迁移到colemak & 双拼。

<!--more-->

# colemak

## tl;dr

- 方案

    - pc切换成colemak，手机维持qwerty。
    - vim修改了`hjkl`相关的键位。
    - 得益于colemak的键位，常用快捷键没有修改。

- 需要考虑的点：

    1. os是否支持。
    1. 双拼在mac上的映射是key映射的，会受到影响。
    1. vim
    1. 快捷键

- 经历

    1. 整个切换小一个月：qwerty 70 wpm -> colemak 60wpm。
    1. 一个周熟悉键位，一个周切换到colemak工作，一个周提升到60wpm，中间有一个周假期没有打字。
    1. 最痛苦的是中间那个周，两种键位都不是很流畅：着急的工作切回qwerty大概一百个单词能熟悉起来。
    1. 做不到两个layout无缝切换。
    1. 1个月之后，63wpm。

## 解决方案

主力打字的机器都是mac，对colemak的支持很好。

### vim && vscode && jetbrain

最大的阻碍是vim的键位。

我选择将`hjkl`映射成`hnei`，影响到的键位也尽量小的做了修改。

```vimrc
""""""""""""""""""""""
" for colemak        "
""""""""""""""""""""""
" movement key
noremap n j
noremap N J
noremap <C-w>n <C-w>j
noremap e k
noremap E K
noremap <C-w>e <C-w>k
noremap i l
noremap I L
noremap <C-w>i <C-w>l

noremap k n
noremap K N
noremap l i
noremap L I
```

**vscode的vim插件支持从`.vimrc`中读取键位map。**

settings搜索：

```@ext:vscodevim.vim vimrc```

或者`settings.json`:

```json
{
    "vim.vimrc.enable": true,
}
```

**jetbrain的vim插件支持读取`~/.ideavimrc`。**

## 一些链接

- [练习网站](https://gnusenpai.net/colemakclub/)

# 双拼

小鹤双拼。

mac上使用了rime。

## 练习日志

- 0606 开始。
- 0619 需要一些思考时间来确认按键。

## rime改动

1. 替换了 ` * 等符号在中文模式下直接输出。
1. `squirrel.custom.yaml#keyboard_layout: last` 使用之前的键盘格式。

## 一些链接

- [改了一个支持不同键盘格式的练习页面](https://sp.er1c.dev)

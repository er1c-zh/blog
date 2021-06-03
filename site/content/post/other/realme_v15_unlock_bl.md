---
title: "Realme V15解锁bootloader并刷入TWRP"
date: 2021-06-03T00:39:43+08:00
draft: false
tags:
    - other
    - realme v15
---

昨日购入了一台realme v15，
搜索了之后发现官方没有放出这个型号的解锁BL工具，
一番尝试之后，
幸运的解锁了BL，刷入TWRP。

**操作有风险，本文不保证正确性。**

<!--more-->

# TL;DR

1. 官方的深度测试工具使用realme V3型号的能够成功申请并解锁。

    > 系统版本 realme UI RMX3092_11_A.33

1. 刷入的TWRP是LR.Team([微博](https://weibo.com/u/6033736159?profile_ftype=1&is_all=1))制作的realme x7型号的版本。

# 遇到的问题

## 进入fastboot的页面和教程说的不一样

联发科的fastboot之后就是两行字，这个时候执行

```powershell
.\fastboot.exe devices
```

是可以看到手机已经进入fastboot模式了。

如果没有可以尝试安装驱动等解决一下。

## 连adb正常但进入fastboot模式电脑无法识别

下载了驱动精灵自动安装了一个驱动之后就可以了。
    
## play商店不显示

下载了一个play商店的安装包就可以了。

尝试过使用一些google框架的安装app，效果不太好，
realme出厂其实是带了google框架的。

# 参考

- [官方解锁bootloader的帖子](https://www.realmebbs.com/post-details/1335782085125746688)
- [一个网站的解锁bl的说明](http://rom.7to.cn/jiaochengdetail/17105)
- [realme v3解锁bootloader的工具](https://xinkid.lanzoui.com/iMU40kcihli)，我用的是这个成功申请解锁了
---
title: "Diy Keyboard"
date: 2023-05-16T20:13:02+08:00
draft: true
---

DIY一个分体式键盘。

<!--more-->

# 目标

- 分体
- colemak & mac
- 有线

# PCB

- 2 * MCU AVR ATMEGA32U4-MU
- 链接两个split
- 2 * type-c 充电 + 连接
- 2 * 旋钮 亮度 + 音量
- switcher colemak / qwerty && win / mac

## MCU

ATMEGA32U4-MU

> [ATMEGA32U4-MU manual](https://atta.szlcsc.com/upload/public/pdf/source/20170515/1494827289828.pdf)
> 
> [VCC及相关引脚](https://electronics.stackexchange.com/questions/330186/do-i-have-to-provide-vcc-to-every-vcc-pin-on-atmega32u4-mcu)

## type-c

[type-c引脚介绍](https://zhuanlan.zhihu.com/p/439068141)

## 连接两个split

选择qmk的文档提到的TRS接口+serial驱动。

> 一篇对TRS的介绍： [TRS？ TRRS？ 正式录制前，您确保麦克风的音频线插对了吗？](https://zhuanlan.zhihu.com/p/144233538)

[TRS母座](https://so.szlcsc.com/global.html?c=&k=C5123139)

# 外壳


# Reference

- [一个开源的EDA KiCad](https://www.kicad.org/)
- [免费的3D建模软件 blender](https://www.blender.org/)
- [qmk 开源键盘固件](https://docs.qmk.fm/#/)
- [avr 和 arm 的关系](https://www.geeksforgeeks.org/difference-between-avr-and-arm/)
- [HID manual](https://www.usb.org/sites/default/files/documents/hut1_12v2.pdf)
- [keyboard layout editor](http://www.keyboard-layout-editor.com/)
- [根据KLE生成CAD file](http://builder.swillkb.com/)
- [绘制EDA](https://www.zfrontier.com/app/flow/eVz53QMw7VMA)

# MEMO

## keyboard-layout-editor raw data

```plain
[{x:3.5},"#\n3",{x:10.5},"*\n8"],
[{y:-0.875,x:2.5},"@\n2",{x:1},"$\n4",{x:8.5},"&\n7",{x:1},"(\n9"],
[{y:-0.875,x:5.5},"%\n5",{a:7},"",{x:5.5,a:4},"^\n6"],
[{y:-0.875,x:1.5},"!\n1",{x:14.5},")\n0",{a:7},"",""],
[{y:-0.995,x:0.5},""],
[{y:-0.38,x:3.5,a:4},"E",{x:10.5},"I"],
[{y:-0.875,x:2.5},"W",{x:1},"R",{x:8.5},"U",{x:1},"O"],
[{y:-0.875,x:5.5},"T",{a:7,h:1.5},"",{x:5.5,a:4},"Y"],
[{y:-0.875,x:1.5},"Q",{x:14.5},"P",{a:7},"",""],
[{y:-0.995,x:0.5},""],
[{y:-0.38,x:3.5,a:4},"D",{x:10.5},"K"],
[{y:-0.875,x:2.5},"S",{x:1},"F",{x:8.5},"J",{x:1},"L"],
[{y:-0.875,x:5.5},"G",{x:6.5},"H"],
[{y:-0.875,x:1.5},"A",{x:14.5},":\n;",{a:7},"","enter"],
[{y:-0.995,x:0.5},""],
[{y:-0.63,x:6.5,h:1.5},""],
[{y:-0.75,x:3.5,a:4},"C",{x:10.5},"<\n,"],
[{y:-0.875,x:2.5},"X",{x:1},"V",{x:8.5},"M",{x:1},">\n."],
[{y:-0.875,x:5.5},"B",{x:6.5},"N"],
[{y:-0.875,x:1.5},"Z",{x:14.5},"?\n/",{a:7},"shift",""],
[{y:-0.995,x:0.5},"shift"],
[{y:-0.38,x:3.5},"",{x:10.5},""],
[{y:-0.875,x:2.5},"",{x:1},"",{x:8.5},"",{x:1},""],
[{y:-0.75,x:0.5},"","",{x:14.5},"","",""],
[{r:30,rx:6.5,ry:4.25,h:2},"",{h:2},"",""],
[{x:2},""],
[{r:-30,rx:13,x:-3},"",{h:2},"",{h:2},""],
[{x:-3},""]
```
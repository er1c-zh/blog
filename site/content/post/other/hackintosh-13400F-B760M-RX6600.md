---
title: "Hackintosh 13400F B760M RX6600"
date: 2023-06-19T20:26:15+08:00
draft: false
tag:
    - how
    - hackintosh
---

Hackintosh on i5-13400F/B760M/RX6600.

<!--more-->

# OpenCore i5-13400F/B760M/RX6600 EFI

## Feature

- OpenCore 0.9.2
- Ventura 13.4
- All kext release version
- [x] 传感器 sensor
    - [x] CPU temperature sensor
        - boot-args 增加 `lilucpu=17` ，覆盖cpu generation。
    - [x] GPU sensor
        - RadeonSensor.Kext
        - SMCRadeonSensor.Kext
- [x] 睿频 turbo boost
- [x] 休眠唤醒
- [x] iService/AirDrop/etc

## Specs

- CPU i5-13400F
- [MS-终结者 B760M D4](https://www.maxsun.com.cn/2023/0104/5870.html) / [MAXSUN Terminator B760M D4](https://www.maxsun.com/products/terminator-b760m-d4)
- RX6600
- Gloway光威 32GB DDR4 3200 台式机内存条 天策系列-皓月白 * 2
- ZHITAI TiPlus5000 1TB
- fenvi FV-T919

## IMG

![system-info](https://github.com/er1c-zh/hackintosh-b760m-i5-13400f-EFI/raw/master/doc/system-info.png)

### wifi && bluetooth

![wifi](https://github.com/er1c-zh/hackintosh-b760m-i5-13400f-EFI/raw/master/doc/wifi.png)
![ble](https://github.com/er1c-zh/hackintosh-b760m-i5-13400f-EFI/raw/master/doc/ble.png)

### sensor

![sensor](https://github.com/er1c-zh/hackintosh-b760m-i5-13400f-EFI/raw/master/doc/sensor-bar.png)

![sensor](https://github.com/er1c-zh/hackintosh-b760m-i5-13400f-EFI/raw/master/doc/sensor.png)

# Usage

1. 按照[OpenCore-Install-Guide](https://dortania.github.io/OpenCore-Install-Guide/)构建OpenCore启动盘。

    - 可能需要生成USBMap。
    - 必要时增加一些`boot-args`输出日志。

1. [luchina-gabriel/BASE-EFI-INTEL-DESKTOP-13THGEN-RAPTOR-LAKE](https://github.com/luchina-gabriel/BASE-EFI-INTEL-DESKTOP-13THGEN-RAPTOR-LAKE)中列出的BIOS设置项。
1. 进行安装，需要按空格键显示安装的选项。

    - 大概装了半个小时，ETA不准。
    - 会有3次（？）黑屏。

1. 按照[OpenCore-Install-Guide](https://dortania.github.io/OpenCore-Install-Guide/)的`Post-Install`完善安装。

    - 最基本的，挂载EFI，将启动盘中的部分拷贝进去。
    - **生成新的序列号避免问题。**

1. Enjoy it!

# MEMO

- OpenCore的选择器按空格才会展示其他选项。 Some options won't show unless space pressed.
- 字体图标渲染发虚，关闭hdr解决。 Disable HDR can solve font blurry for me.

# Reference

- [EFI repo](https://github.com/er1c-zh/hackintosh-b760m-i5-13400f-EFI)
- [OpenCore-Install-Guide](https://dortania.github.io/OpenCore-Install-Guide/)
- [luchina-gabriel/BASE-EFI-INTEL-DESKTOP-13THGEN-RAPTOR-LAKE](https://github.com/luchina-gabriel/BASE-EFI-INTEL-DESKTOP-13THGEN-RAPTOR-LAKE)
- [nsinm/13400F-B760M-EFI](https://github.com/nsinm/13400F-B760M-EFI)

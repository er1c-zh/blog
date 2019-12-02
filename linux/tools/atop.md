# atop

用于监控系统资源消耗的程序。

## 安装

use apt

```shell
sudo apt install atop
```

## 窗口

有若干内容的聚合页面。

*按键切换到目标页面*

1. g(eneric) 默认的页面，包含一些常见信息。
1. m(emory) 内存，显示不同进程的内存占用信息。
1. d(isk) 磁盘读写，不同进程对磁盘资源的使用情况。
1. n(etwork) 网络相关的情况。
1. s(cheduling characteristics) 调度相关的信息。
1. v(arious characteristics) 多种其他的信息。
1. c(ommand) 进程的命令。
1. u(ser) 根据用户维度展示进程信息。
1. p(rogram) 根据程序维度聚合进程信息。

*按键对当前页面进行排序等操作*

1. C 根据CPU占用排序
1. M 根据MEM排序
1. D 根据磁盘排序
1. N 根据网络排序
1. A automatically 自动排序，根据最紧缺的资源进行排序，展示最瓶颈的程序。

*其他的功能*

1. z 冻结当前的屏幕。
1. i 更改采样的时间间隔。
1. r reset 重置所有的计数器。

### PRC process 进程信息

1. sys 内核态运行时间
1. user 用户态运行时间
1. #proc 总进程数
1. #zombie 僵尸进程数
1. #exit 采样周期期间退出的进程数

### CPU cpu整体的信息

以下四部分加起来等于核数:

1. sys/usr 内核态和用户态的运行时间比例
1. irq 处理中断占用的cpu时间比例
1. idle 完全空闲的的时间比例
1. wait 进程等待磁盘IO导致CPU空闲状态的时间比例

### cpu 每个核心的情况

与[CPU](### CPU cpu整体的信息)含义保持一致。

### CPL 显示CPU负载情况

1. avg1/avg5/avg15 一分钟、五分钟、十五分钟平均负载
1. csw 上下文交换次数 时间单位 todo
1. intr 中断发生次数 时间单位 todo

### MEM 内存

1. total 内存总量
1. free 空闲内存大小
1. cache 页缓存内存大小 page cache
1. buff 文件缓存内存大小 filesystem meta data
1. slab 系统内核使用的内存大小

### SWP 交换区

1. tot 总量
1. free 空闲

### PAG 

## usage



## 参考

1. [atoptool.nl](https://www.atoptool.nl/)



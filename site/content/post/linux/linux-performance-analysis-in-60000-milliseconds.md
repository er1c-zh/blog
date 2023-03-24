---
title: "60,000ms分析Linux性能"
date: 2023-03-23T19:47:21+08:00
draft: false
tags:
    - memo
    - linux
    - command
    - how-to
    - translation
---

[Linux Performance Analysis in 60,000 Milliseconds](https://netflixtechblog.com/linux-performance-analysis-in-60-000-milliseconds-accc10403c55) 翻译。

<!--more-->

> You log in to a Linux server with a performance issue: what do you check in the first minute?

当登入有性能问题的linux服务器时，首先需要检查哪些方面呢？

> At Netflix we have a massive EC2 Linux cloud, 
> and numerous performance analysis tools to monitor and investigate its performance. 
> These include Atlas for cloud-wide monitoring, and Vector for on-demand instance analysis. 
> While those tools help us solve most issues, 
> we sometimes need to login to an instance and run some standard Linux performance tools.

Netflix有很大的EC2 Linux集群，和很多用于监控、调查这个集群的性能问题的分析工具：
Atlas用来监控集群级别的问题；Vector用来在需要时分析实例。

尽管这些工具能帮助解决大多数问题，
有的时候，我们还是需要登入一个实例并通过一些标准Linux性能工具（来排查问题）。

# 总结 First 60 Seconds: Summary

> In this post, the Netflix Performance Engineering team will show you the first 60 seconds of an optimized performance investigation at the command line, using standard Linux tools you should have available. In 60 seconds you can get a high level idea of system resource usage and running processes by running the following ten commands. Look for errors and saturation metrics, as they are both easy to interpret, and then resource utilization. Saturation is where a resource has more load than it can handle, and can be exposed either as the length of a request queue, or time spent waiting.

Netflix性能工程团队将会在本文中展示通过常见的标准Linux工具完成性能优化调查的第一步。
可以通过下列的十个指令对整个系统的资源使用和当前运行中的进程有一个大概的了解。
首先寻找异常和饱和的指标（因为较为容易理解），然后是资源的利用情况。
饱和指一个资源已经被过度的使用，可以通过请求队列的长度（过长）或者（请求的）等待时间来发现。

```shell
uptime
dmesg | tail
vmstat 1
mpstat -P ALL 1
pidstat 1
iostat -xz 1
free -m
sar -n DEV 1
sar -n TCP,ETCP 1
top
```

> Some of these commands require the sysstat package installed. The metrics these commands expose will help you complete some of the USE Method: a methodology for locating performance bottlenecks. This involves checking utilization, saturation, and error metrics for all resources (CPUs, memory, disks, e.t.c.). Also pay attention to when you have checked and exonerated a resource, as by process of elimination this narrows the targets to study, and directs any follow on investigation.

一部分指令需要安装`sysstat`包。
这些指令返回的指标可以帮助你完成USE方法：一个定位性能瓶颈的方法论。
USE包括了检查所有资源（CPU、内存、磁盘等）的用量、饱和度、错误。
在检查并排出一个资源（引起问题的可能性）时需要（额外的）小心，
因为排除一个资源后会减小调查的范围，进而影响到后续的调查。

> The following sections summarize these commands, with examples from a production system. For more information about these tools, see their man pages.

后续的章节会结合生产环境的例子来总结这些指令。
通过manual来查看这些命令的更多的信息。

# `uptime`

```shell
$ uptime 
23:51:26 up 21:31, 1 user, load average: 30.02, 26.43, 19.02
```

> This is a quick way to view the load averages, which indicate the number of tasks (processes) wanting to run. On Linux systems, these numbers include processes wanting to run on CPU, as well as processes blocked in uninterruptible I/O (usually disk I/O). This gives a high level idea of resource load (or demand), but can’t be properly understood without other tools. Worth a quick look only.

`uptime`是一个快捷的查看当前平均负载 *(load averages)* 的指令，
平均负载反映了当前期待运行的进程。
在Linux中，这些数字还包含了阻塞在不可中断IO中的进程，通常是磁盘IO。
这个结果只能给到一个很高维度的资源负载情况，无法在不结合其他工具的情况下进行正确的解读。
所以只用来当作一个简单的查看工具。

> The three numbers are exponentially damped moving sum averages with a 1 minute, 5 minute, and 15 minute constant. The three numbers give us some idea of how load is changing over time. For example, if you’ve been asked to check a problem server, and the 1 minute value is much lower than the 15 minute value, then you might have logged in too late and missed the issue.

这三个数是一组以1min 5min 15min作为常数的指数下降滑动平均值三元组。
可以通过这三个数值之间的大小关系来推测系统负载随着时间的变化情况。
举个例子，如果在登入有问题的服务器时发现一分钟的数值远低于十五分钟的数值时，可能意味着登入过晚，已经错过了现场。

> In the example above, the load averages show a recent increase, hitting 30 for the 1 minute value, compared to 19 for the 15 minute value. That the numbers are this large means a lot of something: probably CPU demand; vmstat or mpstat will confirm, which are commands 3 and 4 in this sequence.

上述的例子中可以看到系统的负载正在上升，意味着对cpu的需要不断的增加，
可以通过`vmstat`或者`mpstat`来确定。

# `dmesg | tail`

```shell
$ dmesg | tail
[1880957.563150] perl invoked oom-killer: gfp_mask=0x280da, order=0, oom_score_adj=0
[...]
[1880957.563400] Out of memory: Kill process 18694 (perl) score 246 or sacrifice child
[1880957.563408] Killed process 18694 (perl) total-vm:1972392kB, anon-rss:1953348kB, file-rss:0kB
[2320864.954447] TCP: Possible SYN flooding on port 7001. Dropping request.  Check SNMP counters.
```

> This views the last 10 system messages, if there are any. Look for errors that can cause performance issues. The example above includes the oom-killer, and TCP dropping a request.

`dmesg | tail`展示最后10条系统信息。
尝试在其中寻找造成性能问题的一场。
例子中展示了两个问题，一个是OOM导致的进程退出；另一个是TCP丢弃了一个链接（因为可能是SYN洪泛攻击）。

> Don’t miss this step! dmesg is always worth checking.

不要忽略这一步！`dmesg`值得查看一下。

# `vmstat 1`

```shell
$ vmstat 1
procs ---------memory---------- ---swap-- -----io---- -system-- ------cpu-----
 r  b swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st
34  0    0 200889792  73708 591828    0    0     0     5    6   10 96  1  3  0  0
32  0    0 200889920  73708 591860    0    0     0   592 13284 4282 98  1  1  0  0
32  0    0 200890112  73708 591860    0    0     0     0 9501 2154 99  1  0  0  0
32  0    0 200889568  73712 591856    0    0     0    48 11900 2459 99  0  0  0  0
32  0    0 200890208  73712 591860    0    0     0     0 15898 4840 98  1  1  0  0
^C
```

> Short for virtual memory stat, vmstat(8) is a commonly available tool (first created for BSD decades ago). It prints a summary of key server statistics on each line.

`vmstat`通常是开箱即用的工具，名字来源于虚拟内存状态的缩写，
该指令一行一行的打印出关键的系统统计数据。

> vmstat was run with an argument of 1, to print one second summaries. The first line of output (in this version of vmstat) has some columns that show the average since boot, instead of the previous second. For now, skip the first line, unless you want to learn and remember which column is which.

参数`1`表示每秒钟打印一次一秒内的统计。
第一行表示从启动以来的各项数据的平均值。
现在先跳过这一行，除非你想了解每一列代表了什么含义。

> Columns to check:

不同列表示不同的含义：

> r: Number of processes running on CPU and waiting for a turn. This provides a better signal than load averages for determining CPU saturation, as it does not include I/O. To interpret: an “r” value greater than the CPU count is saturation.

r: 正在运行或等待运行的进程的数量。因为这个数据不包含IO中的进程，所以更加适合用来判断CPU负载的饱和程度。
r值大于CPU数量时表示已经饱和。

> free: Free memory in kilobytes. If there are too many digits to count, you have enough free memory. The “free -m” command, included as command 7, better explains the state of free memory.

free: 空闲内存，KB。如果数字很长的话，那么表示有足够的空闲内存。
第七个指令`free -m`是一个更好的用来查看空闲内存的工具。

> si, so: Swap-ins and swap-outs. If these are non-zero, you’re out of memory.

si, so: 交换入、交换出。如果这两个值不为0，意味着内存是瓶颈。

> us, sy, id, wa, st: These are breakdowns of CPU time, on average across all CPUs. They are user time, system time (kernel), idle, wait I/O, and stolen time (by other guests, or with Xen, the guest’s own isolated driver domain).

us, sy, id, wa, st: 所有CPU时间的平均拆分细则，分别是用户态、内核态、空闲、等待IO、被（其他用户）窃取的。

> The CPU time breakdowns will confirm if the CPUs are busy, by adding user + system time. A constant degree of wait I/O points to a disk bottleneck; this is where the CPUs are idle, because tasks are blocked waiting for pending disk I/O. You can treat wait I/O as another form of CPU idle, one that gives a clue as to why they are idle.

通过用户态与内核态时间相加可以判断CPU是否繁忙。
等待IO的时间超过某些静态的值后表示磁盘是瓶颈，
在等待IO的时间中，CPU是空闲的，因为任务被阻塞在等待磁盘完成IO。
等待IO可以被认为是另一种空闲，只是有线索的空闲。
 
> System time is necessary for I/O processing. A high system time average, over 20%, can be interesting to explore further: perhaps the kernel is processing the I/O inefficiently.

处理IO会占用内核态时间。超过20%的内核态时间暗示了内核在处理IO时可能性能不够好。

> In the above example, CPU time is almost entirely in user-level, pointing to application level usage instead. The CPUs are also well over 90% utilized on average. This isn’t necessarily a problem; check for the degree of saturation using the “r” column.

上述例子中，CPU时间几乎全部在用户态，表示是应用级别消耗了CPU时间。
使用率超过了90%，但是并不意味着引起了切实的问题，检查负载的饱和程度可以通过"r"列。

# `mapstat -P ALL 1`

```shell
$ mpstat -P ALL 1
Linux 3.13.0-49-generic (titanclusters-xxxxx)  07/14/2015  _x86_64_ (32 CPU)

07:38:49 PM  CPU   %usr  %nice   %sys %iowait   %irq  %soft  %steal  %guest  %gnice  %idle
07:38:50 PM  all  98.47   0.00   0.75    0.00   0.00   0.00    0.00    0.00    0.00   0.78
07:38:50 PM    0  96.04   0.00   2.97    0.00   0.00   0.00    0.00    0.00    0.00   0.99
07:38:50 PM    1  97.00   0.00   1.00    0.00   0.00   0.00    0.00    0.00    0.00   2.00
07:38:50 PM    2  98.00   0.00   1.00    0.00   0.00   0.00    0.00    0.00    0.00   1.00
07:38:50 PM    3  96.97   0.00   0.00    0.00   0.00   0.00    0.00    0.00    0.00   3.03
[...]
```

> This command prints CPU time breakdowns per CPU, which can be used to check for an imbalance. A single hot CPU can be evidence of a single-threaded application.

`mapstat`按照CPU核心拆分CPU时间，可以用来检查负载不均的问题。
某个核心负载较高可能是一个单线程应用的问题的证据。

# `pidstat 1`

```shell
$ pidstat 1
Linux 3.13.0-49-generic (titanclusters-xxxxx)  07/14/2015    _x86_64_    (32 CPU)

07:41:02 PM   UID       PID    %usr %system  %guest    %CPU   CPU  Command
07:41:03 PM     0         9    0.00    0.94    0.00    0.94     1  rcuos/0
07:41:03 PM     0      4214    5.66    5.66    0.00   11.32    15  mesos-slave
07:41:03 PM     0      4354    0.94    0.94    0.00    1.89     8  java
07:41:03 PM     0      6521 1596.23    1.89    0.00 1598.11    27  java
07:41:03 PM     0      6564 1571.70    7.55    0.00 1579.25    28  java
07:41:03 PM 60004     60154    0.94    4.72    0.00    5.66     9  pidstat

07:41:03 PM   UID       PID    %usr %system  %guest    %CPU   CPU  Command
07:41:04 PM     0      4214    6.00    2.00    0.00    8.00    15  mesos-slave
07:41:04 PM     0      6521 1590.00    1.00    0.00 1591.00    27  java
07:41:04 PM     0      6564 1573.00   10.00    0.00 1583.00    28  java
07:41:04 PM   108      6718    1.00    0.00    0.00    1.00     0  snmp-pass
07:41:04 PM 60004     60154    1.00    4.00    0.00    5.00     9  pidstat
```

> Pidstat is a little like top’s per-process summary, but prints a rolling summary instead of clearing the screen. This can be useful for watching patterns over time, and also recording what you saw (copy-n-paste) into a record of your investigation.

`pidstat`有一点像`top`的进程版统计，但是输出一个滚动的总结而不是清空整个屏幕。
这对于查看时间维度上的指标变化和相应的查看、拷贝历史数据有帮助。

> The above example identifies two java processes as responsible for consuming CPU. The %CPU column is the total across all CPUs; 1591% shows that that java processes is consuming almost 16 CPUs.

例子中表明两个java进程是消耗CPU的主要进程。
`%CPU`列是所有核心的综合，1591%表示几乎16核。

# `iostat -xz 1`

```shell
$ iostat -xz 1
Linux 3.13.0-49-generic (titanclusters-xxxxx)  07/14/2015  _x86_64_ (32 CPU)

avg-cpu:  %user   %nice %system %iowait  %steal   %idle
          73.96    0.00    3.73    0.03    0.06   22.21

Device:   rrqm/s   wrqm/s     r/s     w/s    rkB/s    wkB/s avgrq-sz avgqu-sz   await r_await w_await  svctm  %util
xvda        0.00     0.23    0.21    0.18     4.52     2.08    34.37     0.00    9.98   13.80    5.42   2.44   0.09
xvdb        0.01     0.00    1.02    8.94   127.97   598.53   145.79     0.00    0.43    1.78    0.28   0.25   0.25
xvdc        0.01     0.00    1.02    8.86   127.79   595.94   146.50     0.00    0.45    1.82    0.30   0.27   0.26
dm-0        0.00     0.00    0.69    2.32    10.47    31.69    28.01     0.01    3.23    0.71    3.98   0.13   0.04
dm-1        0.00     0.00    0.00    0.94     0.01     3.78     8.00     0.33  345.84    0.04  346.81   0.01   0.00
dm-2        0.00     0.00    0.09    0.07     1.35     0.36    22.50     0.00    2.55    0.23    5.62   1.78   0.03
[...]
^C
```

> This is a great tool for understanding block devices (disks), both the workload applied and the resulting performance. Look for:

`iostat`是一个了解块设备（比如磁盘）的很好的工具，比如申请的负载与性能表现。

比如：

> - r/s, w/s, rkB/s, wkB/s: These are the delivered reads, writes, read Kbytes, and write Kbytes per second to the device. Use these for workload characterization. A performance problem may simply be due to an excessive load applied.
> - await: The average time for the I/O in milliseconds. This is the time that the application suffers, as it includes both time queued and time being serviced. Larger than expected average times can be an indicator of device saturation, or device problems.
> - avgqu-sz: The average number of requests issued to the device. Values greater than 1 can be evidence of saturation (although devices can typically operate on requests in parallel, especially virtual devices which front multiple back-end disks.)
> - %util: Device utilization. This is really a busy percent, showing the time each second that the device was doing work. Values greater than 60% typically lead to poor performance (which should be seen in await), although it depends on the device. Values close to 100% usually indicate saturation.

- r/s, w/s, rkB/s, wkB/s: 表示投递到设备的 读、写、读KB、写KB 请求按秒聚合数量。这些指标用来判断负载。性能问题的一个可能原因是太多的读写请求。
- await: 毫秒计数的IO平均时间。这个时间是应用切实体会到的IO耗时，包含等待时间和执行时间。大于预期的平均时间是设备饱和的一个线索，或者有可能是设备出现了问题。
- avgqu-sz: 被指派到设备的请求数的平均值。大于1表示设备已经饱和。尽管设备可以并行的处理请求，比如有多个后端磁盘的虚拟设备。
- %util: 设备使用率。是每秒中设备执行时间的比例。大于60%通常会引起性能问题（可以通过await来观察到），这个阈值还取决于设备。接近100%表示设备已经饱和。

> If the storage device is a logical disk device fronting many back-end disks, then 100% utilization may just mean that some I/O is being processed 100% of the time, however, the back-end disks may be far from saturated, and may be able to handle much more work.

对于存储设备是一个有多个磁盘后端的前端虚拟设备，100%的使用率可能只是表示一些IO操作占用了所有的时间，但是后端磁盘还远远没有饱和，也许可以继续执行更多的工作。

> Bear in mind that poor performing disk I/O isn’t necessarily an application issue. Many techniques are typically used to perform I/O asynchronously, so that the application doesn’t block and suffer the latency directly (e.g., read-ahead for reads, and buffering for writes).

谨记磁盘IO性能较差并不意味着一个切实的问题。许多技术经常被用来完成异步IO操作，所以应用并不会阻塞或者直接影响到延迟。
比如预读或者缓存。

# `free -m`

```shell
$ free -m
             total       used       free     shared    buffers     cached
Mem:        245998      24545     221453         83         59        541
-/+ buffers/cache:      23944     222053
Swap:            0          0          0
```

> The right two columns show:
> - buffers: For the buffer cache, used for block device I/O.
> - cached: For the page cache, used by file systems.

最右的两列分别表示：

- buffers: 用于块设备IO。
- cached: 页缓存，用于文件系统的。

> We just want to check that these aren’t near-zero in size, which can lead to higher disk I/O (confirm using iostat), and worse performance. The above example looks fine, with many Mbytes in each.

这两个数据只要不是近乎0即可，如果接近0的话，有可能会导致更多的磁盘IO（可以通过`iostat`来确定）和更差的性能。
例子看起来没什么问题，都有数MB的大小。

> The “-/+ buffers/cache” provides less confusing values for used and free memory. Linux uses free memory for the caches, but can reclaim it quickly if applications need it. So in a way the cached memory should be included in the free memory column, which this line does. There’s even a website, linuxatemyram, about this confusion.

"-/+ buffers/cache"提供了更加不会被误解的使用的或空闲的内存数量。
Linux使用空闲内存作为缓存，当需要时可以快速的被应用使用。
通常来说，用于缓存的内存和空闲内存可以划等号，这一行就把用于缓存的内存当做了空闲内存。
甚至有一个网站[linuxatemyram](https://www.linuxatemyram.com/)来解释这个令人迷惑的情况。

> It can be additionally confusing if ZFS on Linux is used, as we do for some services, as ZFS has its own file system cache that isn’t reflected properly by the free -m columns. It can appear that the system is low on free memory, when that memory is in fact available for use from the ZFS cache as needed.

当使用ZFS时会有另外的令人迷惑的情况，因为ZFS有自己的文件系统缓存，不会被`free -m`正确的反应。
看起来系统短缺空闲内存，但是有一部分被标为使用了的内存会在需要时从ZFS的文件系统缓存中释放出来。

# `sar -n DEV 1`

```shell
$ sar -n DEV 1
Linux 3.13.0-49-generic (titanclusters-xxxxx)  07/14/2015     _x86_64_    (32 CPU)

12:16:48 AM     IFACE   rxpck/s   txpck/s    rxkB/s    txkB/s   rxcmp/s   txcmp/s  rxmcst/s   %ifutil
12:16:49 AM      eth0  18763.00   5032.00  20686.42    478.30      0.00      0.00      0.00      0.00
12:16:49 AM        lo     14.00     14.00      1.36      1.36      0.00      0.00      0.00      0.00
12:16:49 AM   docker0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00

12:16:49 AM     IFACE   rxpck/s   txpck/s    rxkB/s    txkB/s   rxcmp/s   txcmp/s  rxmcst/s   %ifutil
12:16:50 AM      eth0  19763.00   5101.00  21999.10    482.56      0.00      0.00      0.00      0.00
12:16:50 AM        lo     20.00     20.00      3.25      3.25      0.00      0.00      0.00      0.00
12:16:50 AM   docker0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00
^C
```

> Use this tool to check network interface throughput: rxkB/s and txkB/s, as a measure of workload, and also to check if any limit has been reached. In the above example, eth0 receive is reaching 22 Mbytes/s, which is 176 Mbits/sec (well under, say, a 1 Gbit/sec limit).
> This version also has %ifutil for device utilization (max of both directions for full duplex), which is something we also use Brendan’s nicstat tool to measure. And like with nicstat, this is hard to get right, and seems to not be working in this example (0.00).

`sar`可以用来检查网络接口的吞吐量：rxkB/s和txkB/s。
吞吐量可以用来衡量负载，并且可以来检查是否有触达某些限制。
例子中，`eth0`接收22MB/s，换算为176Mbits/s。

# `sar -n TCP,ETCP 1`

```shell
$ sar -n TCP,ETCP 1
Linux 3.13.0-49-generic (titanclusters-xxxxx)  07/14/2015    _x86_64_    (32 CPU)

12:17:19 AM  active/s passive/s    iseg/s    oseg/s
12:17:20 AM      1.00      0.00  10233.00  18846.00

12:17:19 AM  atmptf/s  estres/s retrans/s isegerr/s   orsts/s
12:17:20 AM      0.00      0.00      0.00      0.00      0.00

12:17:20 AM  active/s passive/s    iseg/s    oseg/s
12:17:21 AM      1.00      0.00   8359.00   6039.00

12:17:20 AM  atmptf/s  estres/s retrans/s isegerr/s   orsts/s
12:17:21 AM      0.00      0.00      0.00      0.00      0.00
^C
```

> This is a summarized view of some key TCP metrics. These include:
> active/s: Number of locally-initiated TCP connections per second (e.g., via connect()).
> passive/s: Number of remotely-initiated TCP connections per second (e.g., via accept()).
> retrans/s: Number of TCP retransmits per second.

这些信息是一些关键的TCP指标，包括：

- active/s: 本地初始化的TCP链接数量。
- passive/s: 远端初始化的TCP链接数量。
- retrans/s: TCP重传的数量。

> The active and passive counts are often useful as a rough measure of server load: number of new accepted connections (passive), and number of downstream connections (active). It might help to think of active as outbound, and passive as inbound, but this isn’t strictly true (e.g., consider a localhost to localhost connection).

主动或被动建立链接的数量可以粗略的用于衡量系统的负载。

> Retransmits are a sign of a network or server issue; it may be an unreliable network (e.g., the public Internet), or it may be due a server being overloaded and dropping packets. The example above shows just one new TCP connection per-second.

重传是网络或系统出现问题的症状，可能是因为不可靠的网络、某个服务器过载并引起了丢包。

例子中只有每秒一个新链接。

# `top`

> The top command includes many of the metrics we checked earlier. It can be handy to run it to see if anything looks wildly different from the earlier commands, which would indicate that load is variable.

`top`指令包括了许多之前看到过的指标。
可以通过`top`来总览并与之前的结果进行比较。

> A downside to top is that it is harder to see patterns over time, which may be more clear in tools like vmstat and pidstat, which provide rolling output. Evidence of intermittent issues can also be lost if you don’t pause the output quick enough (Ctrl-S to pause, Ctrl-Q to continue), and the screen clears.

`top`的缺点是无法进行时间维度上的对比。
问题的证据如果没有及时的暂停很容易就会丢失。

# Follow-on Analysis

> There are many more commands and methodologies you can apply to drill deeper. See Brendan’s Linux Performance Tools tutorial from Velocity 2015, which works through over 40 commands, covering observability, benchmarking, tuning, static performance tuning, profiling, and tracing.

还有很多指令可以用来排查性能问题，可以参考《Linux Performance Tools tutorial》。

Tackling system reliability and performance problems at web scale is one of our passions. If you would like to join us in tackling these kinds of challenges we are hiring!

# Reference

- 关于系统负载的考据 [Linux Load Averages: Solving the Mystery](https://www.brendangregg.com/blog/2017-08-08/linux-load-averages.html)

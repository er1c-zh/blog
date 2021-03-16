---
title: "如何用脚本登录远程机器——expect用法"
date: 2021-03-16T23:14:06+08:00
draft: false
tags:
    - linux
    - command
    - how
---

expect可以通过写命令来处理交互式的程序。

<!--more-->

# 用法

这里通过一个例子来说明：

```expect
#!/usr/bin/expect -d

puts "create cluster start"

spawn ./redis/src/redis-cli --cluster create 127.0.0.1:7000 127.0.0.1:7001 \
        127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
        --cluster-replicas 1

expect -re yes
send "yes\n"

puts ""
puts "create cluster done"

wait
```

第一行指明由`expect`执行。
特别的，`-d`指明开启调试模式，`expect`在运行的时候会输出详细的数据。

第二行 `puts "create cluster start"`
利用`puts`指令向stdout输出了一行。

第三行`spawn commond_to_run`
会启动一个子进程来执行指令。

第四行`expect -re yes`表示正则匹配 *(`-re`)* `yes`。

如果匹配到了，会执行后续的`send "yes\n"`，
表示向程序发送`yes\n` *(回车)*。

这一部分有另一种写法：

```expect
expect {
    -re yes {
        send "yes\n"
        exp_continue
    }
    -re another_val_to_match {
        send "hello expect\n"
    }
}
```

这种写法在需要多个分支匹配、循环匹配等场景更加的有效。
其中`exp_continue`表示需要回到`expect`的开头继续匹配。

最后`wait`表示脚本等待子进程返回EOF。

## 一些其他的指令

- `interact` 可以将交互的控制权交给用户
- `exit` 退出脚本
- `close` 退出进程

# 例子

## 输入用户名和密码登陆

```expect
#!/usr/bin/expect
 
spawn ssh root@127.0.0.1
 
expect {
    -re "Are you sure you want to continue connecting" {
        send "yes\n"
        exp_continue
    }   
    -re "password" {
        send "password_here\n"
    }   
}

interact
```

# 遇到过的问题

1. 执行了之后好像没有效果？

    这发生在我尝试在docker镜像里利用`redis-cli`启动集群的时候，
    现象是程序一切正常，然后调试信息也显示正常的写入了期望的数据，
    但是redis的命令没有执行。

    后来发现，`expect`如果执行完毕，会直接退出，
    执行任务的命令如果没有完成，就会被终止。

    解决的方案是在需要等待子进程时，添加`wait`指令，
    就会等待到子进程完成。

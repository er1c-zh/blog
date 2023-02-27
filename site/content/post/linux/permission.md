---
title: "Linux用户权限"
date: 2023-02-27T14:51:28+08:00
draft: false
tags:
    - linux
    - command
    - how
    - memo
---

linux通过User和Group机制提供简单的权限管理。

<!--more-->

# 权限检查的流程

在用户执行操作时，
系统通过读取用户拥有的权限与文件的权限配置，
判断是否允许执行相应的操作。

# 用户与用户组

用户(user)表示一个具体的用户账号。

用户组是用户的集合，用来集中管理权限。

每个账号都有一个primary group，可能拥有零个或多个secondary group。
区别在于primary group作为“默认的”组，如创建文件时，会将用户的primary group作为文件的owner group。

# 文件与文件夹的权限配置

权限有读写执行三种，分别用`R` `W` `X`表示。

文件和文件夹将权限存储在一个位图上，可以通过`ls -l`进行查看。

> -rw-r--r--  1 eric eric     3792 Feb 24 11:49 .bashrc
> drwx------  6 eric eric     4096 Feb 27 08:58 .cache

其中`-rw-r--r--`表示了文件`.bashrc`的权限配置：

- 第一个字符 `-` file type。

- 第一组权限 `rw-` 表示有读、写权限，没有执行权限；第二、三组权限 `r--` 表示只有读权限。

- 第一组权限表示文件拥有者的权限，第二组表示拥有该文件的组(group)的权限，第三组表示其他用户的权限。

`drwx------` 表示`.cache`是一个文件夹(directory)，其他类似`.bashrc`。

## 权限在文件上的表现

- `r` 可以读取文件
- `w` 可以修改文件
- `x` 可以执行文件

## 权限在文件夹上的表现

- `r` 可以查看该文件夹的内容（如文件）。
- `w` 可以在文件夹中增删文件。
- `x` 可以“访问、使用”该文件夹(access)，可以理解成拥有`x`权限是对文件夹进行任何操作（包括读、写）的前提。

# 一些常用操作

## 用户与组

```shell
# 查看当前执行的user
whoami
# 查看当前user及group
id

# 查看有哪些user
cat /etc/passwd
# 查看有哪些group
cat /etc/group

# 用户user_name加入组another_group
usermod -a -G another_group user_name

# 新建用户username
useradd username
# 修改username的密码
passwd username
```

## 修改文件的权限配置

```shell
# a(ll) u(ser) g(roup) o(thers)
# r(ead) w(rite) x(execute)
chmod a+x ./file # 给user && group && others增加执行权限
chmod ug+rwx ./file # 给user && group增加读写执行权限

# --- --- ---
# rwx rwx rwx
# 111 100 000
#  7   4   0 
chmod 740 ./file # user有rwx权限，group中的用户有读权限，others没有任何权限
```

## 操作用户

# Reference

- [archlinux users_and_groups](https://wiki.archlinux.org/title/users_and_groups)
- [lindoe linux-users-and-groups](https://www.linode.com/docs/guides/linux-users-and-groups/)
- [Linux file permissions explained](https://www.redhat.com/sysadmin/linux-file-permissions-explained)

---
title: "mysql相关操作备忘录"
date: 2020-10-28T19:02:37+08:00
draft: false
tags:
    - mysql
    - memo
    - how-to
---

- 如何在Ubuntu 20上使用apt安装低版本mysql？ mysql5.7 mysql5.6
- 启动mysql失败的时候如何排查？

# 如何在Ubuntu 20上使用apt安装低版本mysql？ mysql5.7 mysql5.6

*2020.10.27有效*

1. 下载[mysql apt配置程序mysql-apt-config_0.8.15-1_all.deb](https://repo.mysql.com/mysql-apt-config_0.8.15-1_all.deb)。
1. `dpkg` 安装。
    1. 如果Ubuntu版本没有被自动识别，可以考虑在[该页面](https://repo.mysql.com/)自行寻找更新的mysql-apt-config。
    1. 安装时会自动弹出配置选项，选择需要的版本。
    1. 该程序会在apt的源列表中添加包含上面选择的版本的源。
1. `apt update`更新。
1. 安装mysql，通过参数指定需要的版本。
    ```shell
    sudo apt install mysql-client=5.7.30-1debian8
    sudo apt install mysql-community-server=5.7.30-1debian8
    sudo apt install mysql-server=5.7.30-1debian8
    ```

# 启动mysql失败的时候如何排查？

关键字：错误日志 error log

不论是以`service`或`systemctl`的形式来启动的，mysql总会留有启动的错误日志。

通常可以考虑`/val/log/mysql/*`下面。

## 遇到过的错误

1. 因为卸载不完全等原因，导致数据文件路径*(`可以考虑/var/lib/mysql`)*下残留数据文件。
新安装的mysql的innodb加载数据，因为数据不符合预期导致启动终止。清理相关文件即可。






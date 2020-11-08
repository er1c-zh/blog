---
title: "MySQL相关备忘录"
date: 2020-10-28T19:02:37+08:00
draft: false
tags:
    - MySQL
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

# EXPLAIN 备忘
[5.7 explain](https://dev.mysql.com/doc/refman/5.7/en/explain-output.html)

- id
- select_type
  - 查询的类型，如简单查询
- table
  - 表名
- partitions
- type
  - [mysql 5.7 explain-join-types](https://dev.mysql.com/doc/refman/5.7/en/explain-output.html#explain-join-types)
  - 访问的形式 以下由好到坏
  - system 很快
  - const 表示最多从该表中获取一行，并且在开始查询的时候就能获得到结果，被其他查询可以当作常量
  - eq_ref 前一个查询的任意一行，在该表中最多返回一行
  - ref 相比于eq_ref，从该表中可能返回多行
  - fulltext 使用FULLTEXT索引
  - ref_or_null 类似ref，只是多扫描NULL的情况
  - index_merge 表示索引合并优化开启
  - unique_subquery
  - index_subquery
  - range 扫描部分索引
  - index 遍历索引
  - ALL 全表扫描
- possible_keys
  - 可能选择的索引
- key
  - 最终选择的索引
- key_len
  - 选择的索引使用的长度 字节数
- ref
- rows
  - 要扫描的行数
- filtered
- Extra 其他相关的信息，以下为较重要的
  - [5.7 explain-extra-information](https://dev.mysql.com/doc/refman/5.7/en/explain-output.html#explain-extra-information)
  - Using filesort 需要额外的排序来满足需求 
  - Using index 覆盖索引
  - Using index condition 索引下推
  - Using index for group-by 可以通过索引来满足`GROUP BY`或`DISTINCT`查询
  - Using temporary 需要创建临时表来处理结果，通常因为查询存在由不同的列`GROUP BY`或`ORDER BY`
  - Using where 表示在引擎获取到数据之后，系统需要再次过滤。这暗示了可以新建索引来优化。

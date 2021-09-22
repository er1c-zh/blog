---
title: "Docker备忘录"
date: 2021-09-22T08:52:37+08:00
draft: false
tag:
    - docker
    - memo
    - how-to
---

<!--more-->

# Docker

# 遇到的问题及解决方案

## 如何实现重启docker之后容器自动启动，固定连入的`network`的ip

*20210922/centos7/docker_20.10.8_1.41*

问题的场景是我有连入同一个`network`的两个容器，其中一个是MySQL实例。
希望能够在NAS重启时自动启动这两个容器。

首先通过`systemctl enable docker`等方法令docker开机自启动。

然后参考
[自动启动容器](https://docs.docker.com/config/containers/start-containers-automatically/)
来配置容器restart策略来实现自动启动。

```shell
# 将已经存在的名称为 mysql 的容器的restart策略设置为 always
docker update --restart always mysql 
# 创建时设置为 unless-stopped
docker run -d --restart unless-stopped redis
```

其中restart策略的备选项可以参考上述的
[文档](https://docs.docker.com/config/containers/start-containers-automatically/)。

重启之后发现mysql连入`network`的ip发生了变动。

这可以在连入`network`时配置容器的固定ip。

```shell
# 如果已经连接到了 network 可以先断开连接
docker network disconnect freshrss-network mysql
# 将 mysql 容器 连入 freshrss-network 使用指定的ip 172.18.0.3
docker network connect --ip=172.18.0.3 freshrss-network mysql
```


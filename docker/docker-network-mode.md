# docker的网络模式和跨主机通信

[TOC]

## docker网络模式 network-mode

### bridge模式

docker中， `bridge` 模式使用一个软件网桥实现连接到同一个网桥中的容器可以互相连通，未连接到同一网桥的容器互相隔离。

当docker第一次启动的时候，就会创建一个默认的桥接网络。新运行的容器默认会使用 `bridge模式` 并连接到这个 `桥接网络` 上。

> $ sudo ifconfig
> docker0   Link encap:Ethernet  HWaddr 02:42:da:2e:c5:ed
>           inet addr:172.16.11.65  Bcast:0.0.0.0  Mask:255.255.255.224
>           UP BROADCAST MULTICAST  MTU:1500  Metric:1
>           RX packets:0 errors:0 dropped:0 overruns:0 frame:0
>           TX packets:0 errors:0 dropped:0 overruns:0 carrier:0
>           collisions:0 txqueuelen:0
>           RX bytes:0 (0.0 B)  TX bytes:0 (0.0 B)

#### 自定义bridge network

[自定义桥接网络](https://docs.docker.com/network/bridge/#manage-a-user-defined-bridge)

### host模式

简单来说就是直接使用宿主机网络空间

```shell
docker run --name=container-name --net=host image
```

### overlay

实现了跨主机 `容器` 通信

### macvlan

虚拟一个有不同mac地址的网卡在链路层进行转发

### none

不配置容器的网络

## 跨主机通信

> 一分钟看懂Docker的网络模式和跨主机通信 https://www.cnblogs.com/yy-cxd/p/6553624.html

### 直接路由方式

### 桥接方式（如*pipework*）

### Overlay隧道方式（如*flannel*、*ovs+gre*）

## 参考

- [docker-document-network v18.03](https://docs.docker.com/network/)
- [一分钟看懂Docker的网络模式和跨主机通信](https://www.cnblogs.com/yy-cxd/p/6553624.html)
- [使用Docker的macvlan为容器提供桥接网络及跨主机通讯](https://www.cnblogs.com/eshizhan/p/7255565.html)
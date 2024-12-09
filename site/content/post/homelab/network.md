---
title: "无复杂组件的透明代理解决方案"
date: 2024-12-09T16:57:12+08:00
draft: false
tags:
    - homelab
---

之前使用嗅探的方式处理 https 请求的透明代理表现不好，
因为重置链接，导致日常使用中初次建链的延迟巨大。
以及所有的流量都要走到用户态处理导致 udp 代理实际上不可用。

故改为如下方案：

利用防火墙将根据 IP ( MAC / package meta / etc ) 将需要走代理的流量劫持到代理软件，其他的放行。

features:

- IPv6 / UDP / 路由器本机
- 相对较低的性能消耗
- 代理链不可用不影响国内访问
- 思路简单，理解容易

核心原理：

1. 为解决 dns 污染 / 代理就近访问 等问题： dns 根据域名分别请求内外的 dns 服务器， 此时 dns 请求会根据防火墙规则被分流。
1. 防火墙根据需要劫持需要代理的流量到代理软件。

pros:

- 无需解包嗅探 / 直连流量不走用户态
- 代理软件逻辑简单便于升级 / 替换
- 基于防火墙控制流量十分灵活

cons:

- 与 taliscale 冲突，路由器上的改造之前的 tailscale 在改造完成后不可用，暂时没有细究原因。

# dns

mosdns

- 支持根据规则指定 dns 服务器
- lightweight

# nftables setting

- 防止无限循环： add package meta mark
- IPv6 / udp 支持： tproxy
- 路由器本机流量代理： package mark 之后会重定向到 Forward 链，与局域网客户端处理一致。

```uci
// append to /etc/config/network

config route6
        option interface 'loopback'
        option type 'local'
        option table '106'
        option target '::/0'

config rule
        option mark '0x01'
        option lookup '100'

config route
        option interface 'loopback'
        option type 'local'
        option table '100'
        option target '0.0.0.0/0'

config rule6
        option mark '0x01'
        option lookup '106'

config interface 'loopback6'
        option device 'lo'
        option proto 'static'
        list ip6addr '::1'
```

```nft
# china ipv4
include "/etc/nftables.d/ipv4_cn.ips"
# china ipv6
include "/etc/nftables.d/ipv6_cn.ips"

set cn_v4 {
    type ipv4_addr
    flags interval
    elements = $chn_v4
}

set cn_v6 {
    type ipv6_addr
    flags interval
    elements = $chn_v6
}

set private {
    type ipv4_addr
    flags interval
    elements = { 0.0.0.0/8, 10.0.0.0/8, 100.64.0.0/10, 127.0.0.0/8,
        169.254.0.0/16, 172.16.0.0/12, 192.0.0.0/24, 192.0.2.0/24,
        192.88.99.0/24, 192.168.0.0/16, 198.18.0.0/15, 198.51.100.0/24,
        203.0.113.0/24, 224.0.0.0/4, 240.0.0.0/4 }
}

set private6 {
    type ipv6_addr
    flags interval
    elements = { ::/128, ::1/128, ::ffff:0:0/96, 100::/64,
        64:ff9b::/96, 2001::/32, 2001:10::/28, 2001:20::/28,
        2001:db8::/32, 2002::/16, fc00::/7, fe80::/10,
        ff00::/8 }
}

set selfproxy {
    type ipv4_addr
    flags interval
    elements = { 10.0.0.0 }
}

set selfproxy6 {
    type ipv6_addr
    flags interval
    elements = { ff00:: }
}

chain direct {
    # 非 tcp udp 忽略
    meta l4proto != { tcp, udp } counter accept
    # 处理过的忽略
    meta mark 255 accept

    # 内网忽略
    ip daddr @private accept comment "ip4 private"
    ip6 daddr @private6 accept comment "ip6 private"

    # 国内的 ip 忽略
    ip daddr @cn_v4 accept comment "ip4 cn"
    ip6 daddr @cn_v6 accept comment "ip6 cn"

    # 需要直连的 ip
    ip daddr @selfproxy accept comment "ip4 proxy"
    ip6 daddr @selfproxy6 accept comment "ip6 proxy"

    # 根据 mac 地址忽略
    ether saddr D2:C6:3E:00:00:00 accept comment "bt"
}

# prerouting 链
chain tp_prerouting {
    type filter hook prerouting priority filter; policy accept;

    jump direct

    # 转发需要走代理的流量到代理
    meta l4proto { tcp, udp } mark set 1 tproxy ip to :6000 accept
    meta l4proto { tcp, udp } mark set 1 tproxy ip6 to :6000 accept
}

chain tp_output {
    type route hook output priority filter; policy accept;

    jump direct

    # 标记后会被路由到 prerouting
    meta l4proto { tcp, udp } meta mark set 1 accept
}
```

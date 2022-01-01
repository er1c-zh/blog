# hugo相关

## “系列”的用法

1. 在`tags`中增加系列的id。
1. 在文章需要展示系列列表的地方增加
    
    ```hugo
        {{% serial_index serial_id %}}
    ```

1. 通过在Front-Matter中增加`order`，整数，小的在前大的在后，来明确顺序，。

## 截断摘要

```hugo
<!--more-->
```

## ditaa

使用代码段，语言使用ditaa标记，开头标记ditaa与配置。

````
```ditaa
ditaa
+-----+
| img |
+-----+
```
````

# TODO list

- blog维护相关
    - [x] 自动触发gitalk生成文章的issue
    - [ ] 截图上传到onedrive的跨平台应用
- 技术学习
    - [x] unicode && utf-8
    - [ ] 弄清md5 sha-1等加密算法
    - [ ] react/vue的入门，结合截图上传工具来实践以下
    - [x] git rebase
    - [ ] redis
        - [x] 单机命令实现
        - [x] 持久化
        - [ ] 集群
    - [ ] 了解mongodb
    - [ ] 学习go的runtime的实现，结合linux来看
        - [x] chan
        - [ ] select
        - [ ] 调度 sudog
    - [ ] 了解一下rust
    - [ ] 简单研究nginx的设计思想
    - [ ] 简单了解innodb的实现
    - [ ] gossip算法
    - [ ] ruby
        - [x] 基础语法
        - [ ] 函数式编程
    - [ ] 编译原理的基础部分
    - [ ] 多人编辑专题
        - [x] CRDT
        - [ ] OT
        - [x] diff
        - [ ] 版本向量
    - [ ] 监控系统
        - [ ] 整理一下 监控的哲学
        - [ ] Prometheuse简单实现
    - [ ] HTTP/1.1协议的笔记
        - [x] rfc7230 "Message Syntax and Routing" (this document)
        - [ ] rfc7231 "Semantics and Content" [RFC7231]
        - [ ] rfc7232 "Conditional Requests" [RFC7232]
        - [ ] rfc7233 "Range Requests" [RFC7233]
        - [ ] rfc7234 "Caching" [RFC7234]
        - [ ] rfc7235 "Authentication" [RFC7235]
    - [ ] HTTP/2
- [ ] 尝试为一些小的开源项目提供代码
- [ ] 构建自己的dotfiles
    - [x] mac
    - [ ] linux
- books wait read
    - [ ] hackers delight
    - [ ] ~https://swtch.com/~rsc/regexp/regexp1.html~
    - [ ] 两次全球大危机的比较研究
    - [x] DDIA
    - [x] Linux内核设计与实现
    - [ ] 编译原理
    - [ ] SQL与关系数据库理论
    - [ ] 应用密码学：协议、算法与C源程序（原书第2版）

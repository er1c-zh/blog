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


# todo list

- [x] unicode && utf-8
- [ ] 弄清md5 sha-1等加密算法
- blog维护相关
    - [x] 自动触发gitalk生成文章的issue
    - [ ] 截图上传到onedrive的跨平台应用
- 技术学习2020
    - [x] redis的命令实现
    - [ ] react/vue的入门，结合截图上传工具来实践以下
    - [x] git rebase
- 技术学习2021
    - [ ] redis的持久化、集群
        - [x] 持久化
        - [ ] 集群
    - [ ] 了解mongodb
    - [ ] 学习go的runtime的实现，结合linux来看
    - [ ] 了解一下rust
    - [ ] 简单研究nginx的设计思想
    - [ ] 简单了解innodb的实现
    - [ ] gossip算法
- [ ] 尝试为一些小的开源项目提供代码

# books wait read

- hackers delight
- ~https://swtch.com/~rsc/regexp/regexp1.html~
- 两次全球大危机的比较研究
- [x] DDIA
- [ ] SQL与关系数据库理论
- [ ] 应用密码学：协议、算法与C源程序（原书第2版）


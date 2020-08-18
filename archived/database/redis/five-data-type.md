---
title: redis中的数据类型
categories:
  - database
  - redis
---

## 存储模型

redis的存储都基于kv对进行存储、搜索。

### key

key的数据类型是一个 **字符串**。具有以下特点

- 二进制安全
- key的大小[0, 512MB]

使用中需要注意一些问题：

- key的长度不能太长
  - 会消耗内存
  - 在进行key的比较时，消耗较大
- key需要保持足够的可读性
- key的模式应该尽量保持一致
  - 比如一种模式是用冒号进行切分不同层次的概念，用连接号链接一个层次中的多个单词

例如，user:{user_id}:followers比u{user_id}f要很多，一是可读性好很多，二是用于增加可读性的辅助字符与user_id来说，并不会影响太多的内存

## 数据类型

### 字符串 strings

字符串是redis最基本、最简单的数据类型。对于字符串类型的值，redis支持将值看作string、bitmap、整数和浮点数，并为各个形式的字符串提供了一些特殊的操作。

#### common

- GET/SET 查询/设置
  - GET key
  - SET key value
- MGET/MSET 批量设置
  - MGET key [key ...]
  - MSET key value [key value ...]
- SETNX/MSETNX 当key不存在时设置
  - SETNX key value
  - MSETNX key value [key value ...]
- SETEX/PSETEX 设置key的值同时设置过期时间
  - SETEX key seconds value
  - PSETEX key milliseconds value
- GETSET 设置key的值，并把原有的值返回

#### string

- APPEND key value
- STRLEN key
- GETRANGE/SETRANGE
  - GETRANGE key start end
  - SETRANGE key offset value 从[offset:]开始覆写

### 整数

- INCR/DECR 原子加1/减1
  - INCR key
  - DECR key
- INCRBY/DECRBY
  - INCRBY key increment
  - DECRBY key decrement

#### bitmap

- BITCOUNT key [start end]
  - 从start字节，到end字节，闭区间中的1的bit数
- BITFIELD (from 3.2.0)
- BITOP operation destkey [key ...]
  - 支持四种位操作，与或异或非
  - 多个src的值的长度不一致时，短的高位补0
  - BITOP AND dest [src1 ...]
  - BITOP OR dest [src1 ...]
  - BITOP XOR dest [src1 ...]
  - BITOP NOT dest src
- BITPOS key bit [start] [end]
  - 返回第一个值为{bit}的位的index
- GETBIT/SETBIT设置/读取指定位的值
  - GETBIT key offset
  - SETBIT key offset value
    - 返回值为offset原来的值

#### 浮点数

- INCRBYFLOAT key increment
  - 参数可正可负

### 有序列表 lists

认为list的左边为头，右边为尾。

#### 列表

- LLEN key 获得列表的长度

#### 对元素读

- LINDEX key index 从列表key中获得下标index的元素
- LRANGE key start stop
  - 从列表中获得一个节点序列，闭区间
  - 任何下标的异常，如start大于stop、stop大于实际的长度，都不会报错，会提供符合直觉的结果

#### 对元素修改

- LSET key index value
  - 将列表key中index的元素设置成value
- LINSERT key BEFORE|AFTER pivot value
  - 在列表key中，寻找到值为pivot的节点，在BEFORE|AFTER插入value节点
- LREM key count value
  - 移除列表key中的等于value的元素
  - 如果count=0，移除所有的等于value的元素
  - 如果count<0，移除从尾到头第`abs(count)`个等于value的元素
  - 如果count>0，移除从头到尾第count个等于value的元素
- LTRIM key start stop
  - 只保留[start, stop]的元素

#### 队列

- LPOP/RPOP key
  - 从队列的头/尾pop一个节点
- LPUSH/RPUSH key value [value ...]
  - 从头/尾插入若干节点
- LPUSHX/RPUSHX key value
  - 只有队列存在时，插入节点
- RPOPLPUSH source destination
  - 从src尾部pop一个元素，push到dest头部
  - 返回值为元素

#### 阻塞

- BLPOP/BRPOP/BRPOPLPUSH 

### 集合 sets

### 有序集合 sorted sets

### 哈希 hashes

### HyperLogLogs

### Streams




---
title: "位图的基本实现"
date: 2020-11-22T00:27:36+08:00
draft: true
tags:
  - redis
  - what
  - how-redis-work
order: 2
---

redis位图相关指令的具体实现。

<!--more-->

{{% serial_index how-redis-work %}}

分类按照命令表的flag来确定，字符串类的是包含`@bitmap`标签的指令。

*基于redis 6.0*

# `SETBIT`/`GETBIT`

写入复杂一点：
- 首先通过`getBitOffsetFromArgument`和`getLongFromObjectOrReply`来获取偏移量和要设置的值。
- 然后通过`lookupStringForBitCommand`获取目标对象，函数中首先确定是否存在，相应的补全或新建出来需要长度的字符串，用零值填充。
- 通过offset来计算出来第几个字符和在字符中的偏移量，计算修改后的字符值，设置。

  ```c
  /* Get current values */
  byte = bitoffset >> 3; // 除8 获得要修改的位在那个字符上
  byteval = ((uint8_t*)o->ptr)[byte]; 
  bit = 7 - (bitoffset & 0x7); // (0x7 -> 111b) 获得8位里面的偏移量，然后反转顺序
  bitval = byteval & (1 << bit); // 存储之前的值

  /* Update byte with new bit value and return original value */
  byteval &= ~(1 << bit); // 清除那一位
  byteval |= ((on & 0x1) << bit); // 根据入参来设置
  ((uint8_t*)o->ptr)[byte] = byteval; // 保存
  ```
- 触发事件，写入结果，之前是1返回1，是0返回0。

读取比较简单，通过`getBitOffsetFromArgument`获取对应的值，
然后使用与设置的相同的方法，定位到对应的位置，返回结果。
特别的，如果位数超过长度，那么返回0。

# `BITFIELD`/`BITFIELD_RO`

两个命令最终落到`bitfieldGeneric`上完成执行。

首先是一个循环，遍历了所有的参数：
- 获取指令，如果不合规返回错误
- 获取参数，如果不合规返回错误
- 保存指令和参数到`ops`，是一个`bitfieldOp`的数组
  ```c
  struct bitfieldOp {
    uint64_t offset;    /* Bitfield offset. */
    int64_t i64;        /* Increment amount (INCRBY) or SET value */
    int opcode;         /* Operation id. */
    int owtype;         /* Overflow type to use. */
    int bits;           /* Integer bitfield bits width. */
    int sign;           /* True if signed, otherwise unsigned op. */
  };
  ```

随后用`lookupStringForBitCommand`获取要操作的键值对。

然后是一个循环，遍历`ops`：
- 如果是写操作(`SET`/`INCRBY`)，按照有符号和无符号进行类似的操作：
  - 查找出原有的数值
  - 对`SET`/`INCRBY`采用对应的方式，计算出新的值
  - 写入数值
  - 返回结果
- 对于读`GET`，实现比较简单。比较有意思的是，首先会拷贝9字节（72位）的数据到本地缓存，便于处理。

最后触发相应的事件，返回结果。

# `BITOP`

# `BITCOUNT`

# `BITPOS`

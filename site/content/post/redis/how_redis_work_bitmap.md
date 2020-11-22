---
title: "位图的基本实现"
date: 2020-11-22T00:27:36+08:00
draft: false
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

**需要注意跨字节的计算，如果对齐时，采用的是大端序。**

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

*时间复杂度O(N)*

1. 首先判断采取哪种运算，有意思的地方是，先比较第一位，符合后再比较了剩下的字符。
1. 随后获取每一个键对应的值，作简单的检查后，暂存起来
1. 如果有不位空的运算的值，那么进行运算。（对于快慢路径的分离，go中运用的也很多）
  1. 判断如果对齐、长度超过4*long位、要操作的键小于等于16个，那么先走一个快速路径
    1. 依次遍历所有值，计算4*long位的结果
    1. 更新结果的指针位置、剩余要计算的长度，重新回到快速路径检查
  1. 剩下还有一些需要继续计算的数据，比如值的长度更长、位数不足4*long等。循环按照字节递增，直到超过，`maxlen`（字节数）。超过目标位数来进行位运算不会影响结果。
    1. 依次遍历每个值
      1. 检查每个值的长度是否合规
      1. 根据不同的运算，采取对应的操作
      1. 设置字节到应有的位置
1. 如果结果为空，那么尝试删除，并触发时间；否则通过`setKey`设置结果，并触发事件。
1. 返回结果的长度（字节数，不足一个字节的补足）

# `BITCOUNT`

首先获取值，简单检查，比较简单。
核心的计数功能由`redisPopcount`实现：
1. 首先把没有对齐到4字节的值，每字节的计算位数（使用预设的表）
1. 随后每次循环计算28字节的位数
  1. 获取7个32位的变量`aux1`-`aux7`
  1. 计算这些变量的1的数量
1. 最后计算剩下的字节

比较有意思的部分是计算28字节的1的数量。
简单来说，通过若干步骤，依次获得每两位、每四位、每八位的数量，随后合并七个的数值的1数量的和。

首先从两位开始，考虑两位有四种情况`11b/10b/01b/00b`，分别有2/1/1/0个1，
特别地，对于高位为1的情况，都可以通过减1来获取“该两位有几个1”的结果；
对于高位为0的情况，可以通过减0来获取。
现在问题就转换成**“通过高位是否为1获取要减去1或0”**了。
为了解决这个问题，可以将数值右移一位，并与`01b`进行与运算。*(高位为0使得前一组的低位的值无效)*
现在再看代码就清晰了：
```c
aux1 = aux1 - ((aux1 >> 1) & 0x55555555); // 0x5 -> 0101b
```

现在则需要统合四位一组的1的个数。
考虑具体情况，每四位一组分为高低两部分，
分别存储了两个1的个数，现在就需要将两部分的值加起来。
这一部分简单的一些，分别对当前值用mask和位移来清空高低两部分，然后在相加即可。

```c
// 0x3 -> 0011b 可以将高位清除
// 第二部分将数值右移两位，然后与0x3与，
// 就将前一组的低位清除并将当前组的高位移动到低位上
aux1 = (aux1 & 0x33333333) + ((aux1 >> 2) & 0x33333333)
```

现在32位整数里，每四位的值为每四位的1的个数，
最终的答案自然只需要两步：
1. 先将8组的数值相加
1. 随后将7个值的数值相加

redis将这两步合并起来，简化了实现：

```c
bits += (
  (
    ((aux1 + (aux1 >> 4)) & 0x0F0F0F0F) + // 将每八位的数统合了起来
    ((aux2 + (aux2 >> 4)) & 0x0F0F0F0F) +
    ((aux3 + (aux3 >> 4)) & 0x0F0F0F0F) +
    ((aux4 + (aux4 >> 4)) & 0x0F0F0F0F) +
    ((aux5 + (aux5 >> 4)) & 0x0F0F0F0F) +
    ((aux6 + (aux6 >> 4)) & 0x0F0F0F0F) +
    ((aux7 + (aux7 >> 4)) & 0x0F0F0F0F)
  ) * 0x01010101) >> 24; // 这里先进行了乘法，对于高于32位的相当于溢出会被截断
                         // 右移24位相当于保留最高的八位(通过竖式不难看出)
                         // 高八位正好是划分的8位一组的四组的和，就是结果
```

# `BITPOS`

*用来查询第一个0或1的位置*

核心实现通过`redisBitpos`来实现。

首先按照`char`过滤没有对齐到`long`类型的部分，

```c
skipval = bit ? 0 : UCHAR_MAX;
c = (unsigned char*) s;
found = 0;
while((unsigned long)c & (sizeof(*l)-1) && count) {
    if (*c != skipval) {
        found = 1;
        break;
    }
    c++;
    count--;
    pos += 8;
}
```

随后如果没有找到符合的部分，在按照`long`来依次过滤全不符合的数据。

```c
l = (unsigned long*) c;
if (!found) {
    skipval = bit ? 0 : ULONG_MAX;
    while (count >= sizeof(*l)) {
        if (*l != skipval) break;
        l++;
        count -= sizeof(*l);
        pos += sizeof(*l)*8;
    }
}
```

随后加载下一部分数据，并检查位置。

```c
c = (unsigned char*)l;
for (j = 0; j < sizeof(*l); j++) {
    word <<= 8;
    if (count) {
        word |= *c; // 加载数据到word
        c++;
        count--;
    }
}
if (bit == 1 && word == 0) return -1; // 如果不存在1，返回-1
                                      // 对于0，默认会填充0，所以认为不会有不存在的情况

// 构建一个最高位是1的值
one = ULONG_MAX; /* 所有位都是1 */
one >>= 1;       /* 右移1位，最高位填充0 */
one = ~one;      /* 取反，获得需要的值 */

while(one) {
    // 通过移动one中的1，来依次检测每一个位置
    if (((one & word) != 0) == bit) return pos;
    pos++; // 每次移动增加位置
    one >>= 1;
}
```

最后返回结果。





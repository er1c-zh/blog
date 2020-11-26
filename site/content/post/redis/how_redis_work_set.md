---
title: "集合的实现"
date: 2020-11-26T01:34:30+08:00
draft: true
tags:
  - redis
  - what
  - how-redis-work
order: 4
---

redis列表相关指令的具体实现。

<!--more-->

{{% serial_index how-redis-work %}}

分类按照命令表的flag来确定，集合的是包含`@set`标签的指令。

*基于redis 6.0*

# 数据结构

集合有两种底层数据结构，第一种是
[整数集合]({{< relref "data_struct.md#整数集合">}})
或第二种是
[哈希表]({{< relref "data_struct.md#哈希表" >}})
。

```c
robj *setTypeCreate(sds value) {
  if (isSdsRepresentableAsLongLong(value,NULL) == C_OK)
    // 最终由string2ll尝试转换来判断
    return createIntsetObject(); // 整数集合
  return createSetObject(); // 其实是初始化一个哈希表
}
```

# 一些普遍的设计思想

类似于`list`类型，对于任何移除元素的指令，
都会检查移除的是否是最后一个元素，
如果是会删除对应的键值对来便于阻塞指令的实现。

另外会将一些复杂的指令的事件由若干简单指令来替代。

# `SADD`/`SREM`

`SADD`向集合中增加一个值，通过`setTypeAdd`来实现。
特别的，如果对应的集合是空，那么初始化一个。

`SREM`利用`setTypeRemove`从集合中移除一个特定的元素。
特别的，如果删除后集合为空，那么将该键删除。

# ```SMOVE```

命令先从`src`移除目标元素，如果成功移除，
那么将元素调用`setTypeAdd`增加到`det`集合。
在通知时，这个指令会触发`SREM`和`SADD`。

# ```SISMEMBER```

通过`setTypeIsMember`来实现具体功能。
对于哈希表类型，使用`dictFind`；
对于整数集合，使用`intsetFind`。

# ```SCARD```

平平无奇，```setTypeSize```。

# ```SPOP```

首先利用`setTypeRandomElement`随即选取一个元素，
然后利用`intsetRemove`或`setTypeRemove`来移除。

随机依赖标准库的`random`(哈希表)或`rand`(整数集合)来实现。

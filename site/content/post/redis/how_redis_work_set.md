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

集合有两种底层数据结构，第一种是数字集合或第二种是哈希表。

```c
robj *setTypeCreate(sds value) {
  if (isSdsRepresentableAsLongLong(value,NULL) == C_OK)
    // 最终由string2ll尝试转换来判断
    return createIntsetObject();
  return createSetObject();
}
```

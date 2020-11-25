---
title: "redis中pqsort的简单分析"
date: 2020-11-24T14:34:03+08:00
draft: false
tags:
  - redis
  - algorithm
  - what
---

`pqsort`提供从指定位置开始，排序指定长度的元素的功能。
主要使用在`SORT`指令上。
其基础思想是快排的分组。

主要的优化思想有如下几个：

1. 在数量较小的时候，采用插入排序而不是分治。
1. 只选择需要的范围进行精确的排序。
1. 选择比较用的值的时候，通过采样若干点来确定更中间的值。
1. 在分治时，后半部分用循环代替一次递归，减少栈的消耗。
1. 交换的时候尽量的提高每次交换的字长。

*src/pqsort.c*

首先看下交换，redis对容器对齐到long且元素长度为long的整数倍情况的交换采用了一些优化。

```c
// swaptype标志了元素的交换类型
// SWAPINIT计算出swaptype
// 0 容器对齐到long且元素长度是long的长度倍数，元素的长度为long
// 1 容器对齐到long且元素长度是long的长度倍数，元素的长度不为long：即long整数倍的长度，可以加速
// 2 如果要交换的容器的地址没有对齐long或元素的长度不为long的整数倍
#define SWAPINIT(a, es) swaptype = ((char *)a - (char *)0) % sizeof(long) || \
  es % sizeof(long) ? 2 : es == sizeof(long)? 0 : 1;

// swap宏是pqsort交换的入口
#define swap(a, b)            \
if (swaptype == 0) {           \
  long t = *(long *)(void *)(a);      \
  *(long *)(void *)(a) = *(long *)(void *)(b);  \
  *(long *)(void *)(b) = t;      \
} else              \
  swapfunc(a, b, es, swaptype) // 其他类型调用swapfunc来交换

static inline void
swapfunc(char *a, char *b, size_t n/*es*/, int swaptype)
{
  if (swaptype <= 1) // 对齐且长度是long两倍或以上，这个时候可以以long为单位来交换
    swapcode(long, a, b, n)
  else // 没有对齐或者长度不为long的整数倍，意味着不能按照long作为单位来交换，换用char
    swapcode(char, a, b, n)
}

#define swapcode(TYPE, parmi, parmj, n) {     \
  size_t i = (n) / sizeof (TYPE);     \ /* 计算要交换几次 */
  TYPE *pi = (TYPE *)(void *)(parmi);     \
  TYPE *pj = (TYPE *)(void *)(parmj);     \
  do {             \
    // 交换特定的次数
    TYPE  t = *pi;      \
    *pi++ = *pj;        \
    *pj++ = t;        \
  } while (--i > 0);        \
}
```

然后就是排序的具体实现：

```c
// a 要排序的容器
// n 容器的长度
// es entry_size 每个元素的长度？
// cmp 比较函数
//    - 1  val1 > val2
//    - 0   val1 == val2
//    - -1   val1 < val2
// lrange 开始
// rrang 结束
static void
_pqsort(void *a, size_t n, size_t es,
    int (*cmp) (const void *, const void *), void *lrange, void *rrange)
{
  char *pa, *pb, *pc, *pd, *pl, *pm, *pn;
  size_t d, r;
  int swaptype, cmp_result;

loop:  SWAPINIT(a, es); // 初始化交换时需要的一些值
  if (n < 7) { // 如果容器的长度小于7，完成一个插入排序
    for (pm = (char *) a + es; pm < (char *) a + n * es; pm += es)
      for (pl = pm; pl > (char *) a && cmp(pl - es, pl) > 0;
           pl -= es)
        swap(pl, pl - es);
    return;
  }
  // pm的含义是一个可能是中间大小的值的下标
  pm = (char *) a + (n / 2) * es; // 简单的获取中间
  if (n > 7) {
    // 数量大于7的时候
    // 采样获取一个更可能偏向中间的值的下标到pm上
    pl = (char *) a; // pl指向开头
    pn = (char *) a + (n - 1) * es; // pl指向结尾
    if (n > 40) {
      // 数量大于40的时候
      // 采样更多的点（9个）
      d = (n / 8) * es;
      pl = med3(pl, pl + d, pl + 2 * d, cmp); // med3 从三个下标中取出值位于中间的那个下标
      pm = med3(pm - d, pm, pm + d, cmp);
      pn = med3(pn - 2 * d, pn - d, pn, cmp);
    }
    pm = med3(pl, pm, pn, cmp); // 再结合一次
  }
  swap(a, pm); // 将采样的中等大小的值交换到第一个下标上，以下简称分割值
  pa = pb = (char *) a + es; // pa/pb指向第二个下标

  pc = pd = (char *) a + (n - 1) * es; // pc/pd指向最后一个下标
  // 这个循环结束后，[a,pa)和(pd, a + n * es)是等于分割值的元素
  // [pa,pc]是小于分割值的部分
  // [pb,pd]是大于分割值的部分
  for (;;) {
    while (pb <= pc && (cmp_result = cmp(pb, a)) <= 0) {
      // 这个循环用来找到第一个大于分割值的下标
      // cmp_result <= 0意味着 pb <= a
      if (cmp_result == 0) {
        swap(pa, pb);
        pa += es;
        // 这个时候，交换完之后的pb的值一定是小于等与分割的值
        // 考虑第一次比较，此时pa、pb指向同一个值，
        // 如果这个值大于分割值，那么就会与尾部交换，
        // 直到这个值小于等于分割值
      }
      pb += es;
    }
    while (pb <= pc && (cmp_result = cmp(pc, a)) >= 0) {
      // 和上一个循环类似，从尾部找到第一个小于分割值的下标
      if (cmp_result == 0) {
        swap(pc, pd);
        pd -= es;
      }
      pc -= es;
    }
    // 现在pb/pc分别指在大于和小于分割值的地方上，两者需要交换
    // pa之前，pd之后是等于分割值的地方
    if (pb > pc)
      break;
    swap(pb, pc); // 完成交换
    pb += es; // 进入下一次寻找
    pc -= es;
  }

  pn = (char *) a + n * es; // 指向尾部
  // 接下来调用两次vecswap来将等于的部分放到中间
  // 比较有趣的一点，不需要考虑是小于分割值的部分和等于分割值的部分哪个长一些
  // 因为只需要将[a, a + r]和[pb - r, pb]进行交换即可
  r = min(pa - (char *) a, pb - pa);
  vecswap(a, pb - r, r);
  r = min((size_t)(pd - pc), pn - pd - es); // 类似
  vecswap(pb, pn - r, r);
  if ((r = pb - pa) > es) {
    // 如果前半部分有值且要排序的范围与前半部分有关联，那么排序
    void *_l = a;
    void *_r = ((unsigned char*)a) + r - 1;
    if (!((lrange < _l && rrange < _l) || (lrange > _r && rrange > _r))) {
      _pqsort(a, r / es, es, cmp, lrange, rrange);
    }
  }
  if ((r = pd - pc) > es) {
    void *_l, *_r;

    // 如果后半部分有值且要排序的范围与后半部分有关联，那么排序
    a = pn - r;
    n = r / es;
    _l = a; // 更新了要排序的容器的基地址
    _r = ((unsigned char*)a)+r-1;
    if (!((lrange < _l && rrange < _l) || (lrange > _r && rrange > _r))) {
      // 这里避免了一次递归来减少栈空间
      goto loop;
    }
  }
}
```

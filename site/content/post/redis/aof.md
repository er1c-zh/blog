---
title: "redis AOF的实现"
date: 2021-03-02T23:39:28+08:00
draft: true
tags:
  - redis
  - how
---

上一篇学习了RDB相关的实现，接下来就轮到AOF了。

# 是什么

AOF是一种持久化方式，根据Append Only File，
可以得知是通过追加写的方式来实现的。

AOF通过追加写入指令来保存数据。


# 代码分析

我们从执行指令的`call`函数开始分析，
可以看到，
检查一些配置后，如果需要将指令扩展到下游，
就会调用`propagate`。

在`propatgate`中：

```c
if (server.aof_state != AOF_OFF && flags & PROPAGATE_AOF)
        feedAppendOnlyFile(cmd,dbid,argv,argc);
```

经过检查之后，通过调用`feedAppendOnlyFile`来写入到AOF中。
看起来这里就是将指令转化为AOF文件内容的地方。

## `feedAppendOnlyFile`

总的来说，这个函数是用来将执行的`redisCommand`对象，
转换为字符串存储到`server.aof_buf`，等待后续的刷盘。

```c
if (dictid != server.aof_selected_db) {
    char seldb[64];

    snprintf(seldb,sizeof(seldb),"%d",dictid);
    buf = sdscatprintf(buf,"*2\r\n$6\r\nSELECT\r\n$%lu\r\n%s\r\n",
        (unsigned long)strlen(seldb),seldb);
    server.aof_selected_db = dictid;
}
```

首先判断了`server.aof_selected_db`是否是当前指令的执行db，
如果不是，就写入一个`select`指令，切换数据库。

随后是一个大if，序列化指令到局部sds变量`buf`。

```c
if (server.aof_state == AOF_ON)
    server.aof_buf = sdscatlen(server.aof_buf,buf,sdslen(buf));
```

最后将`buf`通过`sdscatlen`追加到`server.aof_buf`。

特别的，针对后台执行AOF重写的情况，
调用`aofRewriteBufferAppend`来加速相关工作。

### 序列化指令

对于各种指令，有一些指令会进行归集。

- `EXPIRE` `PEXPIRE` `EXPIREAT`

    这三个指令将被归集到`PEXPIREAT`。
    目的一个是提高精度，另一个是转换为绝对时间。

    转换由`catAppendOnlyExpireAtCommand`完成。

    函数的实现非常的清晰，转换时间单位，相对时间转换为绝对时间，
    最后通过`catAppendOnlyGenericCommand`来序列化成字符串。

- `SETEX` `PSETEX`

    类似于上面的三个设置过期时间的指令，
    会转换为一个`SET`和一个`PEXPIREAT`。
    分别由`catAppendOnlyGenericCommand`写入`SET`
    和`catAppendOnlyExpireAtCommant`写入`PEXPIREAT`。

- 有过期时间的`SET`

    类似`SETEX`，如果有过期时间参数会尝试拆分转换。

- 其他指令

    利用`catAppendOnlyGenericCommand`来转换。

自然的，需要关注`catAppendOnlyGenericCommand`的实现，还是非常清晰的：

```c
sds catAppendOnlyGenericCommand(sds dst, int argc, robj **argv) {
    char buf[32];
    int len, j;
    robj *o;

    buf[0] = '*';
    len = 1+ll2string(buf+1,sizeof(buf)-1,argc); // 写入参数个数
    buf[len++] = '\r';
    buf[len++] = '\n';
    dst = sdscatlen(dst,buf,len); // 将参数个数写入dst

    for (j = 0; j < argc; j++) { // 遍历参数
        o = getDecodedObject(argv[j]);
        buf[0] = '$';
        len = 1+ll2string(buf+1,sizeof(buf)-1,sdslen(o->ptr)); // 写入参数内容的长度
        buf[len++] = '\r';
        buf[len++] = '\n';
        dst = sdscatlen(dst,buf,len); // 将参数内容的长度写入dst
        dst = sdscatlen(dst,o->ptr,sdslen(o->ptr)); // 写入参数内容
        dst = sdscatlen(dst,"\r\n",2); // 写入换行符
        decrRefCount(o);
    } 
    return dst;
}
```

## `server.aof_buf`何时落盘？

每次执行指令之后，都会将指令追加到`server.aof_buf`，
完成持久化只需要将这些数据写入盘中就完成了。

通过跟踪`server.aof_buf`的引用，
可以发现这些操作是由`flushAppendOnlyFile`完成的。

进一步，触发刷盘操作的地方分别有：

- `serverCron`

    运行时定时运行。

- `beforeSleep`

    时间循环的每次循环被调用。

- `prepareForShutdown`

    停机之前。

- `stopAppendOnly`

    停止AOF的时候刷盘。

---
title: "redis AOF的实现"
date: 2021-03-02T23:39:28+08:00
draft: false
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

## 执行落盘的`flushAppendOnlyFile`

`flushAppendOnlyFile`核心逻辑分为两部分：

1. 调用`write`将`server.aof_buf`写入到文件。
1. 根据条件选择是异步调用或者同步调用`fsync`保证数据刷到磁盘上。

下面开始关注更加细节的实现。

首先检查系统是否需要立即执行保存，如果不需要或者可以推迟的话，就直接返回。

保存的第一步是调用`write`的包装函数`aofWrite`将数据写入文件。
可以看到内部的实现是简洁清晰的。

```c
// flushAppendOnlyFile
nwritten = aofWrite(server.aof_fd,server.aof_buf,sdslen(server.aof_buf));

// aofWrite
ssize_t aofWrite(int fd, const char *buf, size_t len) {
    ssize_t nwritten = 0, totwritten = 0;

    while(len) {
        nwritten = write(fd, buf, len);

        if (nwritten < 0) {
            if (errno == EINTR) continue;
            return totwritten ? totwritten : -1;
        }

        len -= nwritten;
        buf += nwritten;
        totwritten += nwritten;
    }

    return totwritten;
}
```

写入后，检查是否写入成功。
如果失败，会将已经写入的数据清除，避免后续重试时出现重复写入的情况。

完成写入之后就会去完成刷盘的操作。

根据配置，如果配置了`AOF_FSYNC_ALWAYS`，
就同步的调用`redis_fsync`宏来完成刷盘，
具体的工作是`fdatasync`完成的。

反之，如果没有正在运行的异步刷盘人物，
就会调用`aof_background_fsync`来创建一个异步`bio`任务 *(`bioCreateBackgroundJob`)* 来完成刷盘。

# 学习到的

1. 序列化的时候考虑将相对的时间转换成绝对时间，能够简化实现。
1. 对于异常应该区分好临时异常，指可通过重试来避免的情况；和非临时异常。
1. 对于有状态的任务，要做好中途失败后状态的清理。

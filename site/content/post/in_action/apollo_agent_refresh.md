---
title: "[有趣事情的发生]携程apollo配置更新推送的实现"
date: 2021-01-21T20:04:39+08:00
draft: true
tags:
    - how
    - feat-in-action
order: 1
---

最近两个月在学习上一个是看了DDIA，一个是看了一些开源项目的源码。

看过redis的实现之后感觉阅读源码变得容易了一些。

看了一些项目之后发现有些遗忘了，~~在兄弟的鼓励下意识到~~ 博客好久不更新，正好今天周会无所事事，便开一个新的系列来记录这些学习的经历。

这个系列我打算稍微思考之后起一个名字。
思来想去，感觉这个系列是想记录一些纤细巧妙的实现，
所以就叫做：有趣事情的发生。

*希望能够坚持下去。*

<!--more-->

{{% serial_index feat-in-action %}}

[携程的apollo](https://github.com/ctripcorp/apollo)是一个分布式、高可用的配置服务。
因为工作中有一个也叫做apollo也是分布式、高可用的配置服务。
~~本来以为是魔改版的，不过到现在我也没有搞清楚两者的关系~~

配置服务，顾名思义就是配置一些参数，
然后主动、被动的令客户端感知。

根据功能来看，如果是单机实现，是很简单的：
1. 一个配置后台提供增删改
1. 客户端sdk查询配置内容

要求分布式与高可用，就需要考虑多一些东西。
1. 考虑到配置系统是一个明显的少写多读且无写热点的业务，
性能上更多的关注下读取的性能与配置生效的延迟
1. 分布式与高可用通过Eureka来提供了基础的服务发现功能，
在调用的上游做负载均衡、错误重试等

详细的设计官方提供了比较好的[文档](https://ctripcorp.github.io/apollo/#/zh/README)这里就不再赘述。

在开始介绍更新变更的通知的实现之前，不得不赞叹一句：
Spring在工业化开发上确实有可取之处。

# 配置投送的流程

首先引入官方的设计图：

![apollo-design-from-official](https://raw.githubusercontent.com/ctripcorp/apollo/master/doc/images/overall-architecture.png)

配置从右上角的`portal`接入，走`remoteCallAPI`到`Admin Service`保存到`ConfigDB`。

`ConfigDB`是一个Mysql数据库，每次变更都会在表中记录变更的记录。

`Admin Service`会在每次配置变更的时候向数据库(MySQL)写入通知，`Config Service`通过定时扫描对应的表来查询是否有变化。

写入的地方在`com.ctrip.framework.apollo.biz.message.DatabaseMessageSender#sendMessage`，
可以看到在release的时候会调用。

扫描的地方在`com.ctrip.framework.apollo.biz.message.ReleaseMessageScanner#scanMessages`。
具体的会每次扫描五百条release消息，
然后更新本地的消费过的消息id(`com.ctrip.framework.apollo.biz.message.ReleaseMessageScanner#maxIdScanned`)，
触发注册的handler(`com.ctrip.framework.apollo.biz.message.ReleaseMessageScanner#fireMessageScanned`)。

可以看到代码非常的简单，也没有什么奇特的操作。
```java
private void fireMessageScanned(List<ReleaseMessage> messages) {
  for (ReleaseMessage message : messages) {
    for (ReleaseMessageListener listener : listeners) {
      try {
        listener.handleMessage(message, Topics.APOLLO_RELEASE_TOPIC);
      } catch (Throwable ex) {
        Tracer.logError(ex);
        logger.error("Failed to invoke message listener {}", listener.getClass(), ex);
      }
    }
  }
}
```

现在唯一不清楚的地方就在于什么时候注册了handler到`listener`中，
另外我还对这种实现有一些性能上的担忧，如果连接的client过多会不会影响性能。

通过追踪`listener`的调用，
发现只有`com.ctrip.framework.apollo.biz.message.ReleaseMessageScanner#addMessageListener`调用了`add`方法。
然后发现只注册了有限的handler，不是每个client都会注册handler。

```java
public ReleaseMessageScanner releaseMessageScanner() {
  ReleaseMessageScanner releaseMessageScanner = new ReleaseMessageScanner();
  //0. handle release message cache
  releaseMessageScanner.addMessageListener(releaseMessageServiceWithCache);
  //1. handle gray release rule
  releaseMessageScanner.addMessageListener(grayReleaseRulesHolder);
  //2. handle server cache
  releaseMessageScanner.addMessageListener(configService);
  releaseMessageScanner.addMessageListener(configFileController);
  //3. notify clients
  releaseMessageScanner.addMessageListener(notificationControllerV2);
  releaseMessageScanner.addMessageListener(notificationController);
  return releaseMessageScanner;
}
```

可以看到注册了很多handler，其中实现下发的handler是`notificationControllerV2`。

继续跟踪下去，省略了一些解析message和检查的代码，
可以看到实际下发的操作是向`DeferredResult`写入结果来实现的，
而且之前担心的性能问题在这里通过超过阈值后会异步通知来避免阻塞扫描-通知线程。

具体的，在异步任务中，会每发送一组后sleep一段时间，
因为对于Java的调度不了解，所以这里先猜测一下是让出资源，
避免影响其他线程执行，以后有机会再来了解一下。

```java
//do async notification if too many clients
if (results.size() > bizConfig.releaseMessageNotificationBatch()) {
  largeNotificationBatchExecutorService.submit(() -> {
    logger.debug("Async notify {} clients for key {} with batch {}", results.size(), content,
        bizConfig.releaseMessageNotificationBatch());
    for (int i = 0; i < results.size(); i++) {
      if (i > 0 && i % bizConfig.releaseMessageNotificationBatch() == 0) {
        try {
          TimeUnit.MILLISECONDS.sleep(bizConfig.releaseMessageNotificationBatchIntervalInMilli());
        } catch (InterruptedException e) {
          //ignore
        }
      }
      logger.debug("Async notify {}", results.get(i));
      results.get(i).setResult(configNotification);
    }
  });
  return;
}

logger.debug("Notify {} clients for key {}", results.size(), content);

for (DeferredResultWrapper result : results) {
  result.setResult(configNotification);
}
```

现在就很明确了，
写入`com.ctrip.framework.apollo.configservice.controller.NotificationControllerV2#deferredResults`的地方就是注册客户端的地方。

`com.ctrip.framework.apollo.configservice.controller.NotificationControllerV2#pollNotification`就是我们要找的最后的一个节点，
这是一个HTTP的endpoint，返回的是一个`DeferredResult`，这表示这会是一个长HTTP链接。

*com/ctrip/framework/apollo/configservice/controller/NotificationControllerV2.java:118*

```java
DeferredResultWrapper deferredResultWrapper = new DeferredResultWrapper(bizConfig.longPollingTimeoutInMilli());
```

这里可以看到设置了过期时间，默认是60s。

```java
//register all keys
for (String key : watchedKeys) {
  this.deferredResults.put(key, deferredResultWrapper);
}
```

然后注册了关注的key，在完成的回调中移除了。

到此为止，就是整个配置发版之后通知到client的流程。

# 分析与小结

总的来说，功能的实现非常的简单，实现了一个**极少写，及时感知变化**的系统，下面从最一开始来分析整个系统的实现思路。

系统的目标有一个是配置下发系统，要求可靠，以及在下发的时候需要一些性能。

可靠这一点apollo通过Eureka的可靠性来保证，存储上利用MySQL来保证。
以上的两点都是保证了“服务的可用”，
稍微多一些，实际的业务中的可靠可能更加的宽松一些，只要client能够读取到配置就可以了，
那么可以在client上做一些缓存的策略，对于一台机器来说，能部署的实例和需要的配置项都是有限的，所以应该是可行的。

对于性能，我一直认为要根据业务的特性来设计，apollo的设计我觉得就很贴合业务场景。
配置的变更是一个频率很小的事情，
而读取则要大的多，所以是一个很明显的少写多读的场景。
相对于普通的少写多读，配置系统还要求生效的时效性。

这么一来，解决方案就朝着主动推送去了，
这里考虑到频繁的轮询对于性能和资源是有影响和浪费的，
所以换成一个长链接来监听数据的变化，通过`Admin Service`内部的合并，最终转化成较少的sql查询。

整个系统通过一步一步的转换，没有增加更多的外部依赖就实现了需求。

这种简单的实现方式令人印象深刻，
在日常的开发中，
我会经常会想用很多“花里胡哨”的方案，显然，这是一种错误。
为了避免这种问题，一个是思路上要避免过度的追求新技术，
另一个是要关注业务的特点，发现可以做trade-off的地方。

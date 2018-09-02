---
categories:
  - elasticsearch-and-lucene
---
# ES接口

*Elasticsearch 6.3*

[TOC]

## Document APIs

> CURD APIs

- 用于CURD文档的接口
- 只针对一个index
- 分类
  - Single document APIs
  - Multi-document APIs

## Search APIs

- 搜索接口

- 允许给出一个搜索条件 *（一个参数或者DSL）* 得到符合搜索条件的结果集

- 搜索接口允许在

  - 一个index中，多个type
  - 多个index

  中搜索

## Indices APIs

- 用于索引的curd和其他操作 *Index management*
- 用于管理Mapping*（映射）* *Mapping management*
- 用于管理别名 *alias management*
- 用于管理index设置 *Index settings*
- 用于管理索引 *Monitoring*
- 用于查看、管理索引状态 *Status management*

## cat APIs

- 用于给出一些 **更加贴合人观察数据习惯** 的接口
- 通常用于展示一些状态

## Cluster APIs

- 集群的状态
- 通常可以使用参数限制需要查询的节点
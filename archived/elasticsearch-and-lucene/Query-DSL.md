---
title: Query DSL
categories:
  - elasticsearch-and-lucene
---

## 引入

ES支持使用 **基于JSON表示的Query-DSL** 定义查询。

## DSL

Domain Sepcific Language

领域特定语言

## Query-DSL

> Think of the Query DSL as an AST (Abstract Syntax Tree) of queries, consisting of two types of clauses。

- 可以认为 Query-DSL 是一个 查询语法树
- Query-DSL包含两种 **规则 clauses**

### 规则 clauses

#### 分类

- `Leaf query clauses`

  - > look for a particular value in a particular field, such as the`match`, `term` or `range` queries. 

  - 通过不同的比较方法，比较特定字段是否符合特定的值

- `Compund query clauses`

  - 包含 `Leaf query clauses` 和 `Compound query clauses`
  - 两者使用使用特殊的逻辑关键字组合起来

#### 规则的行为由规则所处的上下文决定

### 上下文 context

#### 分类

- Query context
- Filter context

#### Query context

一个规则在 Query context 中，做两件事情：

- 判断 `document` 是否符合给出的规则
- 计算 `_score` 表示 `document` 的符合程度

#### Filter context

一个规则在 Filter context 中，只判断 `document` 是否符合规则

常用的 `过滤器 Filters` ，将会被ES自动缓存来改善性能

#### 如何判断处于哪一种上下文？

> Query context is in effect whenever a query clause is passed to a `query`parameter.
>
> Filter context is in effect whenever a query clause is passed to a `filter`parameter.

- 当Query-DSL被作为`query` 参数传入的时候，处于 `Query context`
- 被作为 `filter` 参数传入的时候，处于 `Filter context`

#### 一个case

```shell
curl -X GET "localhost:9200/_search" -H 'Content-Type: application/json' -d'
{
  "query": { // query context
    "bool": { // query context
      "must": [
        { "match": { "title":   "Search"        }}, // query context
        { "match": { "content": "Elasticsearch" }}  // query context
      ],
      "filter": [ // filter context
        { "term":  { "status": "published" }},  // filter context
        { "range": { "publish_date": { "gte": "2015-01-01" }}}  // filter context
      ]
    }
  }
}
'
```

## 无条件查询

``` json
{
    "query": {
        "match_all": {} // 匹配所有
    }
}

```

``` json
{
    "query": {
        "match_all": {
            "boost": 1.2 // 修改此条规则在评分时的权值
        }
    }
}
```

``` json
{
    "query": {
        "match_none": {} // 不匹配任何一条记录
    }
}
```



## 全文查询 Full text queries

通常使用于搜索 `text` 字段

搜索行为依赖于字段被如何分析，并且字段的 `analyzer` 或 `search_analyzer` 会在查询规则执行之间被应用

### 分类

- match 包括 模糊查询 *fuzzy* 短语查询 *phrase* 临近匹配 *proximity*
- match_phrase 精确的 短语查询 和 *word proximity matches*
- match_phrase_prefix 类似于 `match_phrase` 但 *does a wildcard search on the final word*
- multi_match
- common
- query_string
- simple_query_string

### match


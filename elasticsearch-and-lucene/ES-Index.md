# ES 索引

[TOC]

## 索引

- 索引有一个或者多个 `分片` 组成
- 每个 `分片` 有0到若干个 `副本`
- 分片和 *（每个分片）* 副本的数量在索引创建时可以规定或者忽略信息使用配置文件或者实现中的默认值 *5个分片 （每个分片）1个副本*

### 分片

- 每个 `分片` 对应一个 `Lucene索引`
- 每个 `分片` 包含 `索引文档集` 的一部分
- 分片的数量只有在创建索引的时候可以规定， **一旦创建好了索引，更改分片数量的唯一途径就是创建另一个索引并重新索引数据**

### 副本

### 创建索引

#### 自动创建

当 `action.auto_create_index` 配置为 `true` 或者 匹配 `action.auto_create_index` 规定的 `pattern` 的时候，如果向这个 **没有创建** 的索引写入文档时，就会自动创建这个索引

#### 手动创建

``` shell
curl -XPUT http://es:9200/indexName -d '{
    "settings": {
        "number_of_shards": 3, // 分片数
        "number_of_replicas": 2 // 每个分片的副本数
    }
}' // 新建的索引拥有 3 * (1 + 2) 个物理Lucene索引
```

### 删除索引

```shell
curl -XDELETE http://es:9200/indexName
```

## 映射

### 类型确定机制

- ES使用定义 `document` 的JSON猜测文档结构

- ```json
  {
      "field1": 10, // Number
      "field2": "10", // String
  }
  ```

- 当数据源提供的数据全部为字符串时，通过定义 `映射` 来处理输入

- ```shell
  curl -XPUT 'http://es:9200/indexName/' -d'{
      "mappings": {
          "article": { // 定义了 类型article 的映射
              "numeric_detection": true,
              "dynamic_date_formats": [
                  "yyyy-MM-dd hh:mm",
                  "yyyy-MM-dd hh:mm:ss"
              ]
          }
      }
  }'
  ```

- 处于精确性和可靠性的考虑，我们会选择关闭 `字段类型猜测`

- ```shell
  curl -XPUT 'http://es:9200/indexName' -d'{
      "mappings": {
          "typeName": {
              "dynamic":"false",
              "properties": {
                  "字段名"： {
                      "type": "string"
                  },
                  "field2": {
                      "type": "string"
                  }
              }
          }
      } // 在 索引indexName 中，关闭了 字段类型猜测。
  }' // 意味着 在 索引indexName 类型typeName中插入文档，除了在properties中出现过的字段 其他字段都会被忽略
  ```

### 索引结构映射

`模式映射 schema mapping` 用于定义索引结构

#### 核心类型

每个 **字段** 都可以指定为ES提供的一个核心类型

##### 分类

- string

- number

- date
- boolean
- binary

##### 字段的公共属性 *除binary*

- `index_name` 定义在索引中的字段名称，如果为空，将以对象的名称作为字段的名称 Q2A
- `index`  [analyzed, no] 特别的对于 `string` 还有值 `not_analyzed` 可选
  - `analyzed` 将会编入索引进行搜索
  - `no` 不会编入索引，无法以这个字段进行搜索
  - `not_analyzed` 将字段不进行分析而直接编入索引 *（即搜索过程中必须全部匹配）*
- `store` [yes, no] 规定字段 **原始值是否写入索引** **区别于是否编入索引，此项配置为不写入也可以搜索**
  - `yes` 字段将会被返回
  - `no` 字段只会在 `_source` 字段中返回
- `boost` 字段的权值
- `null_value` 指定该字段 **写入索引** 的默认值
- `copy_to` 此属性指定一个字段，字段的所有值都将复制到该指定字段。
- `include_in_all` 此属性指定该字段是否应包括在_all字段中

##### 各 *字段* 核心类型私有属性

- string
  - `term_vector` 它定义是否要计算该字段的Lucene词 向量（term vector）。如果你使用高亮，那就需要计算这个词向量。Q2A
  - `omit_norms` Q2A
  - `analyzer` 定义用于索引和搜索的 `分析器` 名称，默认为全局定义的分析器名称
  - `index_analyzer` 用于定义建立索引的 `分析器` 名称
  - `search_analyze` 用于定义搜索的 `分析器` 名称
  - `norms.enabled` Q2A
  - `norms.loading` Q2A
  - `position_offset_gap` Q2A
  - `index_options` Q2A
  - `ignore_above` 定义字段中字符的最大值。当字段的长度高于指定值时，`分析器` 会将其忽略。
- number
  - 数字类型，包括如下的类型
    - `byte`
    - `short`
    - `integer`
    - `long`
    - `float`
    - `double`
  - 私有属性
    - `precision_step` 此属性指定为某个字段中每个值生成的词条数。Q2A
    - `ignore_malformed` [true, false] 若要忽略格式错误的值，则应设置属性值为true。
- boolean
- binary
  - 以二进制数据的Base64表示
  - **只被存储，不被索引；只能提取，不能搜索**
  - 只支持 `index_name`
- date
  - 默认使用UTC表示
  - 私有属性
    - `format` 指定日期的格式
    - `precision_step` 此属性指定为某个字段中每个值生成的词条数。Q2A
    - `ignore_malformed` [true, false] 若要忽略格式错误的值，则应设置属性值为true。

#### 多字段

> 有时候你希望两个字段中有相同的字段值，例如，一个字段用于搜索，一个字段用于排序； 或一个经语言分析器分析，一个只基于空白字符来分析。Elasticsearch允许加入多字段对象来拓 展字段定义，从而解决这个需求。它允许把几个核心类型映射到单个字段，并逐个分析。

Q2A

#### 其他类型

##### IP地址类型

```json
{
    "ip":{
        "type": "ip",
        "store": "yes"
    }
}
// document
{
    "ip": "127.0.0.1"
}
```

##### token_count类型

> token_count字段类型允许存储有关索引的字数信息，而不是存储及检索该字段的文本。它 接受与number类型相同的配置选项，此外，还可以通过analyzer属性来指定分析器。

```json
{
    "address_count": {
        "type": "token_count",
        "store": "yes"
    }
}
```

#### 使用分析器

> Elasticsearch使我们 能够在索引和查询时使用不同的分析器，并且可以在搜索过程的每个阶段选择处理数据的方式。 使用分析器时，只需在指定字段的正确属性上设置它的名字，就这么简单。

- 开箱即用的分析器

  - `standard` 大多数欧洲语言的分析器

  - `simple` 基于 **非字母字符** 分离所提供的值，并转换为小写

  - `whitesapce` 基于 **空格字符** 分离

  - `stop` 类似于 `simple分析器` ，并能基于提供的 `stop word` 过滤数据

  - `keyword` 与 `not_analyzed` 提供相同的功能

  - `pattern` 利用regex

  - `language`

  - `snowball` 类似于 `standard分析器` ，并提供了 `词干提取算法 stemming-algorithm`

    - > 词干提取(stemming)是将 屈折词和派生词 值其基本形式的过程
      >
      > e.g. cars -> car

- 自定义分析器

  - 向 `映射文件` 中添加 `settings节` ，包含了es需要的创建索引时所需要的有用信息

  - `settings` 例子

    ```json
    {
        "mappings": {
            // 定义映射
        },
        "settings": {
            "index": {
                "analysis": {
                    "analyzer": {
                        "en": { // 指定一个新的分析器 en
                            "tokenizer": "standard", // 分词器
                            "filter": [ // 多个过滤器
                                "asciifolding", // es自带的
                                "lowercase", // es自带的
                                "ourEnglishFilter" // 自定义的
                            ]
                        }
                    },
                    "filter": {
                        "ourEnglishFilter": { // 自定义的过滤器的定义
                            "type": "kstem"
                        }
                    }
                }
            }
        }
    }
    ```

- `分析器字段 _analyzer` 存储一个 `字段名` ，这个字段的内容将被作为此份文档的 `分析器名称`

- `默认分析器` 在 `settings` 中添加一个用 `default` 命名的分析器

#### JSON定义映射例子

```json
{
    "mappings": {
        "post": { // 类型 post
            "_analyzer": {
                "path": "language"
            },
            "properties": { // 属性
                "id": { // 字段 id
                    "type": "long", // 核心类型
                    "store": "yes",
                    "precision_step": "0"
                },
                "name": { // 字段 name
                    "type": "stirng",
                    "store": "yes",
                    "index": "analyzed"
                },
                "publised": {
                    "type": "date",
                    "store": "yes",
                    "precision_step":"0"
                },
                "contents": {
                    "type": "string",
                    "store": "no",
                    "index": "analyzed"
                }
            }
        },
        "other_type": {
            "properties": {
                
            }
        }
    },
    "settings": {
        "index": {
            "analysis": {
                "analyzer": {
                    "en": { // 指定一个新的分析器 en
                        "tokenizer": "standard", // 分词器
                        "filter": [ // 多个过滤器
                            "asciifolding", // es自带的
                            "lowercase", // es自带的
                            "ourEnglishFilter" // 自定义的
                        ]
                    },
                    "default": { // 默认的分析器
                        "tokenizer": "standard",
                        "filter": [
                            "asciifolding"
                        ]
                    }
                },
                "filter": {
                    "ourEnglishFilter": { // 自定义的过滤器的定义
                        "type": "kstem"
                    }
                }
            }
        }
    }
}
```

### 相似度模型

> 相似度模型是定义了相似度的抽象和度量。

换句话说就是 **量化的反映了两个实体之间的相似程度**

#### 文本相似度模型

文本相似度模型可以分为两类

- 文档分类：将文档划分到已知有限类集合中的某一类
- 信息检索：找到与查询最相似的文档

#### ES提供的相似度模型

- tf/idf
- bm25
- 随机性偏差模型 *divergence from randomness*
- 信息基础模型 *information-based*

### 信息格式

> Apache Lucene 4.0的显著变化之一是，可以改变索引文件写入的方式。Elasticsearch利用了此
> 功能，可以为每个字段指定信息格式。有时你需要改变字段被索引的方式以提高性能，比如为了
> 使主键查找更快。

“设置字段的存储格式来获得特殊的性能提升”

- `default` 默认的格式。提供实时的压缩

- `pulsing` 将 `高基数字段 high cardinality field` 的信息列表编码为 `词条矩阵`

  - > 这让Lucene检索文档时可以少执行一个搜索。对高基数字段使用此信息格式可以加快此
    > 字段的查询速度。

  - `高基数字段` 区分度高的字段

  - `词条矩阵` Q2A

- `direct` 

- `memory`

- `bloom_default`

- `bloom_pulsing`

#### 配置信息格式

```json
{
    "mappings": {
        "post": {
            "properties": {
                "id": {
                    "type": "long",
                    "store": "yes",
                    "precision_step": "0",
                    "postings_format": "pulsing" // 设置字段的信息格式
                },
                "name": {
                    "type": "string",
                    "store": "yes",
                    "index": "analyzed"
                },
                "contents": {
                    "type": "string",
                    "store": "no",
                    "index": "analyzed"
                }
            }
        }
    }
}
```

### 文档值

> 文档值允许定义一个给定字段的值被写入一个具有较高内存效率的列式结构，以便进行高效的排
> 序和切面搜索。使用了文档值的字段将有专属的字段数据缓存实例，无需像标准字段一样倒排（以
> 避免像第1章所描述的方法一样存储）。因此，它使索引刷新操作速度更快，让你可以在磁盘上存
> 储字段数据，从而节省堆内存的使用。

#### 优势

- 高效的排序和切面搜索

- 加快索引刷新操作的速度

- > 让你可以在磁盘上存储字段数据，从而节省堆内存的使用。

#### 分类

不同的文档值在 **内存使用** 和 **性能** 上进行了不同的妥协

- `default` 
- `disk` 
- `memory` 

#### 例子

```json
{
    "mappings": {
        "post": {
            "properties": {
                "votes": {
                    "type": "integer",
                    "doc_values_format": "memory" // 文档值
                }
            }
        }
    }
}
```

## 批量索引以提高索引速度

*需要分清楚这一节中 `索引` 有时是动词，有时是名词*

### 进行批量索引的数据格式

#### 操作分类

- `index` 在索引中增加或更换现有文档
- `delete` 从索引中移除文档
- `create` 当索引中不存在其他文档时，在索引中增加新文档

#### 数据格式

```json
// 必须一个一行
{ "index": { "_index": "addr", "_type": "contact", "_id": 1 }} // 操作-index
{ "name": "Fyodor Dostoevsky", "country": "RU" } // 操作需要的数据
{ "create": { "_index": "addr", "_type": "contact", "_id": 2 }}
{ "name": "Erich Maria Remarque", "country": "DE" }
{ "create": { "_index": "addr", "_type": "contact", "_id": 2 }}
{ "name": "Joseph Heller", "country": "US" }
{ "delete": { "_index": "addr", "_type": "contact", "_id": 4 }}
{ "delete": { "_index": "addr", "_type": "contact", "_id": 1 }}
```

#### 进行索引

```shell
curl -XPOST 'es:9200/_bulk' --data-binary @data.json
curl -XPOST 'es:9200/indexName' --data-binary @data.json # 指定默认的索引
curl -XPOST 'es:9200/indexName/typeName' --data-binary @data.json # 指定默认的 索引 类型
```

## 用附加的内部信息扩展索引结构

- `_uid`
- `_type`
- `_all`
- `_source`
- `_index`
- `_size`
- `_timestamp`
- `_ttl`

## 段合并

### 为什么要段合并

ES和Lucene中的数据一旦被写入 *某些结构* ，就不再改变 *（这个描述过于绝对，某些情况也会修改文档）*，这导致了：

- 被删除的文档并没有真正的删除，而是占据了 **磁盘空间** ，并且会 **拖慢搜索和增加搜索时需要的内存和CPU资源**
- 组成一个索引的 **段越来越多** ，搜索时需要的 **内存和CPU资源** 更多

### 段合并的过程

1. 由Lucene获取若干段
2. 创建新段，保存上一步中选择的段中除去被删除的文档之外的所有文档
3. 删除原段

### 合并策略

- `tiered` *（默认的合并策略）* 合并尺寸大致相似的段，并考虑每个 `层 tier` 允许的最大段数量
- `log_byte_size` *随着时间的推移* ，将产生由索引 *（字节数）* 大小的对数构成的索引，其中存在着一些较大的段以及一些 `合并因子` 较小的段
- `log_doc` 和 `log_byte_size` 类似，不同在于不使用索引字节数而是使用文档数

### 合并调度器

`合并调度器` 指示ES合并过程的方式

- `并发合并调度器` 默认的合并调度器，在独立的线程中执行，定义好的线程数量可以并行合并
- `串行合并调度器` 合并过程在 `调用线程（即执行索引的线程）` 执行，合并进程会一直阻塞线程直到合并完成

### 合并因子

`合并因子` 指定了段合并的频率。较小的时候搜索速度更快，占用的内存也更少，索引的速度会变慢；较大时，索引的速度加快，这是因为 **发生的合并较少** ，但是搜索的速度变慢，占用的内存也会变大。

### 调节

`调节 throttling` 允许限制 `合并` 的速度，也可以用于控制使用数据存储的所有操作

#### 两个配置条目

- `indices.store.throttle.type`
  - `none` 不打开 `调节`
  - `merge` 只在 `合并` 过程中有效
  - `all` 在所有的数据存储过程中有效
- `indices.stroe.throttle.max_bytes_per_sec`

## 路由介绍

`路由` 允许选择用于索引和搜索数据的分片

### 例子

```shell
# 索引文档 单个路由参数
curl -XPUT 'http://es:9200/index_name/type_name/doc_id?routing=router_val' -d @data.json
# 查询 单个路由参数
curl -XGET 'http://es:9200/index_name/type_name/_search?routing=router_val&q=k:v'
curl -XGET 'http://es:9200/index_name/type_name/_search?routing=router_val1,router_val2&q=k:v' # 多个路由参数由','隔开
```

```json
// 指定路由参数字段
{
    // 类型定义
    "_routing": {
        "required": true,
        "path": "routing_field_name"
    }
}
```

## 参考

- [Elasticsearch中的相似度模型(原文：Similarity in Elasticsearch)](https://www.cnblogs.com/sheeva/p/6847309.html)
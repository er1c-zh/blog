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


# mapping

*v7.1*

mapping用于定义一个文档和其中的字段是如何存储和被索引的。

可以被用作：

- 指定某个字段被当作全文 *(full text)* 来处理
- 说明每个字段的格式，如数字、日期或者地理位置等
- 一些自定义的规则用来控制动态mapping的插入

## Mapping Type

每个索引有且一种mapping，用来决定索引中的文档应该被如何索引。

> 从v6.0.0开始，一个索引应该只有一个mapping。随着版本的迭代，后续会不再支持一个索引多个mapping。
> [为什么要移除mapping的类型](https://www.elastic.co/guide/en/elasticsearch/reference/7.1/removal-of-types.html)

### Mapping Type 有什么

#### Meta-fields 元数据

todo 

#### Fields or properties 字段与属性

与要存储的的文档相关的字段与属性的列表

## 创建mapping要注意的点 / 创建mapping的最佳实践

### 避免mapping的过度膨胀

在一个mapping中定义过多的字段，会导致性能上的下降。

使用一些属性来避免过度膨胀：

- `index.mapping.total_fields.limit`
- `index.mapping.depth.limit`
- `index.mapping.nested_fields.limit`
- `index.mapping.nested_objects.limit`

### mapping的字段一但定义便无法修改

可以通过新建字段，并reindex来实现修改的目的。

### 动态mapping与声明式 *(explicit)* mapping

默认的，当插入新的文档时，如果有新的字段，将会自动在索引的mapping中，添加对应的字段。
这种mapping称为动态mapping。

某些时候，我们不想让ES自动生成和修改mapping，那么我们可以使用如下的方式解决：

1. 创建索引时定义mapping
1. 使用 `PUT` 接口修改mapping
1. 将 `dynamic` 属性设置为 `strict` 来禁用dynamic mapping
    ```
        PUT /index
        {
                "mappings": {
                        "properties": {
                                "dynamic": "strict"
                        }
                }
        }
    ```

## 字段的数据类型

每个字段都有固定的类型，有如下分类：

### 基础数据类型

#### 字符串

#### 数字

#### 日期

#### Date nanoseconds

#### 布尔值

#### 二进制

#### range

### 复合数据类型

#### Object 单个Json对象

#### Nested

nested for arrays of JSON object

### 地理位置数据类型

#### 地理点 Geo-Point

#### 地理区域 Geo-shape

用于表示一个地理区域

### 特殊用途的数据类型

#### IP

用于存储ip地址，支持ipv4和ipv6

#### 自动补全

用于提供自动补全支持的字段类型

#### 词数统计

用于统计一个字符串中有多少（经过分词之后产生的）词

#### mapper-murmur3

#### mapper-annotated-text

#### percolator

#### Join

#### Alias 别名

#### Rank feature

#### Rank features

#### Dense vector

#### Sparse vector

### 数组

### multi-fields 复合字段

对于一个字段来说，经常会遇到一份数据，按照多种方式解析的情景。复合字段就是用来处理这种情况的。



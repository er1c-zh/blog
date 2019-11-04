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

#### Object 对象

对象类型可以看作是JSON对象。JSON格式的数据自身就是分层的。每个文档可能会包含一个JSON对象，如此递归下去。
就像这样：

```crul
PUT my_index/_doc/1
{
  "region": "US",
  "manager": {
    "age":     30,
    "name": {
      "first": "John",
      "last":  "Smith"
    }
  }
}
```
文档自身就是一个JSON对象，其中的 `manager` 字段也是一个JSON对象，显然， `name` 也是一个JSON对象。

在ES内部，整个文档会被存储为一个平铺的kv对列表，多层的对象会被转换为：

```json
{
  "region":             "US",
  "manager.age":        30,
  "manager.name.first": "John",
  "manager.name.last":  "Smith"
}
```

##### 如何声明一个对象类型的字段

```curl
PUT my_index
{
  "mappings": {
    "properties": {
      "region": {
        "type": "keyword"
      },
      "manager": {
        "properties": {
          "age":  { "type": "integer" },
          "name": {
            "properties": {
              "first": { "type": "text" },
              "last":  { "type": "text" }
            }
          }
        }
      }
    }
  }
}
```

与其他字段类型不同，Object字段类型不需要显示的用 `type` 字段来声明。

##### 参数

- `dynamic` 是否允许动态的修改 `mapping` 的 `properties`
- `enabled`
- `properties`

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

用于给索引中的一个字段增加一个别名。

- 一些限制
    - 别名只能指向一个具体的字段，而不能是另一个别名
    - 别名只能指向一个已经存在的字段
    - todo 嵌套对象的别名
- 写入接口均不支持别名
    - 因为别名不会存储在文档的source中

#### Rank feature

#### Rank features

#### Dense vector

#### Sparse vector

### 数组

es中没有专门的数组类型，每一个字段都可被看作是一个数组。但一个数组中只能有一种类型的值。

#### 对象数组

### multi-fields 复合字段

对于一个字段来说，经常会遇到一份数据，按照多种方式解析的情景。复合字段就是用来处理这种情况的。

## Mapping的字段类型的参数

不同的字段类型，支持不同的参数。

### dynamic

用于控制索引或 `object` 是否允许动态的插入新的字段，以及不允许时，新增的字段会被如何处理。

**特别的，`子Object` 会从 `父Object` 或 `mapping` 继承这个属性。**

- `true` 默认值，在索引新的文档时，如果有新的字段，会自动的插入到 `mapping` 中。
- `false` 忽略新增的字段，不会被索引，即无法被搜索。但是仍然会被存储，并在 `_source` 中返回。
- `strict` 严格模式，如果有新的字段，会抛出异常。


> 该属性可被修改

### enabled

用于控制字段是否需要被索引。**只有 `mapping` 或 `Object` 支持。**

- `true` 会被索引
- `false` 不会被索引，但仍然会被存储，可以在 `_source` 中查询到。

### properties

用于表明 `mapping` 和 `Object` 和 `nested` 下的字段。


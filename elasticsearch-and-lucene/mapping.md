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

## 字段的数据类型

每个字段都有固定的类型，有如下类型：

- 简单类型
    - text
    - keyword
    - date
    - long
    - double
    - boolean
    - ip
- 类json类型的数据
- 特殊类型的数据
    - 地理位置相关类型
    - 复数
    - etc

### multi-fields 复合字段

对于一个字段来说，经常会遇到一份数据，按照多种方式解析的情景。复合字段就是用来处理这种情况的。

### 避免mapping的过度膨胀

在一个mapping中定义过多的字段，会导致性能上的下降。

使用一些属性来避免过度膨胀：

- `index.mapping.total_fields.limit`
- `index.mapping.depth.limit`
- `index.mapping.nested_fields.limit`
- `index.mapping.nested_objects.limit`

## 动态mapping与声明式 *(explicit)* mapping

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

## mapping的字段一但定义便无法修改

可以通过新建字段，并reindex来实现修改的目的。


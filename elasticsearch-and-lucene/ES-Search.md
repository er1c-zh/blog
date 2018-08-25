# ES 搜索

[TOC]

## 查询

### ES支持的两种查询

- `基本查询` 
- `复合查询` 

### 简单查询

```shell
curl -XGET 'es:9200/index_name/type_name/_search?q=k1:v1'
curl -XGET 'es:9200/index_name/type_name/_search' -d '{
    "query": {
        "query_string": {
            "query": "k1:v1"
        }
    }
}'
```

### 分页，记录大小和文档版本值

```shell
# 从1220个文档开始，返回20个文档 <strong>不包含第1220个文档</strong>
# 返回文档的版本值
curl -XGET 'es:9200/index_name/type_name/_search' -d '{
    "from": 1220,
    "size": 20,
    "version": true,
    "query": {
        "query_string": {
            "query": "k1:v1"
        }
    }
}'
```

### 限制得分

```shell
# 只返回得分大于等于 0.75 的文档
curl -XGET 'es:9200/index_name/type_name/_search' -d '{
    "min_score": 0.75,
    "query": {
        "query_string" {
            "query": "k1:v1"
        }
    }
}'
```

### 选择要返回的字段

```shell
# 只返回 field1 field2 field3
curl -XGET 'es:9200/index_name/type_name/_search' -d '{
    "fields": [
        "field1",
        "field2",
        "field3"
    ],
    "query": {
        "query_string": {
            "query": "k1:v2"
        }
    }
}'
```

> 可以看到，一切按预期工作。与你分享以下3点：
>
> - 如果没有定义fields数组，它将用默认值，如果有就返回_source字段；_
> - 如果使用_source字段，并且请求一个没有存储的字段，那么这个字段将从_source字
>   段中提取（然而，这需要额外的处理）；
> - 如果想返回所有的存储字段，只需传入星号（*）作为字段名字。
> - 从性能的角度，返回_source字段比返回多个存储字段更好。


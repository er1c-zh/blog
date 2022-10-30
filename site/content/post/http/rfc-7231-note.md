---
title: "RFC7231笔记"
date: 2022-10-28T23:06:56+08:00
draft: false
tags:
    - http
    - rfc
    - memo
    - http-1_1-rfc
---

本文根据RFC文档的章节划分记录相关笔记，
是HTTP/1.1定义的的第二篇，
内容主题是语义 *(semantics)* 和内容 *(content)* 。

<!--more-->

{{% serial_index http-1_1-rfc %}}

# 1 简介

所有的HTTP报文要么是请求报文要么是响应报文。
服务端在链接上等待请求，
解析接受到的报文，
解释报文的语义，
生成一个或多个响应报文来响应请求。
客户端构建请求报文提出自己的期望，
检测收到的响应是否满足需要，
决定如何解释结果。

HTTP提供了与资源交互的统一的接口，
与类型、性质和实现无关。

HTTP语义由请求方法来定义意图，
可能由首部字段来扩充语义，
状态码表明机器友好的的响应结果，
响应的首部字段可能会包含其他的控制数据或资源元数据。

# 2 资源

HTTP请求的目标被称作资源。
协议不规定资源的属性，
仅仅定义与资源交互的接口。
资源由URI区分。

客户端发送请求时，
会将标识目标资源的URI按照RFC7230中的一个形式传入。
服务端接收到请求时，
重建有效的URI。

一个HTTP的设计目标的是将资源描述符与请求的语义解耦，
这通过请求方法和一些首部字段来实现。
如果请求方法与URI定义的语义冲突，
请求方法的语言更加优先。

# 3 表现 Representations

考虑到资源可以是任何形式的，
和HTTP协议提供的接口像是定义了可以观察可操作资源的窗户，
我们需要定义一个抽象的概念来描述资源的当前状态或需要（变成）的状态。
这个抽象就是表现。

根据HTTP协议的目标，
表现是一种信息，
满足：

- 能反应资源的过去、现在或需要的状态
- 格式可以容易的通过HTTP来传输
- 包含元数据和需要表现的数据

服务器有可能需要支持接受多种表现
或生成多种表现，
这个时候就需要服务器通过一些算法
选出最适合请求的表现，
这通常由内容协商机制来完成。
选中的表现用来为evaluating conditional requests提供数据和元数据，
和为GET请求的200或304响应构建载荷。

## 3.1 表现元数据

表现首部字段提供表现需要的元数据。
当报文包含载荷体时，
表现首部字段指明如何解析报文的载荷体中的表现数据。
对于HEAD请求的响应报文来说，
表现首部字段指明如何解析只有方法改为GET的请求的响应报文的载荷体中的表现数据。

### 3.1.1 处理表现数据

#### 3.1.1.1 Media Type

HTTP协议使用Internet media types在
`Content-Type`字段和`Accept`字段中
实现开放和可扩展的数据类型和类型协商机制。
Media type定义了数据格式和多种处理数据的模型。

Media Type形如`type "/" subtype *( OWS ";" OWS parameter )`，
type、subtype和parameter的name大小写不敏感。
parameter的value是否大小写敏感取决于parameter的语义。

parameter的value是否有引号没有关系。

#### 3.1.1.2 Charset

`Charset`用于协商文本类型的表现的字符编码格式。

`Charset`的值大小写不敏感。

#### 3.1.1.3 规范化与文本默认（格式？）

Internet media type在IANA注册时会有一个规范化的格式，
为了在多种编码格式的系统中迁移。
由于许多与与MIME相同的原因，
HTTP使用规范化的格式来传输表现。

与MIME不同的，
HTTP允许文本类型的media使用CR或LF或CRLF作为换行符。

#### 3.1.1.4 Multipart Types

MIME提供了数种“多部分”的类型，
将多个表现封装到同一个报文体中。
所有的多部分类型拥有同样的语法，
media type中有`boundary`参数 *（parameter）* 。

发送者使用CRLF作为消息体多个部分的之间的换行。

HTTP报文不使用`boundary`参数作为消息体长度。

#### 3.1.1.5 Content-Type

`Content-Type`首部字段指明关联的表现的**媒体类型** *（media type）* 。
媒体类型同时定义了数据格式和数据按照`Content-Encoding`解码后应该如何被接受者处理。

发送者发送带有载荷体的报文
SHOULD生成`Content-Type`字段，
除非发送者自己不知道。
如果接收到缺失`Content-Type`字段的报文，
接受者MAY假定是`application/octet-stream`类型，
或通过测试数据来决定类型。

实践中，资源的提供方，服务器，不总是能正确的设置`Content-Type`，
所以客户端有时会覆盖收到的报文中的`Content-Type`字段，
这会导致安全上的风险和其他的问题，如一种数据能满足多种`Content-Type`。

### 3.1.2 用于压缩或正确性 *(Integrity)* 的编码

#### 3.1.2.1 Content Codings

`Content coding`的值指出表现可以或已经被应用的转换编码。
通常，被编码过的表现只会在最后的接收者处解开。

`content-coding`的值大小写不敏感，
且应该是被注册的。

`content-coding`被用于`Accept-Encoding`字段
和`Content-Encoding`字段。

RFC7230定义了三个`content-coding`：

- `compress` `x-compress`
- `deflate`
- `gzip` `x-gzip`

#### 3.1.2.2 `Content-Encoding`

`Content-Encoding`指明表现已经应用的编码 *(content encoding)* ，
在按照`Content-Type`指明的媒体类型处理之前，
需要按照`Content-Encoding`解码。

可以连续应用多种编码，
发送者MUST将多种编码按照应用的顺序写入`Content-Encoding`中。

区别于`Transfer-Encoding`，
`Content-Encoding`是表现的属性。

### 3.1.3 Audience Language

### 3.1.4 Identification

## 3.2 Representation Data

表现用的数据要么放置在payload中，或者通过message语义和提供的URI来定位。

## 3.3 负载的语义 Payload Semantics

在请求中，payload的语义由方法决定。（PUT/POST/etc）

在响应中，payload的语义由方法和响应的状态码共同决定。

## 3.4 内容协商

对于同一个资源，
不同的设备或者用户，
可能有不同的最佳展现形式。
因此，HTTP协议提供了内容协商机制。

内容协商机制有两种形式：

- `proactive` server根据请求的header等信息来做出选择。
- `reactive` server提供一个展示信息的列表，由user agent来选择。

### 3.4.1 主动协商

### 3.4.2 被动协商

# 4 请求方法

## 4.1 Overview

request method表明了这个请求希望是什么含义。

大小写敏感。

```ditaa
ditaa
+---------+-------------------------------------------------+-------+
| Method  | Description                                     | Sec.  |
+---------+-------------------------------------------------+-------+
| GET     | Transfer a current representation of the target | 4.3.1 |
|         | resource.                                       |       |
| HEAD    | Same as GET, but only transfer the status line  | 4.3.2 |
|         | and header section.                             |       |
| POST    | Perform resource-specific processing on the     | 4.3.3 |
|         | request payload.                                |       |
| PUT     | Replace all current representations of the      | 4.3.4 |
|         | target resource with the request payload.       |       |
| DELETE  | Remove all current representations of the       | 4.3.5 |
|         | target resource.                                |       |
| CONNECT | Establish a tunnel to the server identified by  | 4.3.6 |
|         | the target resource.                            |       |
| OPTIONS | Describe the communication options for the      | 4.3.7 |
|         | target resource.                                |       |
| TRACE   | Perform a message loop-back test along the path | 4.3.8 |
|         | to the target resource.                         |       |
+---------+-------------------------------------------------+-------+
```

其他方法会在[IANA](https://www.iana.org/assignments/http-methods/http-methods.xhtml)注册。

如果server不支持某个方法，那么SHOULD返回501；
如果server不允许某个方法，那么SHOULD返回405。

## 4.2 通用方法属性

### 4.2.1 Safe Methods

如果方法的语义是只读的，那么称方法为safe的。

GET HEAD OPTIONS TRACE 是安全的。

区分是否安全的目的是消除爬虫或者缓存 pre-fetching 的顾虑。
另外，也允许user-agent针对潜在的风险作出限制。

资源的拥有者有义务来保证action和方法语义保持一致。

### 4.2.2 幂等方法

PUT DELETE 和 safe 方法是幂等方法。

### 4.2.3 缓存方法

如果一个方法是cacheable的，表明了响应可以被缓存和再次使用。

不依赖当前状态或者鉴权的安全方法的响应市可缓存的。
GET HEAD POST是可缓存的。

## 4.3 方法定义

### 4.3.1 GET

### 4.3.2 HEAD

### 4.3.3 POST

### 4.3.4 PUT

### 4.3.5 DELETE

### 4.3.6 CONNECT

CONNECT方法表明，
接收到这个方法请求的组件，需要建立一个tunnel到request-target指定的目标server。
接受者可以选择直接建立链接或者将connect转发到另一个server。

### 4.3.7 OPTIONS

### 4.3.8 TRACE

# 5 请求Header

TBD

# 6 响应状态码

- 1xx Informational: 请求被接受到，继续处理。
- 2xx Successful: 请求成功的收到、理解、接收。
- 3xx Redirection: 需要完成更多的操作来完成请求。
- 4xx Client Error: 请求有错误。
- 5xx Server Error: 服务端生成响应失败。

```ditaa
ditaa
+------+-------------------------------+--------------------------+
| Code | Reason-Phrase                 | Defined in...            |
+------+-------------------------------+--------------------------+
| 100  | Continue                      | Section 6.2.1            |
| 101  | Switching Protocols           | Section 6.2.2            |
| 200  | OK                            | Section 6.3.1            |
| 201  | Created                       | Section 6.3.2            |
| 202  | Accepted                      | Section 6.3.3            |
| 203  | Non-Authoritative Information | Section 6.3.4            |
| 204  | No Content                    | Section 6.3.5            |
| 205  | Reset Content                 | Section 6.3.6            |
| 206  | Partial Content               | Section 4.1 of [RFC7233] |
| 300  | Multiple Choices              | Section 6.4.1            |
| 301  | Moved Permanently             | Section 6.4.2            |
| 302  | Found                         | Section 6.4.3            |
| 303  | See Other                     | Section 6.4.4            |
| 304  | Not Modified                  | Section 4.1 of [RFC7232] |
| 305  | Use Proxy                     | Section 6.4.5            |
| 307  | Temporary Redirect            | Section 6.4.7            |
| 400  | Bad Request                   | Section 6.5.1            |
| 401  | Unauthorized                  | Section 3.1 of [RFC7235] |
| 402  | Payment Required              | Section 6.5.2            |
| 403  | Forbidden                     | Section 6.5.3            |
| 404  | Not Found                     | Section 6.5.4            |
| 405  | Method Not Allowed            | Section 6.5.5            |
| 406  | Not Acceptable                | Section 6.5.6            |
| 407  | Proxy Authentication Required | Section 3.2 of [RFC7235] |
| 408  | Request Timeout               | Section 6.5.7            |
| 409  | Conflict                      | Section 6.5.8            |
| 410  | Gone                          | Section 6.5.9            |
| 411  | Length Required               | Section 6.5.10           |
| 412  | Precondition Failed           | Section 4.2 of [RFC7232] |
| 413  | Payload Too Large             | Section 6.5.11           |
| 414  | URI Too Long                  | Section 6.5.12           |
| 415  | Unsupported Media Type        | Section 6.5.13           |
| 416  | Range Not Satisfiable         | Section 4.4 of [RFC7233] |
| 417  | Expectation Failed            | Section 6.5.14           |
| 426  | Upgrade Required              | Section 6.5.15           |
| 500  | Internal Server Error         | Section 6.6.1            |
| 501  | Not Implemented               | Section 6.6.2            |
| 502  | Bad Gateway                   | Section 6.6.3            |
| 503  | Service Unavailable           | Section 6.6.4            |
| 504  | Gateway Timeout               | Section 6.6.5            |
| 505  | HTTP Version Not Supported    | Section 6.6.6            |
+------+-------------------------------+--------------------------+
```

# 7 响应 Header fields

TBD

# 参考

- [rfc7231](https://datatracker.ietf.org/doc/html/rfc7231)

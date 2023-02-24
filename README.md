# picpac

自用的图片api，非常简陋，不建议直接使用

## build

```shell
go mod tidy
go build
```

## database

目前只支持PostgreSQL，使用其他数据库需要改代码。
按照pictures.sql创建即可

## config

config.json需要和二进制文件放在相同目录

### bind_addr

绑定的地址，格式 ip:port

### dsn

dataSourceName，数据库配置

### imgdir

存储图片的目录，支持绝对路径和相对路径

### url_prefix

图片地址的前缀，包括协议和ip地址或域名，结尾无斜线'/'，例如
```
http://x.x.x.x:8080
https://example.com
```

## usage

替换以下`127.0.0.1:11451`为实际地址



### 获取图片地址（这里的key就是图片的sha512）

```
GET http://127.0.0.1:11451/get/:key

Response: 
    Content-Type: text/plain
    Body: 如果图片存在，返回图片的url，图片不存在返回"not found"
```
### 上传图片

```
POST http://127.0.0.1:11451/upload

Request:
    form-data:
        name: "file"

Response: 
    Content-Type: text/plain
    Body: 如果上传成功，返回图片的sha512
```
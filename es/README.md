# Elasticsearch 适配器

> `es` 是 `go-logger-adapter` 的 Elasticsearch 日志适配器插件，将`go-logger` 的日志条目格式化为 NDJSON 格式，通过 Elasticsearch 的 `_bulk` API 批量写入 Elasticsearch

## 📖 概述

Elasticsearch 适配器将 `go-logger` 的日志条目格式化为 NDJSON 格式，通过 Elasticsearch 的 `_bulk` API 批量写入，支持自动索引命名、Ingest Pipeline、多种认证方式和 Gzip 压缩

## 🚀 核心特性

- **📦 Bulk API**: 使用 Elasticsearch Bulk API 批量写入，高效可靠
- **📅 自动索引**: 支持时间格式化的索引名称（如 `logs-2026.04.22`）
- **🔐 多认证方式**: 支持 Basic、API Key、Bearer Token 认证
- **🗜️ Gzip 压缩**: 支持 Gzip 压缩减少网络传输
- **🔄 自动重试**: 内置指数退避重试机制
- **🛡️ 背压控制**: 防止内存溢出
- **📊 统计信息**: 实时统计发送成功/失败/重试次数

## 📦 安装

```sh
go get -u github.com/kamalyes/go-logger-adapter/es
```

## 🔧 配置

### Config 结构体

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `Endpoints` | `[]string` | ES 节点地址列表 | 必填 |
| `IndexFormat` | `string` | 索引名称格式（支持时间格式化） | `logs-%s` |
| `Auth` | `adapter.AuthConfig` | 认证配置 | 无认证 |
| `TLS` | `*TLSConfig` | TLS 配置 | nil |
| `Pipeline` | `string` | Ingest Pipeline 名称 | 空 |
| `Common` | `*adapter.Config` | 通用引擎配置 | 默认配置 |

### TLSConfig 结构体

| 字段 | 类型 | 说明 |
|------|------|------|
| `InsecureSkipVerify` | `bool` | 跳过 TLS 证书验证 |
| `CACertFile` | `string` | CA 证书文件路径 |
| `CertFile` | `string` | 客户端证书文件路径 |
| `KeyFile` | `string` | 客户端私钥文件路径 |
| `ServerName` | `string` | 服务器名称 |

### 认证类型

| AuthType | 说明 | 所需字段 |
|----------|------|----------|
| `AuthNone` | 无认证 | - |
| `AuthBasic` | Basic 认证 | `Username`, `Password` |
| `AuthAPIKey` | API Key 认证 | `APIKey` |
| `AuthBearer` | Bearer Token 认证 | `BearerToken` |

## 🚀 使用示例

### 基础用法

```go
package main

import (
 "github.com/kamalyes/go-logger"
 "github.com/kamalyes/go-logger-adapter/es"
)

func main() {
 writer, err := es.NewWriter(&es.Config{
  Endpoints:   []string{"http://localhost:9200"},
  IndexFormat: "app-logs-%s",
 })
 if err != nil {
  panic(err)
 }

 log := logger.NewLogger().WithOutput(writer)
 log.Info("Hello Elasticsearch!")

 writer.Close()
}
```

### Basic 认证

```go
writer, err := es.NewWriter(&es.Config{
 Endpoints:   []string{"http://localhost:9200"},
 IndexFormat: "app-logs-%s",
 Auth: adapter.AuthConfig{
  AuthType: adapter.AuthBasic,
  Username: "elastic",
  Password: "changeme",
 },
})
```

### API Key 认证

```go
writer, err := es.NewWriter(&es.Config{
 Endpoints:   []string{"http://localhost:9200"},
 IndexFormat: "app-logs-%s",
 Auth: adapter.AuthConfig{
  AuthType: adapter.AuthAPIKey,
  APIKey:   "your-api-key",
 },
})
```

### 自定义通用配置

```go
writer, err := es.NewWriter(&es.Config{
 Endpoints:   []string{"http://localhost:9200"},
 IndexFormat: "app-logs-%s",
 Common: &adapter.Config{
  MaxBatchSize:   1024 * 1024, // 1MB
  MaxBatchCount:  1000,
  LingerMs:       5000,
  MaxRetries:     5,
  Compression:    adapter.CompressionGzip,
  MaxIoWorkers:   10,
  TotalSizeLimit: 50 * 1024 * 1024,
 },
})
```

### 使用 Ingest Pipeline

```go
writer, err := es.NewWriter(&es.Config{
 Endpoints:   []string{"http://localhost:9200"},
 IndexFormat: "app-logs-%s",
 Pipeline:    "my-pipeline",
})
```

## 📊 数据格式

日志条目会被格式化为 Elasticsearch Bulk API 所需的 NDJSON 格式：

```
{"index":{"_index":"app-logs-2026.04.22"}}
{"@timestamp":"2025-12-06T12:00:00.000Z","message":"Hello","log_level":"INFO"}
{"index":{"_index":"app-logs-2026.04.22"}}
{"@timestamp":"2025-12-06T12:00:01.000Z","message":"World","log_level":"DEBUG","user_id":123}
```

## 🏗️ 项目结构

```
es/
├── config.go   # 配置定义和校验
├── format.go   # Bulk API 格式化器
├── plugin.go   # Writer 和 Plugin 实现
└── README.md   # 本文档
```

## 许可证

该项目使用 MIT 许可证，详见 [LICENSE](../LICENSE) 文件

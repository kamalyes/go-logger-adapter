# OpenObserve 适配器

> `openobserve` 是 `go-logger-adapter` 的 OpenObserve 日志适配器插件，将`go-logger` 的日志条目格式化为 JSON 数组格式，通过 OpenObserve 的 `_json` Ingest API 批量写入 OpenObserve

## 📖 概述

OpenObserve 适配器将 `go-logger` 的日志条目格式化为 JSON 数组格式，通过 OpenObserve 的 `_json` Ingest API 批量写入，支持多组织隔离、多种认证方式和 Gzip 压缩

## 🚀 核心特性

- **📦 Ingest API**: 使用 OpenObserve JSON Ingest API 批量写入，高效可靠
- **🏢 多组织**: 支持组织 ID（zo-org-id）隔离
- **🔐 多认证方式**: 支持 Basic、API Key、Bearer Token 认证
- **🗜️ Gzip 压缩**: 支持 Gzip 压缩减少网络传输
- **🔄 自动重试**: 内置指数退避重试机制
- **🛡️ 背压控制**: 防止内存溢出
- **📊 统计信息**: 实时统计发送成功/失败/重试次数

## 📦 安装

```sh
go get -u github.com/kamalyes/go-logger-adapter/openobserve
```

## 🔧 配置

### Config 结构体

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `Endpoint` | `string` | OpenObserve 服务地址 | `http://localhost:5080` |
| `StreamName` | `string` | 日志流名称 | `default` |
| `OrgID` | `string` | 组织 ID | `default` |
| `Auth` | `adapter.AuthConfig` | 认证配置 | 无认证 |
| `Common` | `*adapter.Config` | 通用引擎配置 | 默认配置 |

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
 "github.com/kamalyes/go-logger-adapter/openobserve"
)

func main() {
 writer, err := openobserve.NewWriter(&openobserve.Config{
  Endpoint:   "http://localhost:5080",
  StreamName: "app-logs",
  OrgID:      "default",
 })
 if err != nil {
  panic(err)
 }

 log := logger.NewLogger().WithOutput(writer)
 log.Info("Hello OpenObserve!")

 writer.Close()
}
```

### Basic 认证

```go
writer, err := openobserve.NewWriter(&openobserve.Config{
 Endpoint:   "http://localhost:5080",
 StreamName: "app-logs",
 OrgID:      "my-org",
 Auth: adapter.AuthConfig{
  AuthType: adapter.AuthBasic,
  Username: "root@example.com",
  Password: "Complexpass#123",
 },
})
```

### 自定义通用配置

```go
writer, err := openobserve.NewWriter(&openobserve.Config{
 Endpoint:   "http://localhost:5080",
 StreamName: "app-logs",
 OrgID:      "default",
 Common: &adapter.Config{
  MaxBatchSize:   1024 * 1024,
  MaxBatchCount:  2000,
  LingerMs:       3000,
  MaxRetries:     5,
  Compression:    adapter.CompressionGzip,
  MaxIoWorkers:   10,
  TotalSizeLimit: 50 * 1024 * 1024,
 },
})
```

## 📊 数据格式

日志条目会被格式化为 OpenObserve Ingest API 所需的 JSON 数组格式：

```json
[
  {
    "_timestamp": "2025-12-06T12:00:00.000Z",
    "message": "Hello OpenObserve!",
    "level": "INFO"
  },
  {
    "_timestamp": "2025-12-06T12:00:01.000Z",
    "message": "User logged in",
    "level": "INFO",
    "user_id": 123
  }
]
```

### API 路径

数据会发送到以下路径：

```
POST /api/{org_id}/{stream_name}/_json
```

## 🏗️ 项目结构

```
openobserve/
├── config.go   # 配置定义、校验和数据格式化
├── plugin.go   # Writer 和 Plugin 实现
└── README.md   # 本文档
```

## 许可证

该项目使用 MIT 许可证，详见 [LICENSE](../LICENSE) 文件

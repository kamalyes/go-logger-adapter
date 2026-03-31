# Loki 适配器

> `loki` 是 `go-logger-adapter` 的 Grafana Loki 日志适配器插件，将`go-logger` 的日志条目格式化为 Loki Push API 所需的 JSON 格式，通过 Loki Push API 批量写入 Loki

## 📖 概述

Loki 适配器将 `go-logger` 的日志条目格式化为 Loki Push API 所需的 JSON 格式，支持流标签分组、结构化标签、多租户和 Gzip 压缩

## 🚀 核心特性

- **📦 Push API**: 使用 Loki Push API 批量写入，高效可靠
- **🏷️ 流标签**: 支持静态标签和动态日志级别标签
- **📋 结构化标签**: 支持将日志字段提取为 Loki 结构化标签
- **🏢 多租户**: 支持 X-Scope-OrgID 多租户隔离
- **🔐 多认证方式**: 支持 Basic、Bearer Token 认证
- **🗜️ Gzip 压缩**: 默认启用 Gzip 压缩减少网络传输
- **🔄 自动重试**: 内置指数退避重试机制
- **🛡️ 背压控制**: 防止内存溢出
- **📊 统计信息**: 实时统计发送成功/失败/重试次数

## 📦 安装

```sh
go get -u github.com/kamalyes/go-logger-adapter/loki
```

## 🔧 配置

### Config 结构体

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `Endpoint` | `string` | Loki 服务地址 | `http://localhost:3100` |
| `Labels` | `map[string]string` | 静态标签 | `{"job": "go-logger"}` |
| `TenantID` | `string` | 多租户 ID | 空 |
| `Auth` | `adapter.AuthConfig` | 认证配置 | 无认证 |
| `UseJSON` | `bool` | 使用 JSON 格式 | `true` |
| `Common` | `*adapter.Config` | 通用引擎配置 | 默认配置 |

### 认证类型

| AuthType | 说明 | 所需字段 |
|----------|------|----------|
| `AuthNone` | 无认证 | - |
| `AuthBasic` | Basic 认证 | `Username`, `Password` |
| `AuthBearer` | Bearer Token 认证 | `BearerToken` |

## 🚀 使用示例

### 基础用法

```go
package main

import (
 "github.com/kamalyes/go-logger"
 "github.com/kamalyes/go-logger-adapter/loki"
)

func main() {
 writer, err := loki.NewWriter(&loki.Config{
  Endpoint: "http://localhost:3100",
  Labels:   map[string]string{"job": "my-app", "env": "production"},
 })
 if err != nil {
  panic(err)
 }

 log := logger.NewLogger().WithOutput(writer)
 log.Info("Hello Loki!")

 writer.Close()
}
```

### 多租户

```go
writer, err := loki.NewWriter(&loki.Config{
 Endpoint: "http://localhost:3100",
 Labels:   map[string]string{"job": "my-app"},
 TenantID: "my-organization",
})
```

### Basic 认证

```go
writer, err := loki.NewWriter(&loki.Config{
 Endpoint: "http://localhost:3100",
 Labels:   map[string]string{"job": "my-app"},
 Auth: adapter.AuthConfig{
  AuthType: adapter.AuthBasic,
  Username: "admin",
  Password: "secret",
 },
})
```

### 自定义通用配置

```go
writer, err := loki.NewWriter(&loki.Config{
 Endpoint: "http://localhost:3100",
 Labels:   map[string]string{"job": "my-app"},
 Common: &adapter.Config{
  MaxBatchSize:   512 * 1024,
  MaxBatchCount:  2000,
  LingerMs:       3000,
  MaxRetries:     8,
  Compression:    adapter.CompressionGzip,
  MaxIoWorkers:   20,
  TotalSizeLimit: 100 * 1024 * 1024,
 },
})
```

## 📊 数据格式

日志条目会被格式化为 Loki Push API 所需的 JSON 格式：

```json
{
  "streams": [
    {
      "stream": {
        "job": "my-app",
        "level": "INFO"
      },
      "values": [
        ["1713772800000000000", "Hello Loki!"],
        ["1713772801000000000", "Another log", {"user_id": "123"}]
      ]
    }
  ]
}
```

### 流标签分组

相同标签组合的日志条目会被自动分组到同一个 stream 中，减少推送请求的数量

### 结构化标签

当日志条目包含 Fields 时，字段会被提取为 Loki 结构化标签（structured metadata），支持在 LogQL 中进行高效查询

## 🏗️ 项目结构

```
loki/
├── config.go   # 配置定义、校验和数据格式化
├── plugin.go   # Writer 和 Plugin 实现
└── README.md   # 本文档
```

## 许可证

该项目使用 MIT 许可证，详见 [LICENSE](../LICENSE) 文件

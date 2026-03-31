# VictoriaLogs 适配器

> `victorialogs` 是 `go-logger-adapter` 的 VictoriaLogs 日志适配器插件，将`go-logger` 的日志条目格式化为 JSON Line 格式，通过 VictoriaLogs 的 `/insert/jsonline` API 批量写入，支持多租户隔离、多种认证方式和 Gzip 压缩

## 🚀 核心特性

- **📦 Insert API**: 使用 VictoriaLogs JSON Line Insert API 批量写入，高效可靠
- **🏢 多租户**: 支持 AccountID 租户隔离
- **🔐 多认证方式**: 支持 Basic、Bearer Token 认证
- **🗜️ Gzip 压缩**: 支持 Gzip 压缩减少网络传输
- **🔄 自动重试**: 内置指数退避重试机制
- **🛡️ 背压控制**: 防止内存溢出
- **📊 统计信息**: 实时统计发送成功/失败/重试次数

## 📦 安装

```sh
go get -u github.com/kamalyes/go-logger-adapter/victorialogs
```

## 🔧 配置

### Config 结构体

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `Endpoint` | `string` | VictoriaLogs 服务地址 | `http://localhost:9428` |
| `TenantID` | `string` | 租户 ID（AccountID:ProjectID 格式） | 空 |
| `Auth` | `adapter.AuthConfig` | 认证配置 | 无认证 |
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
 "github.com/kamalyes/go-logger-adapter/victorialogs"
)

func main() {
 writer, err := victorialogs.NewWriter(&victorialogs.Config{
  Endpoint: "http://localhost:9428",
 })
 if err != nil {
  panic(err)
 }

 log := logger.NewLogger().WithOutput(writer)
 log.Info("Hello VictoriaLogs!")

 writer.Close()
}
```

### 多租户

```go
writer, err := victorialogs.NewWriter(&victorialogs.Config{
 Endpoint: "http://localhost:9428",
 TenantID: "12:34", // AccountID:ProjectID 格式
})
```

### Basic 认证

```go
writer, err := victorialogs.NewWriter(&victorialogs.Config{
 Endpoint: "http://localhost:9428",
 Auth: adapter.AuthConfig{
  AuthType: adapter.AuthBasic,
  Username: "admin",
  Password: "secret",
 },
})
```

### 自定义通用配置

```go
writer, err := victorialogs.NewWriter(&victorialogs.Config{
 Endpoint: "http://localhost:9428",
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

日志条目会被格式化为 VictoriaLogs JSON Line 格式（每行一条 JSON）：

```json
{"_ts":"2025-12-06T12:00:00.000Z","message":"Hello VictoriaLogs!","_level":"INFO"}
{"_ts":"2025-12-06T12:00:01.000Z","message":"User logged in","_level":"INFO","user_id":123}
```

### API 路径

数据会发送到以下路径：

```
POST /insert/jsonline
```

### 字段映射

| LogEntry 字段 | VictoriaLogs 字段 | 说明 |
|---------------|-------------------|------|
| `Timestamp` | `_ts` | 时间戳（RFC3339Nano 格式） |
| `Level` | `_level` | 日志级别 |
| `Message` | `message` | 日志消息 |
| `Fields` | 原始键名 | 自定义字段 |

## 🏗️ 项目结构

```
victorialogs/
├── config.go   # 配置定义、校验和数据格式化
├── plugin.go   # Writer 和 Plugin 实现
└── README.md   # 本文档
```

## 许可证

该项目使用 MIT 许可证，详见 [LICENSE](../LICENSE) 文件

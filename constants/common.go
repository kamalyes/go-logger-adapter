/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\constants\common.go
 * @Description: 全局常量定义 - 引擎默认值、认证类型、压缩类型、HTTP 状态码等
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package constants

import "time"

// ==================== 认证相关常量 ====================

const (
	AuthNone   = iota // 无认证
	AuthBasic         // Basic 认证
	AuthAPIKey        // API Key 认证
	AuthBearer        // Bearer Token 认证
	AuthToken         // 自定义 Token 认证
)

// ==================== 压缩相关常量 ====================

const (
	CompressionNone = iota // 无压缩
	CompressionGzip        // Gzip 压缩
	CompressionZlib        // Zlib 压缩
)

// ==================== 引擎默认值常量 ====================

const (
	DefaultMaxBatchSize   = 512 * 1024        // 单个批次最大字节数 (512KB)
	DefaultMaxBatchCount  = 4096              // 单个批次最大条目数
	DefaultLingerMs       = int64(2000)       // 批次等待时间（毫秒）
	DefaultMaxRetries     = int64(10)         // 最大重试次数
	DefaultBaseRetryMs    = int64(100)        // 基础重试间隔（毫秒）
	DefaultMaxRetryMs     = int64(50000)      // 最大重试间隔（毫秒）
	DefaultMaxIoWorkers   = int64(50)         // I/O 工作协程数
	DefaultTotalSizeLimit = 100 * 1024 * 1024 // 内存总量限制 (100MB)
	DefaultMaxBlockSec    = int64(60)         // 背压最大阻塞时间（秒）
	DefaultMaxIdleConns   = 100               // HTTP 最大空闲连接数
	DefaultHealthTimeoutX = 5                 // 健康检查超时倍数
)

var (
	DefaultNoRetryStatusCodes = []int{400, 404} // 不重试的 HTTP 状态码
)

// ==================== 时间相关默认值 ====================

var (
	DefaultRequestTimeout = 30 * time.Second // 请求超时时间
	DefaultFlushInterval  = 5 * time.Second  // 刷新间隔
)

// ==================== 背压控制常量 ====================

const DefaultBackPressureSleep = 10 * time.Millisecond // 背压等待时的短睡眠间隔

// ==================== 工作池常量 ====================

const DefaultWorkerPoolQueueSize = 10000 // WorkerPool 默认队列大小

// ==================== HTTP 状态码常量 ====================

const (
	HTTPStatusOK          = 200 // 请求成功
	HTTPStatusNoContent   = 204 // 无内容
	HTTPStatusBadRequest  = 400 // 错误请求
	HTTPStatusNotFound    = 404 // 未找到
	HTTPStatusTooManyReq  = 429 // 请求过多
	HTTPStatusServerError = 300 // 服务端错误起始码（>=300 视为错误）
	HTTPStatus5xxStart    = 500 // 5xx 错误起始码
)

// ==================== 插件默认端点常量 ====================

const (
	DefaultElasticsearchEndpoint = "http://localhost:9200" // Elasticsearch 默认端点
	DefaultLokiEndpoint          = "http://localhost:3100" // Loki 默认端点
	DefaultOpenObserveEndpoint   = "http://localhost:5080" // OpenObserve 默认端点
	DefaultVictoriaLogsEndpoint  = "http://localhost:9428" // VictoriaLogs 默认端点
)

// ==================== 插件默认配置常量 ====================

const (
	DefaultESIndexFormat     = "logs-%s"   // Elasticsearch 默认索引格式
	DefaultLokiJobLabel      = "go-logger" // Loki 默认 job 标签
	DefaultOpenObserveStream = "default"   // OpenObserve 默认流名称
	DefaultOpenObserveOrgID  = "default"   // OpenObserve 默认组织 ID
)

// ==================== 时间戳格式常量 ====================

const (
	ESIndexDateFormat       = "2006.01.02" // Elasticsearch 索引日期格式
	TimestampMsDivisor      = int64(1000)  // 毫秒时间戳转秒的除数
	TimestampNsMultiplier   = int64(1e6)   // 毫秒转纳秒的乘数
	LokiTimestampMultiplier = int64(1e6)   // 毫秒转纳秒的乘数（Loki）
)

// ==================== 内容类型映射 ====================

const (
	ContentTypeKey    = "Content-Type"         // Content-Type 头键
	ContentTypeJSON   = "application/json"     // JSON 内容类型
	ContentTypeNDJSON = "application/x-ndjson" // NDJSON 内容类型
)

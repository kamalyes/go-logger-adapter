/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\config.go
 * @Description: 引擎配置定义和默认值
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"net/http"
	"time"

	"github.com/kamalyes/go-logger-adapter/constants"
)

// Config 引擎配置参数
type Config struct {
	MaxBatchSize       int64         // 单个批次最大字节数
	MaxBatchCount      int           // 单个批次最大条目数
	LingerMs           int64         // 批次等待时间（毫秒）
	MaxRetries         int64         // 最大重试次数
	BaseRetryMs        int64         // 基础重试间隔（毫秒）
	MaxRetryMs         int64         // 最大重试间隔（毫秒）
	MaxIoWorkers       int64         // I/O 工作协程数
	TotalSizeLimit     int64         // 内存总量限制（字节）
	MaxBlockSec        int64         // 背压最大阻塞时间（秒）
	Compression        int           // 压缩类型（0=无压缩, 1=Gzip, 2=Zlib）
	RequestTimeout     time.Duration // 请求超时时间
	FlushInterval      time.Duration // 刷新间隔
	NoRetryStatusCodes []int         // 不重试的 HTTP 状态码
	HTTPClient         *http.Client  // 自定义 HTTP 客户端
}

// DefaultConfig 返回默认引擎配置
func DefaultConfig() *Config {
	return &Config{
		MaxBatchSize:       constants.DefaultMaxBatchSize,
		MaxBatchCount:      constants.DefaultMaxBatchCount,
		LingerMs:           constants.DefaultLingerMs,
		MaxRetries:         constants.DefaultMaxRetries,
		BaseRetryMs:        constants.DefaultBaseRetryMs,
		MaxRetryMs:         constants.DefaultMaxRetryMs,
		MaxIoWorkers:       constants.DefaultMaxIoWorkers,
		TotalSizeLimit:     constants.DefaultTotalSizeLimit,
		MaxBlockSec:        constants.DefaultMaxBlockSec,
		Compression:        constants.CompressionNone,
		RequestTimeout:     constants.DefaultRequestTimeout,
		FlushInterval:      constants.DefaultFlushInterval,
		NoRetryStatusCodes: constants.DefaultNoRetryStatusCodes,
	}
}

// Option 引擎配置选项函数
type Option func(*Config)

// WithBatchSize 设置单个批次最大字节数
func WithBatchSize(size int64) Option {
	return func(c *Config) { c.MaxBatchSize = size }
}

// WithBatchCount 设置单个批次最大条目数
func WithBatchCount(count int) Option {
	return func(c *Config) { c.MaxBatchCount = count }
}

// WithLingerMs 设置批次等待时间
func WithLingerMs(ms int64) Option {
	return func(c *Config) { c.LingerMs = ms }
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(retries int64) Option {
	return func(c *Config) { c.MaxRetries = retries }
}

// WithWorkers 设置 I/O 工作协程数
func WithWorkers(workers int64) Option {
	return func(c *Config) { c.MaxIoWorkers = workers }
}

// WithCompression 设置压缩类型
func WithCompression(compression int) Option {
	return func(c *Config) { c.Compression = compression }
}

// WithTotalSizeLimit 设置内存总量限制
func WithTotalSizeLimit(limit int64) Option {
	return func(c *Config) { c.TotalSizeLimit = limit }
}

// WithRequestTimeout 设置请求超时时间
func WithRequestTimeout(timeout time.Duration) Option {
	return func(c *Config) { c.RequestTimeout = timeout }
}

// WithNoRetryStatusCodes 设置不重试的 HTTP 状态码
func WithNoRetryStatusCodes(codes ...int) Option {
	return func(c *Config) { c.NoRetryStatusCodes = codes }
}

// WithHTTPClient 设置自定义 HTTP 客户端
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) { c.HTTPClient = client }
}

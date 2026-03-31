/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\base_writer.go
 * @Description: 基础写入器实现，消除各插件 Writer 的重复代码
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"sync/atomic"
	"time"

	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/mathx"

	logger "github.com/kamalyes/go-logger"
)

// BaseWriter 提供各插件 Writer 的公共实现
// 各插件（es/loki/openobserve/victorialogs）可嵌入此结构体
// 避免重复实现 Start/Flush/Close/Write/WriteLevel 等方法
type BaseWriter struct {
	engine  *Engine
	healthy int32
}

// NewBaseWriter 创建基础写入器实例
func NewBaseWriter(engine *Engine) *BaseWriter {
	return &BaseWriter{
		engine:  engine,
		healthy: 1,
	}
}

// Start 启动写入器
func (w *BaseWriter) Start() { w.engine.Start() }

// Flush 刷新缓冲区
func (w *BaseWriter) Flush() error { return w.engine.Flush() }

// IsHealthy 检查写入器健康状态
func (w *BaseWriter) IsHealthy() bool {
	return atomic.LoadInt32(&w.healthy) == 1 && w.engine.IsHealthy()
}

// GetStats 获取写入器统计信息快照
func (w *BaseWriter) GetStats() logger.WriterStatsSnapshot {
	return w.engine.WriterStats()
}

// Write 实现 io.Writer 接口，默认使用 INFO 级别
func (w *BaseWriter) Write(p []byte) (n int, err error) {
	return w.WriteLevel(logger.INFO, p)
}

// WriteLevel 按指定级别写入日志数据
func (w *BaseWriter) WriteLevel(level logger.LogLevel, data []byte) (n int, err error) {
	entry := logger.LogEntry{
		Level:     level,
		Message:   string(data),
		Timestamp: time.Now().UnixMilli(),
	}
	if err := w.engine.Submit(entry, nil); err != nil {
		return 0, err
	}
	return len(data), nil
}

// Close 关闭写入器，依次刷新缓冲区并停止引擎
func (w *BaseWriter) Close() error {
	w.engine.Flush()
	w.engine.Stop()
	atomic.StoreInt32(&w.healthy, 0)
	return nil
}

// WriteEntry 写入单条日志条目
func (w *BaseWriter) WriteEntry(entry logger.LogEntry) error {
	return w.engine.Submit(entry, nil)
}

// WriteEntryWithCallback 写入日志条目并设置回调
func (w *BaseWriter) WriteEntryWithCallback(entry logger.LogEntry, cb Callback) error {
	return w.engine.Submit(entry, cb)
}

// Engine 获取底层引擎实例
func (w *BaseWriter) Engine() *Engine {
	return w.engine
}

// SetHealthy 设置健康状态
func (w *BaseWriter) SetHealthy(healthy bool) {
	atomic.StoreInt32(&w.healthy, mathx.IF(healthy, int32(1), int32(0)))
}

// CommonAdapterOpts 从通用配置生成 Engine 选项列表
// 各插件可复用此函数，避免重复构建 opts
func CommonAdapterOpts(common *Config) []Option {
	opts := []Option{
		WithBatchSize(common.MaxBatchSize),
		WithBatchCount(common.MaxBatchCount),
		WithLingerMs(common.LingerMs),
		WithMaxRetries(common.MaxRetries),
		WithWorkers(common.MaxIoWorkers),
		WithCompression(common.Compression),
		WithTotalSizeLimit(common.TotalSizeLimit),
		WithRequestTimeout(common.RequestTimeout),
		WithNoRetryStatusCodes(common.NoRetryStatusCodes...),
	}
	if common.HTTPClient != nil {
		opts = append(opts, WithHTTPClient(common.HTTPClient))
	}
	return opts
}

// EnsureCommonDefaults 确保通用配置的默认值已设置
// 各插件的 SetCommonDefaults 可复用此函数
func EnsureCommonDefaults(common **Config) {
	if *common == nil {
		*common = DefaultConfig()
	}
	(*common).RequestTimeout = mathx.IfLeZero((*common).RequestTimeout, constants.DefaultRequestTimeout)
	(*common).FlushInterval = mathx.IfLeZero((*common).FlushInterval, constants.DefaultFlushInterval)
}

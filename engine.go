/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\engine.go
 * @Description: 日志适配器引擎，协调批处理、重试、压缩和发送流程
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"context"
	"github.com/kamalyes/go-logger-adapter/constants"
	logger "github.com/kamalyes/go-logger"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"github.com/kamalyes/go-toolbox/pkg/httpx"
	"github.com/kamalyes/go-toolbox/pkg/mathx"
	"sync"
	"sync/atomic"
	"time"
)

// Engine 日志适配器引擎，负责协调各组件完成日志的批处理、压缩、发送和重试
type Engine struct {
	mu           sync.RWMutex  // 读写锁，保护并发访问
	plugin       Plugin        // 日志插件（es/loki/openobserve/victorialogs）
	healthy      int32         // 健康状态标识（atomic: 0=不健康, 1=健康）
	stats        *Stats        // 统计信息收集器
	backPressure *BackPressure // 背压控制器
	accumulator  *Accumulator  // 日志条目累加器
	retryQueue   *RetryQueue   // 重试队列
	mover        *Mover        // 数据搬运器（定时触发批处理）
	pool         *WorkerPool   // I/O 工作池
	compressor   Compressor    // 数据压缩器
	config       *Config       // 引擎配置
	client       *httpx.Client // HTTP 客户端
}

// NewEngine 创建日志适配器引擎实例
// plugin 参数为必填，opts 为可选配置项
func NewEngine(plugin Plugin, opts ...Option) (*Engine, error) {
	if plugin == nil {
		return nil, errorx.NewTypedError(ErrTypePluginRequired, ErrMsgPluginRequired)
	}

	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	stats := NewStats()
	backPressure := NewBackPressure(cfg.TotalSizeLimit, cfg.MaxBlockSec)
	accumulator := NewAccumulator(cfg, backPressure, stats)
	retryQueue := NewRetryQueue()

	e := &Engine{
		plugin:       plugin,
		stats:        stats,
		backPressure: backPressure,
		accumulator:  accumulator,
		retryQueue:   retryQueue,
		config:       cfg,
		compressor:   NewCompressor(cfg.Compression),
		healthy:      1,
	}

	e.client = mathx.IfDo(cfg.HTTPClient != nil,
		func() *httpx.Client { return httpx.NewHttpClient(cfg.HTTPClient) },
		httpx.NewClient(
			httpx.WithTimeout(cfg.RequestTimeout),
			httpx.WithMaxIdleConnsPerHost(constants.DefaultMaxIdleConns),
		),
	)

	wrappedSend := func(batch *Batch) error {
		return e.processBatch(batch)
	}

	e.pool = NewWorkerPool(wrappedSend, cfg.MaxIoWorkers, stats)
	e.mover = NewMover(accumulator, retryQueue, e.pool, stats, cfg.LingerMs)

	return e, nil
}

// Start 启动引擎，开始定时搬运日志
func (e *Engine) Start() {
	e.mover.Start()
}

// Submit 提交一条日志条目到累加器
func (e *Engine) Submit(entry LogEntry, cb Callback) error {
	size := int64(estimateEntrySize(entry))
	return e.accumulator.AddEntry("default", entry, size, cb)
}

// processBatch 处理一个批次的日志数据
// 流程：格式化 → 压缩 → 发送 → 失败重试
func (e *Engine) processBatch(batch *Batch) error {
	begin := time.Now()
	attempt := batch.IncrementAttempt()

	if !batch.CanRetry() && attempt > 1 {
		batch.OnFail("max_retries_exceeded", ErrMsgMaxRetriesExceeded, begin)
		e.backPressure.Release(batch.DataSize())
		return errorx.NewTypedError(ErrTypeMaxRetriesExceeded, ErrMsgMaxRetriesExceeded)
	}

	entries := batch.GetEntries()
	if len(entries) == 0 {
		e.backPressure.Release(batch.DataSize())
		return nil
	}

	body, err := e.plugin.Format(entries)
	if err != nil {
		batch.OnFail("format_error", err.Error(), begin)
		e.backPressure.Release(batch.DataSize())
		e.stats.IncSendErrors()
		return errorx.NewTypedError(ErrTypeFormatFailed, ErrFmtFormatFailed, e.plugin.Name(), err)
	}

	if e.compressor != nil {
		body, err = e.compressor.Compress(body)
		if err != nil {
			batch.OnFail("compress_error", err.Error(), begin)
			e.backPressure.Release(batch.DataSize())
			e.stats.IncSendErrors()
			return errorx.NewTypedError(ErrTypeCompressFailed, ErrFmtCompressFailed, e.plugin.Name(), err)
		}
	}

	headers := map[string]string{
		constants.HeaderContentType: e.contentType(),
	}
	if e.compressor != nil {
		headers[constants.HeaderContentEncoding] = e.compressor.ContentEncoding()
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.config.RequestTimeout)
	defer cancel()

	err = e.plugin.Send(ctx, body, headers)
	if err != nil {
		isRetryable := true
		if httpErr, ok := err.(*HTTPError); ok {
			isRetryable = IsRetryable(httpErr.StatusCode, e.config.NoRetryStatusCodes)
		}

		if isRetryable {
			batch.CalculateNextRetryMs()
			e.retryQueue.Enqueue(batch)
			e.stats.IncRetryCount()
			return err
		}

		batch.OnFail("send_failed", err.Error(), begin)
		e.backPressure.Release(batch.DataSize())
		e.stats.IncSendErrors()
		return err
	}

	batch.OnSuccess(begin)
	e.backPressure.Release(batch.DataSize())
	e.stats.IncSendSuccess()
	e.stats.SetLastSend(time.Now())
	return nil
}

// contentType 根据插件名称返回对应的 Content-Type
func (e *Engine) contentType() string {
	switch e.plugin.Name() {
	case constants.PluginNameElasticsearch, constants.PluginNameVictoriaLogs:
		return httpx.ContentTypeApplicationXNDJSON
	case constants.PluginNameLoki, constants.PluginNameOpenObserve:
		return httpx.ContentTypeApplicationJSON
	default:
		return httpx.ContentTypeApplicationJSON
	}
}

// Stop 停止引擎，关闭搬运器和工作池
func (e *Engine) Stop() {
	e.mover.Stop()
	e.pool.Shutdown()
	atomic.StoreInt32(&e.healthy, 0)
}

// Flush 刷新所有缓冲的日志数据
func (e *Engine) Flush() error {
	e.mover.Flush()
	return nil
}

// Close 关闭引擎，依次刷新并停止
func (e *Engine) Close() error {
	e.Flush()
	e.Stop()
	return nil
}

// IsHealthy 检查引擎健康状态
func (e *Engine) IsHealthy() bool {
	return atomic.LoadInt32(&e.healthy) == 1
}

// Stats 获取详细的统计信息快照
func (e *Engine) Stats() StatsSnapshot {
	return e.stats.DetailedSnapshot()
}

// WriterStats 获取写入器统计信息快照（兼容 go-logger 接口）
func (e *Engine) WriterStats() logger.WriterStatsSnapshot {
	return e.stats.Snapshot()
}

// HTTPClient 获取 HTTP 客户端实例
func (e *Engine) HTTPClient() *httpx.Client {
	return e.client
}

// Plugin 获取当前插件实例
func (e *Engine) Plugin() Plugin {
	return e.plugin
}

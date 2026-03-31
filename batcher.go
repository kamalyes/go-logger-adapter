/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\batcher.go
 * @Description: 日志批次管理和累加器实现
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/kamalyes/go-logger"
)

// LogEntry 日志条目类型别名，与 go-logger 的 LogEntry 保持一致
type LogEntry = logger.LogEntry

// Callback 发送结果回调函数
type Callback func(result *SendResult)

// SendResult 发送结果
type SendResult struct {
	Success      bool          // 是否发送成功
	AttemptCount int           // 尝试次数
	ErrorCode    string        // 错误码
	ErrorMessage string        // 错误消息
	BatchSize    int           // 批次大小
	Duration     time.Duration // 发送耗时
}

// Batch 日志批次，聚合多条日志条目统一发送
type Batch struct {
	mu            sync.Mutex // 保护并发访问
	entries       []LogEntry // 日志条目列表
	totalDataSize int64      // 总数据大小（字节）
	createTimeMs  int64      // 创建时间（毫秒时间戳）
	attemptCount  int        // 发送尝试次数
	nextRetryMs   int64      // 下次重试时间（毫秒时间戳）
	callbacks     []Callback // 回调函数列表
	maxRetries    int64      // 最大重试次数
	baseRetryMs   int64      // 基础重试间隔（毫秒）
	maxRetryMs    int64      // 最大重试间隔（毫秒）
	batchKey      string     // 批次键
}

// NewBatch 创建新的日志批次
func NewBatch(key string, maxBatchCount int, maxRetries, baseRetryMs, maxRetryMs int64) *Batch {
	return &Batch{
		entries:      make([]LogEntry, 0, maxBatchCount),
		createTimeMs: time.Now().UnixMilli(),
		batchKey:     key,
		maxRetries:   maxRetries,
		baseRetryMs:  baseRetryMs,
		maxRetryMs:   maxRetryMs,
	}
}

// AddEntry 添加一条日志条目到批次
func (b *Batch) AddEntry(entry LogEntry, size int64, cb Callback) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = append(b.entries, entry)
	b.totalDataSize += size
	if cb != nil {
		b.callbacks = append(b.callbacks, cb)
	}
}

// IsFull 判断批次是否已满
func (b *Batch) IsFull(maxBatchSize int64, maxBatchCount int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.totalDataSize >= maxBatchSize || len(b.entries) >= maxBatchCount
}

// GetEntries 获取批次中的所有日志条目（返回副本）
func (b *Batch) GetEntries() []LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	result := make([]LogEntry, len(b.entries))
	copy(result, b.entries)
	return result
}

// EntryCount 获取批次中的条目数量
func (b *Batch) EntryCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.entries)
}

// DataSize 获取批次的总数据大小
func (b *Batch) DataSize() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.totalDataSize
}

// IsExpired 判断批次是否已过期
func (b *Batch) IsExpired(lingerMs int64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return time.Now().UnixMilli()-b.createTimeMs >= lingerMs
}

// OnSuccess 处理发送成功，触发所有回调
func (b *Batch) OnSuccess(begin time.Time) {
	result := &SendResult{
		Success:      true,
		AttemptCount: b.attemptCount,
		BatchSize:    len(b.entries),
		Duration:     time.Since(begin),
	}
	for _, cb := range b.callbacks {
		cb(result)
	}
}

// OnFail 处理发送失败，触发所有回调
func (b *Batch) OnFail(errCode, errMsg string, begin time.Time) {
	result := &SendResult{
		Success:      false,
		AttemptCount: b.attemptCount,
		ErrorCode:    errCode,
		ErrorMessage: errMsg,
		BatchSize:    len(b.entries),
		Duration:     time.Since(begin),
	}
	for _, cb := range b.callbacks {
		cb(result)
	}
}

// IncrementAttempt 增加尝试次数并返回当前次数
func (b *Batch) IncrementAttempt() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attemptCount++
	return b.attemptCount
}

// CanRetry 判断是否还可以重试
func (b *Batch) CanRetry() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return int64(b.attemptCount) < b.maxRetries
}

// CalculateNextRetryMs 计算下次重试时间（指数退避）
func (b *Batch) CalculateNextRetryMs() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	backoff := b.baseRetryMs
	for i := 1; i < b.attemptCount; i++ {
		backoff *= 2
		if backoff > b.maxRetryMs {
			backoff = b.maxRetryMs
			break
		}
	}
	b.nextRetryMs = time.Now().UnixMilli() + backoff
	return b.nextRetryMs
}

// GetNextRetryMs 获取下次重试时间
func (b *Batch) GetNextRetryMs() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.nextRetryMs
}

// Accumulator 日志条目累加器，负责收集日志并组织成批次
type Accumulator struct {
	mu           sync.Mutex        // 保护并发访问
	batches      map[string]*Batch // 批次映射
	config       *Config           // 引擎配置
	totalSize    int64             // 当前总内存占用
	backPressure *BackPressure     // 背压控制器
	stats        *Stats            // 统计信息
}

// NewAccumulator 创建日志累加器
func NewAccumulator(config *Config, backPressure *BackPressure, stats *Stats) *Accumulator {
	return &Accumulator{
		batches:      make(map[string]*Batch),
		config:       config,
		backPressure: backPressure,
		stats:        stats,
	}
}

// AddEntry 添加一条日志条目到累加器
func (a *Accumulator) AddEntry(key string, entry LogEntry, size int64, cb Callback) error {
	if a.backPressure != nil {
		if err := a.backPressure.Wait(size); err != nil {
			return err
		}
	}

	a.mu.Lock()
	batch, exists := a.batches[key]
	if !exists || batch == nil {
		batch = NewBatch(key, a.config.MaxBatchCount, a.config.MaxRetries, a.config.BaseRetryMs, a.config.MaxRetryMs)
		a.batches[key] = batch
	}
	batch.AddEntry(entry, size, cb)
	atomic.AddInt64(&a.totalSize, size)

	if a.stats != nil {
		a.stats.IncTotalEntries()
		a.stats.AddTotalBytes(size)
	}

	shouldSubmit := batch.IsFull(a.config.MaxBatchSize, a.config.MaxBatchCount)
	if shouldSubmit {
		a.batches[key] = nil
	}
	a.mu.Unlock()

	if shouldSubmit {
		atomic.AddInt64(&a.totalSize, -batch.DataSize())
		return a.submitBatch(batch)
	}
	return nil
}

// RemoveBatch 移除并返回指定键的批次
func (a *Accumulator) RemoveBatch(key string) *Batch {
	a.mu.Lock()
	defer a.mu.Unlock()
	batch := a.batches[key]
	a.batches[key] = nil
	return batch
}

// ExpireBatches 获取所有已过期的批次
func (a *Accumulator) ExpireBatches(lingerMs int64) []*Batch {
	a.mu.Lock()
	defer a.mu.Unlock()

	var expired []*Batch
	now := time.Now().UnixMilli()
	for key, batch := range a.batches {
		if batch == nil {
			delete(a.batches, key)
			continue
		}
		if now-batch.createTimeMs >= lingerMs {
			expired = append(expired, batch)
			a.batches[key] = nil
		}
	}
	return expired
}

// FlushAll 获取所有批次并清空累加器
func (a *Accumulator) FlushAll() []*Batch {
	a.mu.Lock()
	defer a.mu.Unlock()

	var all []*Batch
	for key, batch := range a.batches {
		if batch != nil && batch.EntryCount() > 0 {
			all = append(all, batch)
		}
		delete(a.batches, key)
	}
	a.batches = make(map[string]*Batch)
	return all
}

// TotalSize 获取当前总内存占用
func (a *Accumulator) TotalSize() int64 {
	return atomic.LoadInt64(&a.totalSize)
}

// submitBatch 提交批次（占位方法，实际由 Mover 驱动）
func (a *Accumulator) submitBatch(batch *Batch) error {
	return nil
}

// estimateEntrySize 估算日志条目的字节大小
func estimateEntrySize(entry LogEntry) int {
	data, err := json.Marshal(entry)
	if err != nil {
		return len(entry.Message) + 256
	}
	return len(data) + 64
}

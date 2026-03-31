/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\stats.go
 * @Description: 统计信息收集器，使用 atomic 优化并发性能
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/kamalyes/go-logger"
)

// Stats 统计信息收集器
type Stats struct {
	totalEntries  int64        // 总条目数（atomic 计数器）
	totalBytes    int64        // 总字节数（atomic 计数器）
	totalBatches  int64        // 总批次数（atomic 计数器）
	sendErrors    int64        // 发送错误数（atomic 计数器）
	sendSuccess   int64        // 发送成功数（atomic 计数器）
	retryCount    int64        // 重试次数（atomic 计数器）
	droppedCount  int64        // 丢弃次数（atomic 计数器）
	startTime     time.Time    // 启动时间（不可变）
	mu            sync.RWMutex // 保护非原子字段的读写锁
	lastSendTime  time.Time    // 最后发送时间
	lastErrorTime time.Time    // 最后错误时间
	lastErrorMsg  string       // 最后错误消息
}

// NewStats 创建统计信息收集器
func NewStats() *Stats {
	return &Stats{
		startTime: time.Now(),
	}
}

// IncTotalEntries 增加总条目数
func (s *Stats) IncTotalEntries() {
	atomic.AddInt64(&s.totalEntries, 1)
}

// AddTotalBytes 增加总字节数
func (s *Stats) AddTotalBytes(n int64) {
	atomic.AddInt64(&s.totalBytes, n)
}

// IncTotalBatches 增加总批次数
func (s *Stats) IncTotalBatches() {
	atomic.AddInt64(&s.totalBatches, 1)
}

// IncSendErrors 增加发送错误数
func (s *Stats) IncSendErrors() {
	atomic.AddInt64(&s.sendErrors, 1)
}

// IncSendSuccess 增加发送成功数
func (s *Stats) IncSendSuccess() {
	atomic.AddInt64(&s.sendSuccess, 1)
}

// IncRetryCount 增加重试次数
func (s *Stats) IncRetryCount() {
	atomic.AddInt64(&s.retryCount, 1)
}

// IncDroppedCount 增加丢弃次数
func (s *Stats) IncDroppedCount() {
	atomic.AddInt64(&s.droppedCount, 1)
}

// SetLastSend 设置最后发送时间
func (s *Stats) SetLastSend(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSendTime = t
}

// SetLastError 设置最后错误信息
func (s *Stats) SetLastError(t time.Time, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastErrorTime = t
	s.lastErrorMsg = msg
}

// Snapshot 获取写入器统计信息快照（兼容 go-logger 接口）
func (s *Stats) Snapshot() logger.WriterStatsSnapshot {
	s.mu.RLock()
	lastSend := s.lastSendTime
	s.mu.RUnlock()

	return logger.WriterStatsSnapshot{
		BytesWritten: atomic.LoadInt64(&s.totalBytes),
		LinesWritten: atomic.LoadInt64(&s.totalEntries),
		ErrorCount:   atomic.LoadInt64(&s.sendErrors),
		LastWrite:    lastSend,
		StartTime:    s.startTime,
		Uptime:       time.Since(s.startTime),
	}
}

// StatsSnapshot 详细统计信息快照
type StatsSnapshot struct {
	TotalEntries  int64         `json:"total_entries"`   // 总条目数
	TotalBytes    int64         `json:"total_bytes"`     // 总字节数
	TotalBatches  int64         `json:"total_batches"`   // 总批次数
	SendErrors    int64         `json:"send_errors"`     // 发送错误数
	SendSuccess   int64         `json:"send_success"`    // 发送成功数
	RetryCount    int64         `json:"retry_count"`     // 重试次数
	DroppedCount  int64         `json:"dropped_count"`   // 丢弃次数
	LastSendTime  time.Time     `json:"last_send_time"`  // 最后发送时间
	LastErrorTime time.Time     `json:"last_error_time"` // 最后错误时间
	LastErrorMsg  string        `json:"last_error_msg"`  // 最后错误消息
	Uptime        time.Duration `json:"uptime"`          // 运行时长
}

// DetailedSnapshot 获取详细统计信息快照
func (s *Stats) DetailedSnapshot() StatsSnapshot {
	s.mu.RLock()
	lastSend := s.lastSendTime
	lastErr := s.lastErrorTime
	lastErrMsg := s.lastErrorMsg
	s.mu.RUnlock()

	return StatsSnapshot{
		TotalEntries:  atomic.LoadInt64(&s.totalEntries),
		TotalBytes:    atomic.LoadInt64(&s.totalBytes),
		TotalBatches:  atomic.LoadInt64(&s.totalBatches),
		SendErrors:    atomic.LoadInt64(&s.sendErrors),
		SendSuccess:   atomic.LoadInt64(&s.sendSuccess),
		RetryCount:    atomic.LoadInt64(&s.retryCount),
		DroppedCount:  atomic.LoadInt64(&s.droppedCount),
		LastSendTime:  lastSend,
		LastErrorTime: lastErr,
		LastErrorMsg:  lastErrMsg,
		Uptime:        time.Since(s.startTime),
	}
}

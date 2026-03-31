/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\mover.go
 * @Description: 数据搬运器，定时将过期批次和重试批次提交到工作池
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"sync"
	"sync/atomic"
	"time"
)

// Mover 数据搬运器，定时触发批次的过期检查和重试处理
type Mover struct {
	accumulator *Accumulator   // 日志累加器
	retryQueue  *RetryQueue    // 重试队列
	pool        *WorkerPool    // I/O 工作池
	stats       *Stats         // 统计信息
	lingerMs    int64          // 批次等待时间（毫秒）
	stopCh      chan struct{}  // 停止信号通道
	wg          sync.WaitGroup // 等待组，确保 goroutine 优雅退出
	running     atomic.Bool    // 运行状态标识
}

// NewMover 创建数据搬运器
func NewMover(accumulator *Accumulator, retryQueue *RetryQueue, pool *WorkerPool, stats *Stats, lingerMs int64) *Mover {
	return &Mover{
		accumulator: accumulator,
		retryQueue:  retryQueue,
		pool:        pool,
		stats:       stats,
		lingerMs:    lingerMs,
		stopCh:      make(chan struct{}),
	}
}

// Start 启动搬运器
func (m *Mover) Start() {
	if m.running.Load() {
		return
	}
	m.running.Store(true)
	m.wg.Add(1)
	go m.run()
}

// run 搬运器主循环
func (m *Mover) run() {
	defer m.wg.Done()
	ticker := time.NewTimer(time.Duration(m.lingerMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.tick()
			ticker.Reset(time.Duration(m.lingerMs) * time.Millisecond)
		}
	}
}

// tick 执行一次搬运操作
func (m *Mover) tick() {
	expiredBatches := m.accumulator.ExpireBatches(m.lingerMs)
	for _, batch := range expiredBatches {
		if batch.EntryCount() > 0 {
			m.pool.Submit(batch)
		}
	}

	retryBatches := m.retryQueue.PopReady()
	for _, batch := range retryBatches {
		if batch.EntryCount() > 0 {
			if m.stats != nil {
				m.stats.IncRetryCount()
			}
			m.pool.Submit(batch)
		}
	}
}

// Stop 停止搬运器
func (m *Mover) Stop() {
	if !m.running.Load() {
		return
	}
	m.running.Store(false)
	close(m.stopCh)
	m.wg.Wait()
}

// Flush 刷新所有待处理的批次
func (m *Mover) Flush() {
	expiredBatches := m.accumulator.FlushAll()
	for _, batch := range expiredBatches {
		if batch.EntryCount() > 0 {
			m.pool.Submit(batch)
		}
	}

	retryBatches := m.retryQueue.FlushAll()
	for _, batch := range retryBatches {
		if batch.EntryCount() > 0 {
			m.pool.Submit(batch)
		}
	}
}

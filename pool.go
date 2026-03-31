/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\pool.go
 * @Description: I/O 工作池，基于 go-toolbox 的 WorkerPool 实现并发发送
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"context"
	"sync/atomic"

	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/syncx"
)

// WorkerPool I/O 工作池，管理并发发送任务
type WorkerPool struct {
	pool         *syncx.WorkerPool  // 底层工作池
	sendFunc     func(*Batch) error // 发送函数
	shutdownFlag atomic.Bool        // 关闭标识
	stats        *Stats             // 统计信息
}

// NewWorkerPool 创建 I/O 工作池
func NewWorkerPool(sendFunc func(*Batch) error, maxWorkers int64, stats *Stats) *WorkerPool {
	wp := syncx.NewWorkerPool(int(maxWorkers), constants.DefaultWorkerPoolQueueSize)
	return &WorkerPool{
		pool:     wp,
		sendFunc: sendFunc,
		stats:    stats,
	}
}

// Submit 提交一个批次到工作池
func (wp *WorkerPool) Submit(batch *Batch) {
	if wp.shutdownFlag.Load() {
		return
	}
	_ = wp.pool.Submit(context.Background(), func() {
		if err := wp.sendFunc(batch); err != nil {
			if wp.stats != nil {
				wp.stats.IncSendErrors()
			}
		}
	})
}

// Shutdown 关闭工作池
func (wp *WorkerPool) Shutdown() {
	wp.shutdownFlag.Store(true)
	_ = wp.pool.Close()
}

// IsShutdown 检查工作池是否已关闭
func (wp *WorkerPool) IsShutdown() bool {
	return wp.shutdownFlag.Load()
}

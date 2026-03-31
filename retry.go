/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\retry.go
 * @Description: 重试队列实现，基于最小堆的优先级队列
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"container/heap"
	"sync"
	"time"
)

// RetryQueue 重试队列，基于最小堆实现，按重试时间排序
type RetryQueue struct {
	mu    sync.Mutex // 保护并发访问
	items []*Batch   // 批次列表（最小堆）
}

// NewRetryQueue 创建重试队列
func NewRetryQueue() *RetryQueue {
	rq := &RetryQueue{
		items: make([]*Batch, 0),
	}
	heap.Init(rq)
	return rq
}

// Enqueue 将批次加入重试队列
func (rq *RetryQueue) Enqueue(batch *Batch) {
	rq.mu.Lock()
	defer rq.mu.Unlock()
	heap.Push(rq, batch)
}

// PopReady 弹出所有已到重试时间的批次
func (rq *RetryQueue) PopReady() []*Batch {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	var ready []*Batch
	now := time.Now().UnixMilli()
	for rq.Len() > 0 {
		batch := heap.Pop(rq).(*Batch)
		if batch.GetNextRetryMs() <= now {
			ready = append(ready, batch)
		} else {
			heap.Push(rq, batch)
			break
		}
	}
	return ready
}

// FlushAll 弹出所有批次（用于关闭时刷新）
func (rq *RetryQueue) FlushAll() []*Batch {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	var all []*Batch
	for rq.Len() > 0 {
		all = append(all, heap.Pop(rq).(*Batch))
	}
	return all
}

// Len 实现 heap.Interface
func (rq *RetryQueue) Len() int {
	return len(rq.items)
}

// Less 实现 heap.Interface，按下次重试时间排序
func (rq *RetryQueue) Less(i, j int) bool {
	return rq.items[i].GetNextRetryMs() < rq.items[j].GetNextRetryMs()
}

// Swap 实现 heap.Interface
func (rq *RetryQueue) Swap(i, j int) {
	rq.items[i], rq.items[j] = rq.items[j], rq.items[i]
}

// Push 实现 heap.Interface
func (rq *RetryQueue) Push(x interface{}) {
	rq.items = append(rq.items, x.(*Batch))
}

// Pop 实现 heap.Interface
func (rq *RetryQueue) Pop() interface{} {
	old := rq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	rq.items = old[:n-1]
	return item
}

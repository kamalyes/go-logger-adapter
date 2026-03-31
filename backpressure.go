/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\backpressure.go
 * @Description: 背压控制器，防止内存溢出
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"sync/atomic"
	"time"
)

// BackPressure 背压控制器，通过内存限制和阻塞等待防止 OOM
type BackPressure struct {
	totalBytes int64 // 当前已使用的内存字节数（atomic 计数器）
	maxBytes   int64 // 最大允许的内存字节数
	maxBlockMs int64 // 最大阻塞等待时间（毫秒），0=不阻塞，-1=无限等待
}

// NewBackPressure 创建背压控制器
func NewBackPressure(maxBytes, maxBlockMs int64) *BackPressure {
	return &BackPressure{
		maxBytes:   maxBytes,
		maxBlockMs: maxBlockMs,
	}
}

// Wait 等待内存空间可用
// 如果加入 addBytes 后未超限，立即返回 nil
// 如果超限，根据 maxBlockMs 配置进行阻塞等待或返回错误
func (bp *BackPressure) Wait(addBytes int64) error {
	newTotal := atomic.AddInt64(&bp.totalBytes, addBytes)
	if newTotal <= bp.maxBytes {
		return nil
	}

	if bp.maxBlockMs == 0 {
		atomic.AddInt64(&bp.totalBytes, -addBytes)
		return errorx.NewTypedError(ErrTypeBackpressure, ErrMsgMemoryLimitExceeded)
	}

	if bp.maxBlockMs < 0 {
		for atomic.LoadInt64(&bp.totalBytes) > bp.maxBytes {
			time.Sleep(constants.DefaultBackPressureSleep)
		}
		return nil
	}

	deadline := time.Now().Add(time.Duration(bp.maxBlockMs) * time.Millisecond)
	for atomic.LoadInt64(&bp.totalBytes) > bp.maxBytes {
		if time.Now().After(deadline) {
			atomic.AddInt64(&bp.totalBytes, -addBytes)
			return errorx.NewTypedError(ErrTypeBackpressure, ErrMsgMemoryLimitExceeded)
		}
		time.Sleep(constants.DefaultBackPressureSleep)
	}
	return nil
}

// Release 释放已使用的内存字节数
func (bp *BackPressure) Release(bytes int64) {
	atomic.AddInt64(&bp.totalBytes, -bytes)
}

// CurrentBytes 获取当前已使用的内存字节数
func (bp *BackPressure) CurrentBytes() int64 {
	return atomic.LoadInt64(&bp.totalBytes)
}

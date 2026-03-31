/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\plugin.go
 * @Description: 日志插件接口定义，各后端（es/loki/openobserve/victorialogs）需实现此接口
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"context"
)

// Plugin 日志插件接口，定义日志后端的格式化、发送和健康检查能力
type Plugin interface {
	// Name 返回插件名称（如 "elasticsearch"、"loki" 等）
	Name() string
	// Format 将日志条目格式化为目标后端所需的字节流
	Format(entries []LogEntry) ([]byte, error)
	// Send 将格式化后的数据发送到目标后端
	Send(ctx context.Context, body []byte, headers map[string]string) error
	// HealthCheck 检查目标后端的健康状态
	HealthCheck(ctx context.Context) error
}

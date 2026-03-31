/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\es\plugin.go
 * @Description: Elasticsearch 日志适配器 Writer 和 Plugin 实现
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package es

import (
	"context"
	"fmt"

	adapter "github.com/kamalyes/go-logger-adapter"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"github.com/kamalyes/go-toolbox/pkg/httpx"

	logger "github.com/kamalyes/go-logger"
)

// Writer Elasticsearch 日志写入器，嵌入 BaseWriter 复用公共方法
type Writer struct {
	*adapter.BaseWriter // 嵌入基础写入器，复用 Start/Flush/Close/Write/WriteLevel 等方法
	config              *Config
}

// NewWriter 创建 Elasticsearch 日志写入器
func NewWriter(config *Config) (*Writer, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	config.SetCommonDefaults()

	plugin := &ESPlugin{config: config}

	opts := adapter.CommonAdapterOpts(config.Common)
	engine, err := adapter.NewEngine(plugin, opts...)
	if err != nil {
		return nil, err
	}

	return &Writer{
		BaseWriter: adapter.NewBaseWriter(engine),
		config:     config,
	}, nil
}

// GetStats 获取写入器统计信息快照
func (w *Writer) GetStats() logger.WriterStatsSnapshot {
	return w.BaseWriter.GetStats()
}

// IsHealthy 检查写入器健康状态
func (w *Writer) IsHealthy() bool {
	return w.BaseWriter.IsHealthy()
}

// ESPlugin Elasticsearch 插件实现
type ESPlugin struct {
	config *Config
	client *httpx.Client
}

// Name 返回插件名称
func (p *ESPlugin) Name() string { return constants.PluginNameElasticsearch }

// Format 将日志条目格式化为 Elasticsearch Bulk API 格式
func (p *ESPlugin) Format(entries []adapter.LogEntry) ([]byte, error) {
	return FormatBulkBody(entries, p.config.IndexFormat)
}

// Send 将格式化后的数据发送到 Elasticsearch
func (p *ESPlugin) Send(ctx context.Context, body []byte, headers map[string]string) error {
	req := p.client.Post(p.config.BuildBulkURL())
	req = req.SetBodyRaw(body).WithContext(ctx)
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	for k, v := range p.config.AuthHeaders() {
		req.SetHeader(k, v)
	}

	resp, err := req.Send()
	if err != nil {
		return errorx.WrapError(fmt.Sprintf("%s: request error", constants.PluginNameElasticsearch), err)
	}
	defer resp.Close()

	respBody, _ := resp.Bytes()
	if resp.StatusCode >= constants.HTTPStatusServerError {
		return adapter.NewHTTPError(resp.StatusCode, string(respBody))
	}
	return nil
}

// HealthCheck 检查 Elasticsearch 健康状态
func (p *ESPlugin) HealthCheck(ctx context.Context) error {
	req := p.client.Get(p.config.Endpoints[0]).WithContext(ctx)
	for k, v := range p.config.AuthHeaders() {
		req.SetHeader(k, v)
	}

	resp, err := req.Send()
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode >= constants.HTTPStatusServerError {
		return adapter.NewUnhealthyError(constants.PluginNameElasticsearch, resp.StatusCode)
	}
	return nil
}

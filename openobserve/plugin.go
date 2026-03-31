/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\openobserve\plugin.go
 * @Description: OpenObserve 日志适配器 Writer 和 Plugin 实现
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package openobserve

import (
	"context"
	"fmt"

	adapter "github.com/kamalyes/go-logger-adapter"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"github.com/kamalyes/go-toolbox/pkg/httpx"

	logger "github.com/kamalyes/go-logger"
)

// Writer OpenObserve 日志写入器，嵌入 BaseWriter 复用公共方法
type Writer struct {
	*adapter.BaseWriter // 嵌入基础写入器，复用 Start/Flush/Close/Write/WriteLevel 等方法
	config              *Config
}

// NewWriter 创建 OpenObserve 日志写入器
func NewWriter(config *Config) (*Writer, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	config.SetCommonDefaults()

	plugin := &OpenObservePlugin{config: config}

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

// OpenObservePlugin OpenObserve 插件实现
type OpenObservePlugin struct {
	config *Config
	client *httpx.Client
}

// Name 返回插件名称
func (p *OpenObservePlugin) Name() string { return constants.PluginNameOpenObserve }

// Format 将日志条目格式化为 OpenObserve Ingest API 请求体
func (p *OpenObservePlugin) Format(entries []adapter.LogEntry) ([]byte, error) {
	return FormatIngestBody(entries)
}

// Send 将格式化后的数据发送到 OpenObserve
func (p *OpenObservePlugin) Send(ctx context.Context, body []byte, headers map[string]string) error {
	req := p.client.Post(p.config.BuildIngestURL()).WithContext(ctx)
	req.SetBodyRaw(body).SetContentType(httpx.ContentTypeApplicationJSON)
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	for k, v := range p.config.AuthHeaders() {
		req.SetHeader(k, v)
	}

	resp, err := req.Send()
	if err != nil {
		return errorx.WrapError(fmt.Sprintf("%s: request error", constants.PluginNameOpenObserve), err)
	}
	defer resp.Close()

	respBody, _ := resp.Bytes()
	if resp.StatusCode >= constants.HTTPStatusServerError {
		return adapter.NewHTTPError(resp.StatusCode, string(respBody))
	}
	return nil
}

// HealthCheck 检查 OpenObserve 健康状态
func (p *OpenObservePlugin) HealthCheck(ctx context.Context) error {
	req := p.client.Get(p.config.Endpoint + constants.OpenObserveHealthPath).WithContext(ctx)
	for k, v := range p.config.AuthHeaders() {
		req.SetHeader(k, v)
	}

	resp, err := req.Send()
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode >= constants.HTTPStatusServerError {
		return adapter.NewUnhealthyError(constants.PluginNameOpenObserve, resp.StatusCode)
	}
	return nil
}

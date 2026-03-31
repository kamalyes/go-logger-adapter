/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\loki\plugin.go
 * @Description: Loki 日志适配器 Writer 和 Plugin 实现
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package loki

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	adapter "github.com/kamalyes/go-logger-adapter"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"github.com/kamalyes/go-toolbox/pkg/httpx"
	"github.com/kamalyes/go-toolbox/pkg/mathx"

	logger "github.com/kamalyes/go-logger"
)

// Writer Loki 日志写入器，嵌入 BaseWriter 复用公共方法
type Writer struct {
	*adapter.BaseWriter // 嵌入基础写入器，复用 Start/Flush/Close/Write/WriteLevel 等方法
	config              *Config
}

// NewWriter 创建 Loki 日志写入器
func NewWriter(config *Config) (*Writer, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	config.SetCommonDefaults()

	plugin := &LokiPlugin{config: config}

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

// CheckHealth 执行健康检查并返回详细状态
func (w *Writer) CheckHealth() (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultHealthTimeoutX*adapter.DefaultConfig().RequestTimeout)
	defer cancel()

	endpoint := strings.TrimRight(w.config.Endpoint, "/")
	req := w.Engine().HTTPClient().Get(endpoint + constants.LokiReadyPath).WithContext(ctx)
	for k, v := range w.config.AuthHeaders() {
		req.SetHeader(k, v)
	}

	resp, err := req.Send()
	if err != nil {
		return false, fmt.Sprintf("%s: health check request error: %v", constants.PluginNameLoki, err)
	}
	defer resp.Close()

	if resp.StatusCode == constants.HTTPStatusNoContent || resp.StatusCode == constants.HTTPStatusOK {
		return true, "healthy"
	}

	body, _ := resp.Bytes()
	var healthResp struct {
		Status string `json:"status"`
	}
	if json.Unmarshal(body, &healthResp) == nil && healthResp.Status == "ready" {
		return true, "healthy"
	}

	return false, fmt.Sprintf("%s: unhealthy, status code: %d", constants.PluginNameLoki, resp.StatusCode)
}

// LokiPlugin Loki 插件实现
type LokiPlugin struct {
	config *Config
	client *httpx.Client
}

// Name 返回插件名称
func (p *LokiPlugin) Name() string { return constants.PluginNameLoki }

// Format 将日志条目格式化为 Loki Push API 请求体
func (p *LokiPlugin) Format(entries []adapter.LogEntry) ([]byte, error) {
	return FormatPushRequest(entries, p.config.Labels)
}

// Send 将格式化后的数据发送到 Loki
func (p *LokiPlugin) Send(ctx context.Context, body []byte, headers map[string]string) error {
	endpoint := strings.TrimRight(p.config.Endpoint, "/")
	pushURL := endpoint + constants.LokiPushPath

	contentType := httpx.ContentTypeApplicationJSON
	if enc, ok := headers[constants.HeaderContentEncoding]; ok && enc == "snappy" {
		contentType = httpx.ContentTypeApplicationXSnappyFramed
	}

	req := p.client.Post(pushURL).WithContext(ctx)
	req.SetBodyRaw(body)
	req.SetContentType(contentType)
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	for k, v := range p.config.AuthHeaders() {
		req.SetHeader(k, v)
	}

	resp, err := req.Send()
	if err != nil {
		return errorx.WrapError(fmt.Sprintf("%s: request error", constants.PluginNameLoki), err)
	}
	defer resp.Close()

	respBody, _ := resp.Bytes()
	if resp.StatusCode >= constants.HTTPStatusServerError {
		return adapter.NewHTTPError(resp.StatusCode, string(respBody))
	}
	return nil
}

// HealthCheck 检查 Loki 健康状态
func (p *LokiPlugin) HealthCheck(ctx context.Context) error {
	endpoint := strings.TrimRight(p.config.Endpoint, "/")
	req := p.client.Get(endpoint + constants.LokiReadyPath).WithContext(ctx)
	for k, v := range p.config.AuthHeaders() {
		req.SetHeader(k, v)
	}

	resp, err := req.Send()
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode != constants.HTTPStatusOK && resp.StatusCode != constants.HTTPStatusNoContent {
		return adapter.NewUnhealthyError(constants.PluginNameLoki, resp.StatusCode)
	}
	return nil
}

// IsRetryable 判断 HTTP 状态码是否可重试
func (p *LokiPlugin) IsRetryable(statusCode int) bool {
	return mathx.IF(statusCode >= constants.HTTPStatus5xxStart || statusCode == constants.HTTPStatusTooManyReq, true, false)
}

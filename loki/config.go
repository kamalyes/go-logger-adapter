/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\loki\config.go
 * @Description: Loki 适配器配置和数据格式化
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package loki

import (
	"encoding/json"
	"fmt"
	adapter "github.com/kamalyes/go-logger-adapter"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"sort"
	"strings"
)

// Config Loki 适配器配置
type Config struct {
	Endpoint string             `json:"endpoint" yaml:"endpoint"`   // Loki 服务地址
	Labels   map[string]string  `json:"labels" yaml:"labels"`       // 静态标签
	TenantID string             `json:"tenant_id" yaml:"tenant_id"` // 多租户 ID
	Auth     adapter.AuthConfig `json:"auth" yaml:"auth"`           // 认证配置
	UseJSON  bool               `json:"use_json" yaml:"use_json"`   // 使用 JSON 格式
	Common   *adapter.Config    `json:"common" yaml:"common"`       // 通用引擎配置
}

// DefaultConfig 返回 Loki 默认配置
func DefaultConfig() *Config {
	return &Config{
		Endpoint: constants.DefaultLokiEndpoint,
		Labels:   map[string]string{"job": constants.DefaultLokiJobLabel},
		UseJSON:  true,
		Common:   adapter.DefaultConfig(),
	}
}

// Validate 校验配置合法性
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errorx.NewTypedError(adapter.ErrTypeEndpointRequired, adapter.ErrFmtEndpointRequired, constants.PluginNameLoki)
	}
	if err := c.Auth.ValidateBasic(constants.PluginNameLoki); err != nil {
		return err
	}
	if err := c.Auth.ValidateBearer(constants.PluginNameLoki); err != nil {
		return err
	}
	return nil
}

// AuthHeaders 生成认证头（包含 TenantID）
func (c *Config) AuthHeaders() map[string]string {
	headers := c.Auth.AuthHeaders()
	if c.TenantID != "" {
		headers[constants.HeaderXScopeOrgID] = c.TenantID
	}
	return headers
}

// SetCommonDefaults 设置通用配置默认值（Loki 默认启用 Gzip 压缩）
func (c *Config) SetCommonDefaults() {
	adapter.EnsureCommonDefaults(&c.Common)
	if c.Common.Compression == constants.CompressionNone {
		c.Common.Compression = constants.CompressionGzip
	}
}

// LokiPushRequest Loki Push API 请求体
type LokiPushRequest struct {
	Streams []LokiStream `json:"streams"`
}

// LokiStream Loki 日志流
type LokiStream struct {
	Stream map[string]string `json:"stream"` // 流标签
	Values []LokiEntry       `json:"values"` // 日志条目
}

// LokiEntry Loki 日志条目
type LokiEntry struct {
	Timestamp string            // 纳秒时间戳字符串
	Line      string            // 日志行内容
	Labels    map[string]string // 结构化标签
}

// MarshalJSON 实现 JSON 序列化，支持结构化标签
func (e LokiEntry) MarshalJSON() ([]byte, error) {
	if len(e.Labels) == 0 {
		return json.Marshal([]string{e.Timestamp, e.Line})
	}
	return json.Marshal([]interface{}{e.Timestamp, e.Line, e.Labels})
}

// FormatPushRequest 将日志条目格式化为 Loki Push API 请求体
func FormatPushRequest(entries []adapter.LogEntry, labels map[string]string) ([]byte, error) {
	streams := make(map[string]*LokiStream)

	for i := range entries {
		entry := &entries[i]
		streamLabels := buildStreamLabels(entry, labels)
		labelKey := sortedLabelString(streamLabels)

		stream, exists := streams[labelKey]
		if !exists {
			stream = &LokiStream{
				Stream: streamLabels,
				Values: make([]LokiEntry, 0),
			}
			streams[labelKey] = stream
		}

		lokiEntry := LokiEntry{
			Timestamp: formatLokiTimestamp(entry.Timestamp),
			Line:      entry.Message,
			Labels:    extractStructuredLabels(entry),
		}
		stream.Values = append(stream.Values, lokiEntry)
	}

	req := &LokiPushRequest{
		Streams: make([]LokiStream, 0, len(streams)),
	}
	for _, stream := range streams {
		req.Streams = append(req.Streams, *stream)
	}

	return json.Marshal(req)
}

// buildStreamLabels 构建流标签（合并静态标签和日志级别）
func buildStreamLabels(entry *adapter.LogEntry, extraLabels map[string]string) map[string]string {
	labels := make(map[string]string)
	for k, v := range extraLabels {
		labels[k] = v
	}
	labels["level"] = entry.Level.String()
	return labels
}

// extractStructuredLabels 从日志条目的 Fields 中提取结构化标签
func extractStructuredLabels(entry *adapter.LogEntry) map[string]string {
	if len(entry.Fields) == 0 {
		return nil
	}
	labels := make(map[string]string)
	for k, v := range entry.Fields {
		labels[k] = fmt.Sprintf("%v", v)
	}
	return labels
}

// sortedLabelString 将标签排序后拼接为字符串（用作流分组键）
func sortedLabelString(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(labels[k])
	}
	return sb.String()
}

// formatLokiTimestamp 将毫秒时间戳转换为 Loki 所需的纳秒时间戳字符串
func formatLokiTimestamp(ts int64) string {
	return fmt.Sprintf("%d", ts*constants.LokiTimestampMultiplier)
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-31 13:27:56
 * @FilePath: \go-logger-adapter\openobserve\config.go
 * @Description: OpenObserve 适配器配置和数据格式化
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package openobserve

import (
	"encoding/json"
	"fmt"
	adapter "github.com/kamalyes/go-logger-adapter"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"strings"
)

// Config OpenObserve 适配器配置
type Config struct {
	Endpoint   string             `json:"endpoint" yaml:"endpoint"`       // OpenObserve 服务地址
	StreamName string             `json:"stream_name" yaml:"stream_name"` // 日志流名称
	OrgID      string             `json:"org_id" yaml:"org_id"`           // 组织 ID
	Auth       adapter.AuthConfig `json:"auth" yaml:"auth"`               // 认证配置
	Common     *adapter.Config    `json:"common" yaml:"common"`           // 通用引擎配置
}

// DefaultConfig 返回 OpenObserve 默认配置
func DefaultConfig() *Config {
	return &Config{
		Endpoint:   constants.DefaultOpenObserveEndpoint,
		StreamName: constants.DefaultOpenObserveStream,
		OrgID:      constants.DefaultOpenObserveOrgID,
		Common:     adapter.DefaultConfig(),
	}
}

// Validate 校验配置合法性
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errorx.NewTypedError(adapter.ErrTypeEndpointRequired, adapter.ErrFmtEndpointRequired, constants.PluginNameOpenObserve)
	}
	if c.StreamName == "" {
		return errorx.NewTypedError(adapter.ErrTypeStreamNameRequired, adapter.ErrFmtStreamRequired, constants.PluginNameOpenObserve)
	}
	if c.OrgID == "" {
		return errorx.NewTypedError(adapter.ErrTypeOrgIDRequired, adapter.ErrFmtOrgIDRequired, constants.PluginNameOpenObserve)
	}
	if err := c.Auth.ValidateBasic(constants.PluginNameOpenObserve); err != nil {
		return err
	}
	if err := c.Auth.ValidateAPIKey(constants.PluginNameOpenObserve); err != nil {
		return err
	}
	if err := c.Auth.ValidateBearer(constants.PluginNameOpenObserve); err != nil {
		return err
	}
	return nil
}

// AuthHeaders 生成认证头（包含 Organization ID）
func (c *Config) AuthHeaders() map[string]string {
	headers := c.Auth.AuthHeaders()
	if c.OrgID != "" {
		headers[constants.HeaderZoOrgID] = c.OrgID
	}
	return headers
}

// SetCommonDefaults 设置通用配置默认值
func (c *Config) SetCommonDefaults() {
	adapter.EnsureCommonDefaults(&c.Common)
}

// BuildIngestURL 构建 Ingest API URL
func (c *Config) BuildIngestURL() string {
	endpoint := strings.TrimRight(c.Endpoint, "/")
	return fmt.Sprintf("%s"+constants.OpenObserveIngestFormat, endpoint, c.OrgID, c.StreamName)
}

// FormatIngestBody 将日志条目格式化为 OpenObserve Ingest API 请求体
func FormatIngestBody(entries []adapter.LogEntry) ([]byte, error) {
	docs := make([]map[string]interface{}, 0, len(entries))
	for i := range entries {
		doc := adapter.FormatEntryToMap(&entries[i], "_timestamp")
		doc["level"] = entries[i].Level.String()
		docs = append(docs, doc)
	}
	return json.Marshal(docs)
}

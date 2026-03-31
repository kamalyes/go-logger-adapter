/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-31 13:28:05
 * @FilePath: \go-logger-adapter\victorialogs\config.go
 * @Description: VictoriaLogs 适配器配置和数据格式化
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package victorialogs

import (
	"encoding/json"
	adapter "github.com/kamalyes/go-logger-adapter"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"strings"
)

// Config VictoriaLogs 适配器配置
type Config struct {
	Endpoint string             `json:"endpoint" yaml:"endpoint"`   // VictoriaLogs 服务地址
	TenantID string             `json:"tenant_id" yaml:"tenant_id"` // 租户 ID（AccountID:ProjectID 格式）
	Auth     adapter.AuthConfig `json:"auth" yaml:"auth"`           // 认证配置
	Common   *adapter.Config    `json:"common" yaml:"common"`       // 通用引擎配置
}

// DefaultConfig 返回 VictoriaLogs 默认配置
func DefaultConfig() *Config {
	return &Config{
		Endpoint: constants.DefaultVictoriaLogsEndpoint,
		Common:   adapter.DefaultConfig(),
	}
}

// Validate 校验配置合法性
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errorx.NewTypedError(adapter.ErrTypeEndpointRequired, adapter.ErrFmtEndpointRequired, constants.PluginNameVictoriaLogs)
	}
	if err := c.Auth.ValidateBasic(constants.PluginNameVictoriaLogs); err != nil {
		return err
	}
	if err := c.Auth.ValidateBearer(constants.PluginNameVictoriaLogs); err != nil {
		return err
	}
	return nil
}

// AuthHeaders 生成认证头（包含 TenantID）
func (c *Config) AuthHeaders() map[string]string {
	headers := c.Auth.AuthHeaders()
	if c.TenantID != "" {
		headers[constants.HeaderAccountID] = c.TenantID
	}
	return headers
}

// SetCommonDefaults 设置通用配置默认值
func (c *Config) SetCommonDefaults() {
	adapter.EnsureCommonDefaults(&c.Common)
}

// BuildInsertURL 构建 Insert API URL
func (c *Config) BuildInsertURL() string {
	endpoint := strings.TrimRight(c.Endpoint, "/")
	return endpoint + constants.VictoriaLogsInsertPath
}

// FormatInsertBody 将日志条目格式化为 VictoriaLogs JSON Line 格式
func FormatInsertBody(entries []adapter.LogEntry) ([]byte, error) {
	var buf strings.Builder
	for i := range entries {
		doc := adapter.FormatEntryToMap(&entries[i], "_ts")
		doc["_level"] = entries[i].Level.String()
		data, err := json.Marshal(doc)
		if err != nil {
			return nil, errorx.NewTypedError(adapter.ErrTypeMarshalFailed, adapter.ErrFmtMarshalEntryFail, constants.PluginNameVictoriaLogs, err)
		}
		buf.Write(data)
		if i < len(entries)-1 {
			buf.WriteByte('\n')
		}
	}
	return []byte(buf.String()), nil
}

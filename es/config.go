/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\es\config.go
 * @Description: Elasticsearch 适配器配置
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package es

import (
	adapter "github.com/kamalyes/go-logger-adapter"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"net/url"
	"strings"
)

// TLSConfig TLS 配置
type TLSConfig struct {
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	CACertFile         string `json:"ca_cert_file" yaml:"ca_cert_file"`
	CertFile           string `json:"cert_file" yaml:"cert_file"`
	KeyFile            string `json:"key_file" yaml:"key_file"`
	ServerName         string `json:"server_name" yaml:"server_name"`
}

// Config Elasticsearch 适配器配置
type Config struct {
	Endpoints   []string           `json:"endpoints" yaml:"endpoints"`       // ES 节点地址列表
	IndexFormat string             `json:"index_format" yaml:"index_format"` // 索引名称格式（支持时间格式化）
	Auth        adapter.AuthConfig `json:"auth" yaml:"auth"`                 // 认证配置
	TLS         *TLSConfig         `json:"tls" yaml:"tls"`                   // TLS 配置
	Pipeline    string             `json:"pipeline" yaml:"pipeline"`         // Ingest Pipeline 名称
	Common      *adapter.Config    `json:"common" yaml:"common"`             // 通用引擎配置
}

// DefaultConfig 返回 Elasticsearch 默认配置
func DefaultConfig() *Config {
	return &Config{
		Endpoints:   []string{constants.DefaultElasticsearchEndpoint},
		IndexFormat: constants.DefaultESIndexFormat,
		Common:      adapter.DefaultConfig(),
	}
}

// Validate 校验配置合法性
func (c *Config) Validate() error {
	if len(c.Endpoints) == 0 {
		return errorx.NewTypedError(adapter.ErrTypeEndpointsRequired, adapter.ErrFmtEndpointsRequired, constants.PluginNameElasticsearch)
	}
	for _, endpoint := range c.Endpoints {
		if _, err := url.Parse(endpoint); err != nil {
			return errorx.NewTypedError(adapter.ErrTypeEndpointInvalid, adapter.ErrFmtEndpointInvalid, constants.PluginNameElasticsearch, endpoint, err)
		}
	}
	if err := c.Auth.ValidateBasic(constants.PluginNameElasticsearch); err != nil {
		return err
	}
	if err := c.Auth.ValidateAPIKey(constants.PluginNameElasticsearch); err != nil {
		return err
	}
	if err := c.Auth.ValidateBearer(constants.PluginNameElasticsearch); err != nil {
		return err
	}
	return nil
}

// BuildBulkURL 构建 Bulk API URL
func (c *Config) BuildBulkURL() string {
	endpoint := strings.TrimRight(c.Endpoints[0], "/")
	u := endpoint + constants.ElasticsearchBulkPath
	if c.Pipeline != "" {
		u += "?pipeline=" + c.Pipeline
	}
	return u
}

// AuthHeaders 生成认证头
func (c *Config) AuthHeaders() map[string]string {
	return c.Auth.AuthHeaders()
}

// SetCommonDefaults 设置通用配置默认值
func (c *Config) SetCommonDefaults() {
	adapter.EnsureCommonDefaults(&c.Common)
}

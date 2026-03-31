/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\auth.go
 * @Description: 统一认证类型和认证头生成
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"encoding/base64"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
)

// AuthType 认证类型枚举
type AuthType = int

// AuthConfig 认证配置，各插件可复用
type AuthConfig struct {
	AuthType    AuthType `json:"auth_type" yaml:"auth_type"`
	Username    string   `json:"username" yaml:"username"`
	Password    string   `json:"password" yaml:"password"`
	APIKey      string   `json:"api_key" yaml:"api_key"`
	BearerToken string   `json:"bearer_token" yaml:"bearer_token"`
	AuthToken   string   `json:"auth_token" yaml:"auth_token"`
}

// AuthHeaders 根据认证类型生成 HTTP 认证头
func (ac *AuthConfig) AuthHeaders() map[string]string {
	headers := make(map[string]string)
	switch ac.AuthType {
	case constants.AuthBasic:
		headers[constants.HeaderAuthorization] = constants.AuthPrefixBasic + EncodeBasicAuth(ac.Username, ac.Password)
	case constants.AuthAPIKey:
		headers[constants.HeaderAuthorization] = constants.AuthPrefixAPIKey + ac.APIKey
	case constants.AuthBearer:
		headers[constants.HeaderAuthorization] = constants.AuthPrefixBearer + ac.BearerToken
	case constants.AuthToken:
		headers[constants.HeaderAuthorization] = constants.AuthPrefixBearer + ac.AuthToken
	}
	return headers
}

// ValidateBasic 校验 Basic 认证配置
func (ac *AuthConfig) ValidateBasic(pluginName string) error {
	if ac.AuthType == constants.AuthBasic && (ac.Username == "" || ac.Password == "") {
		return errorx.NewTypedError(ErrTypeBasicAuthRequired, ErrFmtBasicAuthRequired, pluginName)
	}
	return nil
}

// ValidateAPIKey 校验 API Key 认证配置
func (ac *AuthConfig) ValidateAPIKey(pluginName string) error {
	if ac.AuthType == constants.AuthAPIKey && ac.APIKey == "" {
		return errorx.NewTypedError(ErrTypeAPIKeyRequired, ErrFmtAPIKeyRequired, pluginName)
	}
	return nil
}

// ValidateBearer 校验 Bearer Token 认证配置
func (ac *AuthConfig) ValidateBearer(pluginName string) error {
	if ac.AuthType == constants.AuthBearer && ac.BearerToken == "" {
		return errorx.NewTypedError(ErrTypeBearerRequired, ErrFmtBearerRequired, pluginName)
	}
	return nil
}

// ValidateToken 校验自定义 Token 认证配置
func (ac *AuthConfig) ValidateToken(pluginName string) error {
	if ac.AuthType == constants.AuthToken && ac.AuthToken == "" {
		return errorx.NewTypedError(ErrTypeTokenRequired, ErrFmtTokenRequired, pluginName)
	}
	return nil
}

// EncodeBasicAuth 将用户名和密码编码为 Base64 格式
func EncodeBasicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-22 11:23:38
 * @FilePath: \go-logger-adapter\errors.go
 * @Description: 适配器错误类型定义，直接使用 errorx 实现类型化错误管理
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"github.com/kamalyes/go-toolbox/pkg/errorx"
)

// ==================== 适配器错误类型 ====================
const (
	ErrTypeEndpointRequired   errorx.ErrorType = iota + 10000 // endpoint 缺失
	ErrTypeOrgRequired                                        // org_name 缺失
	ErrTypeStreamRequired                                     // stream_name 缺失
	ErrTypeBasicAuthRequired                                  // Basic 认证信息缺失
	ErrTypeAPIKeyRequired                                     // API Key 认证信息缺失
	ErrTypeBearerRequired                                     // Bearer Token 认证信息缺失
	ErrTypeTokenRequired                                      // 自定义 Token 认证信息缺失
	ErrTypePluginRequired                                     // 插件为空
	ErrTypeFormatFailed                                       // 格式化失败
	ErrTypeCompressFailed                                     // 压缩失败
	ErrTypeSendFailed                                         // 发送失败
	ErrTypeMaxRetriesExceeded                                 // 超过最大重试次数
	ErrTypeBackpressure                                       // 背压超限
	ErrTypeHTTPErr                                            // HTTP 请求错误
	ErrTypeUnhealthy                                          // 健康检查失败
	ErrTypeMarshalFailed                                      // 序列化失败
	ErrTypeEndpointInvalid                                    // endpoint 格式无效
	ErrTypeEndpointsRequired                                  // endpoints 列表为空
	ErrTypeStreamNameRequired                                 // stream_name 缺失
	ErrTypeOrgIDRequired                                      // org_id 缺失
)

// ==================== 适配器错误文案 ====================
const (
	ErrMsgPluginRequired      = "adapter: plugin is required"
	ErrMsgMaxRetriesExceeded  = "max retries exceeded"
	ErrMsgMemoryLimitExceeded = "adapter: memory limit exceeded"

	ErrFmtEndpointRequired  = "%s: endpoint is required"
	ErrFmtEndpointsRequired = "%s: at least one endpoint is required"
	ErrFmtEndpointInvalid   = "%s: invalid endpoint %s: %v"
	ErrFmtStreamRequired    = "%s: stream_name is required"
	ErrFmtOrgIDRequired     = "%s: org_id is required"

	ErrFmtBasicAuthRequired = "%s: basic auth requires username and password"
	ErrFmtAPIKeyRequired    = "%s: api key auth requires api_key"
	ErrFmtBearerRequired    = "%s: bearer auth requires bearer_token"
	ErrFmtTokenRequired     = "%s: token auth requires auth_token"

	ErrFmtFormatFailed      = "%s: format error: %v"
	ErrFmtCompressFailed    = "%s: compress error: %v"
	ErrFmtMarshalActionFail = "%s: failed to marshal action: %v"
	ErrFmtMarshalEntryFail  = "%s: failed to marshal entry: %v"
	ErrFmtSendFailed        = "%s: send failed with status %d, body: %s"
	ErrFmtUnhealthy         = "%s: unhealthy, status %d"
)

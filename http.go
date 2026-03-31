/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-12-06 00:00:00
 * @FilePath: \go-logger-adapter\http.go
 * @Description: HTTP 请求工具和错误类型定义
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"fmt"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"github.com/kamalyes/go-toolbox/pkg/httpx"
)

// HTTPError HTTP 请求错误，包含状态码和响应体
type HTTPError struct {
	StatusCode int    // HTTP 状态码
	Body       string // 响应体内容
}

// Error 实现 error 接口
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

// NewHTTPError 创建 HTTP 错误实例
func NewHTTPError(statusCode int, body string) *HTTPError {
	return &HTTPError{StatusCode: statusCode, Body: body}
}

// IsRetryable 判断 HTTP 状态码是否可重试
// 排除 noRetryCodes 中的状态码，5xx 和 429 视为可重试
func IsRetryable(statusCode int, noRetryCodes []int) bool {
	for _, code := range noRetryCodes {
		if statusCode == code {
			return false
		}
	}
	return statusCode >= constants.HTTPStatus5xxStart || statusCode == constants.HTTPStatusTooManyReq
}

// DoRequest 执行 HTTP 请求的通用方法
func DoRequest(client *httpx.Client, method, url string, headers map[string]string, body []byte) (httpx.Response, error) {
	req := client.NewRequest(method, url)
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	if body != nil {
		req.SetBodyRaw(body)
	}
	return req.Send()
}

// NewSendFailedError 创建发送失败错误
func NewSendFailedError(plugin string, statusCode int, body string) error {
	return errorx.NewTypedError(ErrTypeSendFailed, ErrFmtSendFailed, plugin, statusCode, body)
}

// NewUnhealthyError 创建健康检查失败错误
func NewUnhealthyError(plugin string, statusCode int) error {
	return errorx.NewTypedError(ErrTypeUnhealthy, ErrFmtUnhealthy, plugin, statusCode)
}

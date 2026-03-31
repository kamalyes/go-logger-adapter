/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\compressor.go
 * @Description: 数据压缩器实现，支持 Gzip 和 Zlib 压缩
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/zipx"
)

// CompressionType 压缩类型枚举
type CompressionType = int

// Compressor 压缩器接口
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	ContentEncoding() string
}

// GzipCompressor Gzip 压缩器实现
type GzipCompressor struct{}

// Compress 使用 Gzip 压缩数据
func (c *GzipCompressor) Compress(data []byte) ([]byte, error) {
	return zipx.GzipCompress(data)
}

// ContentEncoding 返回 gzip 编码头
func (c *GzipCompressor) ContentEncoding() string {
	return "gzip"
}

// DecompressGzip 解压 Gzip 数据
func DecompressGzip(data []byte) ([]byte, error) {
	return zipx.GzipDecompress(data)
}

// DecompressZlib 解压 Zlib 数据
func DecompressZlib(data []byte) ([]byte, error) {
	return zipx.ZlibDecompress(data)
}

// NewCompressor 根据压缩类型创建压缩器实例
// 如果类型为 CompressionNone，返回 nil
func NewCompressor(ct CompressionType) Compressor {
	switch ct {
	case constants.CompressionGzip:
		return &GzipCompressor{}
	default:
		return nil
	}
}

/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\constants\path.go
 * @Description: 插件名称和 API 路径常量
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package constants

// ==================== 插件名称常量 ====================

const (
	PluginNameElasticsearch = "elasticsearch" // Elasticsearch 插件名称
	PluginNameLoki          = "loki"          // Loki 插件名称
	PluginNameOpenObserve   = "openobserve"   // OpenObserve 插件名称
	PluginNameVictoriaLogs  = "victorialogs"  // VictoriaLogs 插件名称
)

// ==================== API 路径常量 ====================

const (
	ElasticsearchBulkPath   = "/_bulk"            // Elasticsearch Bulk API 路径
	LokiPushPath            = "/loki/api/v1/push" // Loki Push API 路径
	LokiReadyPath           = "/ready"            // Loki 健康检查路径
	OpenObserveIngestFormat = "/api/%s/%s/_json"  // OpenObserve Ingest API 路径格式
	OpenObserveHealthPath   = "/health"           // OpenObserve 健康检查路径
	VictoriaLogsInsertPath  = "/insert/jsonline"  // VictoriaLogs Insert API 路径
	VictoriaLogsHealthPath  = "/health"           // VictoriaLogs 健康检查路径
)

// ==================== HTTP Header 常量 ====================

const (
	HeaderAuthorization   = "Authorization"    // 认证头
	HeaderContentEncoding = "Content-Encoding" // 内容编码头
	HeaderContentType     = "Content-Type"     // 内容类型头
	HeaderXScopeOrgID     = "X-Scope-OrgID"    // Loki 多租户头
	HeaderZoOrgID         = "zo-org-id"        // OpenObserve 组织 ID 头
	HeaderAccountID       = "AccountID"        // VictoriaLogs 租户 ID 头
)

// ==================== 认证前缀常量 ====================

const (
	AuthPrefixBasic  = "Basic "  // Basic 认证前缀
	AuthPrefixAPIKey = "ApiKey " // API Key 认证前缀
	AuthPrefixBearer = "Bearer " // Bearer 认证前缀
)

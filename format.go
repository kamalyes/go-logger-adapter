/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\format.go
 * @Description: 统一的日志条目格式化工具，消除各插件重复的格式化代码
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package adapter

import (
	"encoding/json"
	"time"

	"github.com/kamalyes/go-logger-adapter/constants"
)

// FormatTimestamp 将毫秒时间戳格式化为 RFC3339Nano 字符串
func FormatTimestamp(ts int64) string {
	return time.Unix(ts/constants.TimestampMsDivisor, (ts%constants.TimestampMsDivisor)*constants.TimestampNsMultiplier).UTC().Format(time.RFC3339Nano)
}

// FormatEntryToMap 将 LogEntry 转换为通用 map 结构
// 各插件可在此基础上添加自定义字段
func FormatEntryToMap(entry *LogEntry, timestampKey string) map[string]interface{} {
	doc := map[string]interface{}{
		timestampKey: FormatTimestamp(entry.Timestamp),
		"message":    entry.Message,
	}

	if entry.Caller != nil {
		doc["caller"] = map[string]interface{}{
			"file":     entry.Caller.File,
			"line":     entry.Caller.Line,
			"function": entry.Caller.Function,
		}
	}

	if len(entry.Fields) > 0 {
		for k, v := range entry.Fields {
			doc[k] = v
		}
	}

	return doc
}

// MarshalEntryToJSON 将 LogEntry 序列化为 JSON 字节
func MarshalEntryToJSON(entry *LogEntry, timestampKey string) ([]byte, error) {
	doc := FormatEntryToMap(entry, timestampKey)
	return json.Marshal(doc)
}

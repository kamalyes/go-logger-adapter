/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-12-06 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-06 00:00:00
 * @FilePath: \go-logger-adapter\es\format.go
 * @Description: Elasticsearch Bulk API 格式化器
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	adapter "github.com/kamalyes/go-logger-adapter"
	"github.com/kamalyes/go-logger-adapter/constants"
	"github.com/kamalyes/go-toolbox/pkg/errorx"
	"time"
)

// ESAction Elasticsearch Bulk API 的 action 行
type ESAction struct {
	Index ESActionIndex `json:"index"`
}

// ESActionIndex Elasticsearch Bulk API 的 index 指令
type ESActionIndex struct {
	Index string `json:"_index"`
}

// FormatBulkBody 将日志条目格式化为 Elasticsearch Bulk API 所需的 NDJSON 格式
func FormatBulkBody(entries []adapter.LogEntry, indexFormat string) ([]byte, error) {
	var buf bytes.Buffer
	now := time.Now()
	indexName := fmt.Sprintf(indexFormat, now.Format(constants.ESIndexDateFormat))

	for i := range entries {
		entry := &entries[i]

		action := ESAction{
			Index: ESActionIndex{
				Index: indexName,
			},
		}
		actionData, err := json.Marshal(action)
		if err != nil {
			return nil, errorx.NewTypedError(adapter.ErrTypeMarshalFailed, adapter.ErrFmtMarshalActionFail, constants.PluginNameElasticsearch, err)
		}
		buf.Write(actionData)
		buf.WriteByte('\n')

		doc := adapter.FormatEntryToMap(entry, "@timestamp")
		doc["log_level"] = entry.Level.String()
		entryData, err := json.Marshal(doc)
		if err != nil {
			return nil, errorx.NewTypedError(adapter.ErrTypeMarshalFailed, adapter.ErrFmtMarshalEntryFail, constants.PluginNameElasticsearch, err)
		}
		buf.Write(entryData)
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}

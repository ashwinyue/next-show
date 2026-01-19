// Package builtin 提供内置工具和中间件.
package builtin

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/reduction"
)

// MiddlewareConfig 中间件配置.
type MiddlewareConfig struct {
	// ClearToolResult 清理旧工具结果配置
	ClearToolResult *ClearToolResultConfig

	// TODO: 更多中间件配置
	// LargeToolResult *LargeToolResultConfig
	// MessageTrimming *MessageTrimmingConfig
}

// ClearToolResultConfig 清理工具结果配置.
type ClearToolResultConfig struct {
	Enabled                    bool     `json:"enabled"`
	ToolResultTokenThreshold   int      `json:"tool_result_token_threshold"`   // 工具结果总 token 阈值
	KeepRecentTokens           int      `json:"keep_recent_tokens"`            // 保留最近消息的 token 数
	ClearToolResultPlaceholder string   `json:"clear_tool_result_placeholder"` // 清理后的占位符
	ExcludeTools               []string `json:"exclude_tools"`                 // 排除的工具列表
}

// DefaultMiddlewareConfig 返回默认中间件配置.
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		ClearToolResult: &ClearToolResultConfig{
			Enabled:                    true,
			ToolResultTokenThreshold:   20000,
			KeepRecentTokens:           40000,
			ClearToolResultPlaceholder: "[旧工具结果已清理]",
			ExcludeTools:               []string{ToolThinking, ToolTodoWrite},
		},
	}
}

// BuildMiddlewares 根据配置构建中间件列表.
func BuildMiddlewares(ctx context.Context, config *MiddlewareConfig) ([]adk.AgentMiddleware, error) {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	var middlewares []adk.AgentMiddleware

	// 1. 清理工具结果中间件
	if config.ClearToolResult != nil && config.ClearToolResult.Enabled {
		clearToolResult, err := reduction.NewClearToolResult(ctx, &reduction.ClearToolResultConfig{
			ToolResultTokenThreshold:   config.ClearToolResult.ToolResultTokenThreshold,
			KeepRecentTokens:           config.ClearToolResult.KeepRecentTokens,
			ClearToolResultPlaceholder: config.ClearToolResult.ClearToolResultPlaceholder,
			ExcludeTools:               config.ClearToolResult.ExcludeTools,
		})
		if err != nil {
			return nil, err
		}
		middlewares = append(middlewares, clearToolResult)
	}

	return middlewares, nil
}

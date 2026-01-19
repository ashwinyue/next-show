// Package middleware 提供 Agent 中间件配置和构建.
package middleware

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/filesystem"
	"github.com/cloudwego/eino/adk/middlewares/reduction"
	"github.com/cloudwego/eino/compose"
)

// 默认排除的工具名称.
const (
	ToolThinking  = "thinking"
	ToolTodoWrite = "todo_write"
)

// Config 中间件配置.
type Config struct {
	ClearToolResult *ClearToolResultConfig
	LargeToolResult *LargeToolResultConfig
	MessageTrimming *MessageTrimmingConfig
}

// LargeToolResultConfig 大工具结果卸载配置.
type LargeToolResultConfig struct {
	Enabled       bool               `json:"enabled"`
	Backend       filesystem.Backend `json:"-"`
	TokenLimit    int                `json:"token_limit"`
	PathGenerator func(ctx context.Context, input *compose.ToolInput) (string, error)
}

// MessageTrimmingConfig 消息裁剪配置.
type MessageTrimmingConfig struct {
	Enabled   bool `json:"enabled"`
	MaxItems  int  `json:"max_items"`
	MaxTokens int  `json:"max_tokens"`
}

// ClearToolResultConfig 清理工具结果配置.
type ClearToolResultConfig struct {
	Enabled                    bool     `json:"enabled"`
	ToolResultTokenThreshold   int      `json:"tool_result_token_threshold"`
	KeepRecentTokens           int      `json:"keep_recent_tokens"`
	ClearToolResultPlaceholder string   `json:"clear_tool_result_placeholder"`
	ExcludeTools               []string `json:"exclude_tools"`
}

// DefaultConfig 返回默认中间件配置.
func DefaultConfig() *Config {
	return &Config{
		ClearToolResult: &ClearToolResultConfig{
			Enabled:                    true,
			ToolResultTokenThreshold:   20000,
			KeepRecentTokens:           40000,
			ClearToolResultPlaceholder: "[旧工具结果已清理]",
			ExcludeTools:               []string{ToolThinking, ToolTodoWrite},
		},
	}
}

// Build 根据配置构建中间件列表.
func Build(ctx context.Context, config *Config) ([]adk.AgentMiddleware, error) {
	if config == nil {
		config = DefaultConfig()
	}

	var middlewares []adk.AgentMiddleware

	if config.LargeToolResult != nil && config.LargeToolResult.Enabled {
		m, err := filesystem.NewMiddleware(ctx, &filesystem.Config{
			Backend:                             config.LargeToolResult.Backend,
			LargeToolResultOffloadingTokenLimit: config.LargeToolResult.TokenLimit,
			LargeToolResultOffloadingPathGen:    config.LargeToolResult.PathGenerator,
		})
		if err != nil {
			return nil, err
		}
		middlewares = append(middlewares, m)
	}

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

	if config.MessageTrimming != nil && config.MessageTrimming.Enabled {
		cfg := *config.MessageTrimming
		middlewares = append(middlewares, adk.AgentMiddleware{
			BeforeChatModel: func(ctx context.Context, state *adk.ChatModelAgentState) error {
				if state == nil || len(state.Messages) == 0 {
					return nil
				}

				msgs := state.Messages
				if cfg.MaxItems > 0 {
					msgs = TrimMessagesWithConsistency(msgs, cfg.MaxItems)
					msgs = RemoveOrphanedToolMessages(msgs)
				}
				if cfg.MaxTokens > 0 {
					msgs = TrimToTokenLimit(msgs, cfg.MaxTokens, EstimateMessagesTokens)
					msgs = RemoveOrphanedToolMessages(msgs)
				}
				state.Messages = msgs
				return nil
			},
		})
	}

	return middlewares, nil
}

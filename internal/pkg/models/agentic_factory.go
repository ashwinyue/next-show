// Package models 提供 AgenticModel 工厂，支持 ARK 和 OpenAI。
package models

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/agenticark"
	"github.com/cloudwego/eino-ext/components/model/agenticopenai"
	"github.com/cloudwego/eino/components/model"
)

// ModelConfig 模型配置。
type ModelConfig struct {
	Provider string `json:"provider"` // "openai", "ark"
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url,omitempty"`

	// Agentic 特有配置
	Thinking    bool     `json:"thinking,omitempty"`     // 启用推理模式
	ServerTools []string `json:"server_tools,omitempty"` // ["web_search"]
}

// CreateAgenticModel 创建 AgenticModel。
func CreateAgenticModel(ctx context.Context, cfg *ModelConfig) (model.AgenticModel, error) {
	switch cfg.Provider {
	case "ark":
		return createARKAgentic(ctx, cfg)
	case "openai":
		return createOpenAIAgentic(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// createARKAgentic 创建 ARK AgenticModel。
func createARKAgentic(ctx context.Context, cfg *ModelConfig) (model.AgenticModel, error) {
	arkConfig := &agenticark.Config{
		Model:  cfg.Model,
		APIKey: cfg.APIKey,
	}

	// 推理模式配置
	if cfg.Thinking {
		// TODO: 根据 ARK API 更新推理模式配置
		// 目前 ARK AgenticModel 可能不支持推理模式配置
	}

	// Server Tools 配置
	for _, tool := range cfg.ServerTools {
		switch tool {
		case "web_search":
			// TODO: 根据 ARK API 更新 Server Tool 配置
			// 目前 ARK AgenticModel 可能不支持 Server Tool 配置
		}
	}

	return agenticark.New(ctx, arkConfig)
}

// createOpenAIAgentic 创建 OpenAI AgenticModel。
func createOpenAIAgentic(ctx context.Context, cfg *ModelConfig) (model.AgenticModel, error) {
	openaiConfig := &agenticopenai.Config{
		Model:  cfg.Model,
		APIKey: cfg.APIKey,
	}

	if cfg.BaseURL != "" {
		openaiConfig.BaseURL = cfg.BaseURL
	}

	return agenticopenai.New(ctx, openaiConfig)
}

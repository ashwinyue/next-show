// Package factory 提供 Agent 和 Provider 工厂.
package factory

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"

	modelDef "github.com/mervyn/next-show/internal/model"
)

// ChatModelFactory ChatModel 工厂.
type ChatModelFactory struct{}

// NewChatModelFactory 创建 ChatModel 工厂.
func NewChatModelFactory() *ChatModelFactory {
	return &ChatModelFactory{}
}

// CreateChatModel 根据 Provider 配置创建 ToolCallingChatModel.
func (f *ChatModelFactory) CreateChatModel(ctx context.Context, provider *modelDef.Provider, modelName string) (model.ToolCallingChatModel, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is nil")
	}

	if modelName == "" {
		modelName = provider.DefaultModel
	}

	switch provider.ProviderType {
	case "openai", "claude", "deepseek":
		return f.createOpenAI(ctx, provider, modelName)
	case "ark":
		return f.createArk(ctx, provider, modelName)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", provider.ProviderType)
	}
}

func (f *ChatModelFactory) createOpenAI(ctx context.Context, provider *modelDef.Provider, modelName string) (model.ToolCallingChatModel, error) {
	cfg := &openai.ChatModelConfig{
		Model:   modelName,
		APIKey:  provider.APIKey,
		BaseURL: provider.BaseURL,
	}

	// 从 provider.Config 读取额外配置
	if provider.Config != nil {
		if temp, ok := provider.Config["temperature"].(float64); ok {
			t := float32(temp)
			cfg.Temperature = &t
		}
		if maxTokens, ok := provider.Config["max_tokens"].(float64); ok {
			mt := int(maxTokens)
			cfg.MaxTokens = &mt
		}
	}

	return openai.NewChatModel(ctx, cfg)
}

func (f *ChatModelFactory) createArk(ctx context.Context, provider *modelDef.Provider, modelName string) (model.ToolCallingChatModel, error) {
	cfg := &ark.ChatModelConfig{
		Model:   modelName,
		APIKey:  provider.APIKey,
		BaseURL: provider.BaseURL,
	}

	return ark.NewChatModel(ctx, cfg)
}

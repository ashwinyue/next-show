// Package embedding 提供 Embedding 模型工厂.
package embedding

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

// ProviderType Embedding 提供商类型.
type ProviderType string

const (
	ProviderDashScope ProviderType = "dashscope"
	ProviderOpenAI    ProviderType = "openai"
)

// Config Embedding 配置.
type Config struct {
	Provider   ProviderType  `json:"provider"`
	APIKey     string        `json:"api_key"`
	BaseURL    string        `json:"base_url,omitempty"`
	Model      string        `json:"model"`
	Dimensions int           `json:"dimensions,omitempty"`
	Timeout    time.Duration `json:"timeout,omitempty"`
}

// DefaultConfig 默认配置.
func DefaultConfig() *Config {
	return &Config{
		Provider:   ProviderDashScope,
		Model:      "text-embedding-v3",
		Dimensions: 1024,
		Timeout:    30 * time.Second,
	}
}

// Factory Embedding 工厂.
type Factory struct{}

// NewFactory 创建 Embedding 工厂.
func NewFactory() *Factory {
	return &Factory{}
}

// Create 根据配置创建 Embedding 模型.
func (f *Factory) Create(ctx context.Context, cfg *Config) (embedding.Embedder, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	switch cfg.Provider {
	case ProviderDashScope:
		return f.createDashScope(ctx, cfg)
	case ProviderOpenAI:
		return f.createOpenAI(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", cfg.Provider)
	}
}

func (f *Factory) createDashScope(ctx context.Context, cfg *Config) (embedding.Embedder, error) {
	dim := cfg.Dimensions
	if dim == 0 {
		dim = 1024
	}

	return dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		APIKey:     cfg.APIKey,
		Model:      cfg.Model,
		Dimensions: &dim,
		Timeout:    cfg.Timeout,
	})
}

func (f *Factory) createOpenAI(ctx context.Context, cfg *Config) (embedding.Embedder, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	dim := cfg.Dimensions
	ocfg := &openai.EmbeddingConfig{
		APIKey:  cfg.APIKey,
		BaseURL: baseURL,
		Model:   cfg.Model,
		Timeout: cfg.Timeout,
	}
	if dim > 0 {
		ocfg.Dimensions = &dim
	}

	return openai.NewEmbedder(ctx, ocfg)
}

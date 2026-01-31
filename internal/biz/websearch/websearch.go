// Package websearch 提供网络搜索配置业务逻辑（对齐 WeKnora）.
package websearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz 网络搜索配置业务接口.
type Biz interface {
	// GetConfig 获取网络搜索配置（单一配置，对齐 WeKnora）.
	GetConfig(ctx context.Context) (*WebSearchConfig, error)
	// UpdateConfig 更新网络搜索配置.
	UpdateConfig(ctx context.Context, req *UpdateConfigRequest) (*WebSearchConfig, error)
}

// WebSearchConfig 网络搜索配置（对齐 WeKnora）.
type WebSearchConfig struct {
	Provider   string   `json:"provider"`
	APIKey     string   `json:"api_key,omitempty"`
	MaxResults int      `json:"max_results"`
	Blacklist  []string `json:"blacklist,omitempty"`
}

// UpdateConfigRequest 更新网络搜索配置请求.
type UpdateConfigRequest struct {
	Provider   string   `json:"provider"`
	APIKey     string   `json:"api_key,omitempty"`
	MaxResults int      `json:"max_results"`
	Blacklist  []string `json:"blacklist,omitempty"`
}

const (
	// webSearchConfigKey 系统设置中的配置 Key.
	webSearchConfigKey = "web-search.config"
)

type bizImpl struct {
	store store.Store
}

// NewBiz 创建网络搜索配置业务实例.
func NewBiz(s store.Store) Biz {
	return &bizImpl{store: s}
}

// GetConfig 获取网络搜索配置.
func (b *bizImpl) GetConfig(ctx context.Context) (*WebSearchConfig, error) {
	setting, err := b.store.Settings().Get(ctx, webSearchConfigKey)
	if err != nil {
		// 返回默认配置
		return &WebSearchConfig{
			Provider:   "duckduckgo",
			MaxResults: 10,
		}, nil
	}

	var config WebSearchConfig
	if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// 设置默认值
	if config.MaxResults <= 0 {
		config.MaxResults = 10
	}

	return &config, nil
}

// UpdateConfig 更新网络搜索配置.
func (b *bizImpl) UpdateConfig(ctx context.Context, req *UpdateConfigRequest) (*WebSearchConfig, error) {
	config := &WebSearchConfig{
		Provider:   req.Provider,
		APIKey:     req.APIKey,
		MaxResults: req.MaxResults,
		Blacklist:  req.Blacklist,
	}

	if config.MaxResults <= 0 {
		config.MaxResults = 10
	}

	// 序列化为 JSON
	value, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	// 保存到系统设置
	setting := &model.SystemSettings{
		Key:       webSearchConfigKey,
		Value:     string(value),
		ValueType: "json",
		Category:  "feature",
		Label:     "Web Search",
	}

	// 尝试获取现有设置
	existing, _ := b.store.Settings().Get(ctx, webSearchConfigKey)
	if existing != nil {
		setting.ID = existing.ID
	}

	if err := b.store.Settings().Set(ctx, setting); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	return config, nil
}

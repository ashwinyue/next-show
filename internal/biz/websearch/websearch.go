// Package websearch 提供网络搜索配置业务逻辑.
package websearch

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz 网络搜索配置业务接口.
type Biz interface {
	// List 列出所有网络搜索配置.
	List(ctx context.Context) ([]*model.WebSearchConfig, error)
	// Get 获取网络搜索配置详情.
	Get(ctx context.Context, id string) (*model.WebSearchConfig, error)
	// GetDefault 获取默认网络搜索配置.
	GetDefault(ctx context.Context) (*model.WebSearchConfig, error)
	// Create 创建网络搜索配置.
	Create(ctx context.Context, req *CreateRequest) (*model.WebSearchConfig, error)
	// Update 更新网络搜索配置.
	Update(ctx context.Context, id string, req *UpdateRequest) (*model.WebSearchConfig, error)
	// Delete 删除网络搜索配置.
	Delete(ctx context.Context, id string) error
	// SetDefault 设置默认网络搜索配置.
	SetDefault(ctx context.Context, id string) error
}

// CreateRequest 创建网络搜索配置请求.
type CreateRequest struct {
	Name           string                  `json:"name"`
	DisplayName    string                  `json:"display_name"`
	Provider       model.WebSearchProvider `json:"provider"`
	APIKey         string                  `json:"api_key"`
	BaseURL        string                  `json:"base_url"`
	MaxResults     int                     `json:"max_results"`
	SearchDepth    string                  `json:"search_depth"`
	IncludeDomains model.JSONSlice         `json:"include_domains"`
	ExcludeDomains model.JSONSlice         `json:"exclude_domains"`
	Config         model.JSONMap           `json:"config"`
	IsDefault      bool                    `json:"is_default"`
}

// UpdateRequest 更新网络搜索配置请求.
type UpdateRequest struct {
	Name           *string                  `json:"name,omitempty"`
	DisplayName    *string                  `json:"display_name,omitempty"`
	Provider       *model.WebSearchProvider `json:"provider,omitempty"`
	APIKey         *string                  `json:"api_key,omitempty"`
	BaseURL        *string                  `json:"base_url,omitempty"`
	MaxResults     *int                     `json:"max_results,omitempty"`
	SearchDepth    *string                  `json:"search_depth,omitempty"`
	IncludeDomains model.JSONSlice          `json:"include_domains,omitempty"`
	ExcludeDomains model.JSONSlice          `json:"exclude_domains,omitempty"`
	Config         model.JSONMap            `json:"config,omitempty"`
	IsEnabled      *bool                    `json:"is_enabled,omitempty"`
}

type bizImpl struct {
	store store.Store
}

// NewBiz 创建网络搜索配置业务实例.
func NewBiz(s store.Store) Biz {
	return &bizImpl{store: s}
}

func (b *bizImpl) List(ctx context.Context) ([]*model.WebSearchConfig, error) {
	return b.store.WebSearch().List(ctx)
}

func (b *bizImpl) Get(ctx context.Context, id string) (*model.WebSearchConfig, error) {
	return b.store.WebSearch().Get(ctx, id)
}

func (b *bizImpl) GetDefault(ctx context.Context) (*model.WebSearchConfig, error) {
	return b.store.WebSearch().GetDefault(ctx)
}

func (b *bizImpl) Create(ctx context.Context, req *CreateRequest) (*model.WebSearchConfig, error) {
	config := &model.WebSearchConfig{
		ID:             uuid.New().String(),
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Provider:       req.Provider,
		APIKey:         req.APIKey,
		BaseURL:        req.BaseURL,
		MaxResults:     req.MaxResults,
		SearchDepth:    req.SearchDepth,
		IncludeDomains: req.IncludeDomains,
		ExcludeDomains: req.ExcludeDomains,
		Config:         req.Config,
		IsEnabled:      true,
		IsDefault:      req.IsDefault,
	}

	if config.MaxResults <= 0 {
		config.MaxResults = 10
	}
	if config.SearchDepth == "" {
		config.SearchDepth = "basic"
	}

	if err := b.store.WebSearch().Create(ctx, config); err != nil {
		return nil, fmt.Errorf("create web search config: %w", err)
	}

	// 如果设置为默认，更新其他配置
	if req.IsDefault {
		if err := b.store.WebSearch().SetDefault(ctx, config.ID); err != nil {
			return nil, fmt.Errorf("set default: %w", err)
		}
	}

	return config, nil
}

func (b *bizImpl) Update(ctx context.Context, id string, req *UpdateRequest) (*model.WebSearchConfig, error) {
	config, err := b.store.WebSearch().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		config.Name = *req.Name
	}
	if req.DisplayName != nil {
		config.DisplayName = *req.DisplayName
	}
	if req.Provider != nil {
		config.Provider = *req.Provider
	}
	if req.APIKey != nil {
		config.APIKey = *req.APIKey
	}
	if req.BaseURL != nil {
		config.BaseURL = *req.BaseURL
	}
	if req.MaxResults != nil {
		config.MaxResults = *req.MaxResults
	}
	if req.SearchDepth != nil {
		config.SearchDepth = *req.SearchDepth
	}
	if req.IncludeDomains != nil {
		config.IncludeDomains = req.IncludeDomains
	}
	if req.ExcludeDomains != nil {
		config.ExcludeDomains = req.ExcludeDomains
	}
	if req.Config != nil {
		config.Config = req.Config
	}
	if req.IsEnabled != nil {
		config.IsEnabled = *req.IsEnabled
	}

	if err := b.store.WebSearch().Update(ctx, config); err != nil {
		return nil, fmt.Errorf("update web search config: %w", err)
	}

	return config, nil
}

func (b *bizImpl) Delete(ctx context.Context, id string) error {
	return b.store.WebSearch().Delete(ctx, id)
}

func (b *bizImpl) SetDefault(ctx context.Context, id string) error {
	return b.store.WebSearch().SetDefault(ctx, id)
}

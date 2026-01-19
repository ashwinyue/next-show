// Package provider 提供 Provider 业务逻辑.
package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz Provider 业务接口.
type Biz interface {
	// ListProviders 列出所有 Provider.
	ListProviders(ctx context.Context) ([]*model.Provider, error)
	// GetProvider 获取 Provider 详情.
	GetProvider(ctx context.Context, id string) (*model.Provider, error)
	// CreateProvider 创建 Provider.
	CreateProvider(ctx context.Context, req *CreateProviderRequest) (*model.Provider, error)
	// UpdateProvider 更新 Provider.
	UpdateProvider(ctx context.Context, id string, req *UpdateProviderRequest) (*model.Provider, error)
	// DeleteProvider 删除 Provider.
	DeleteProvider(ctx context.Context, id string) error

	// ListChatProviders 列出对话模型 Provider.
	ListChatProviders(ctx context.Context) ([]*model.Provider, error)
	// ListEmbeddingProviders 列出 Embedding 模型 Provider.
	ListEmbeddingProviders(ctx context.Context) ([]*model.Provider, error)
	// ListRerankProviders 列出 Rerank 模型 Provider.
	ListRerankProviders(ctx context.Context) ([]*model.Provider, error)
}

// CreateProviderRequest 创建 Provider 请求.
type CreateProviderRequest struct {
	Name          string              `json:"name"`
	DisplayName   string              `json:"display_name"`
	ProviderType  string              `json:"provider_type"`
	ModelCategory model.ModelCategory `json:"model_category"`
	BaseURL       string              `json:"base_url"`
	APIKey        string              `json:"api_key"`
	APISecret     string              `json:"api_secret"`
	DefaultModel  string              `json:"default_model"`
	Config        model.JSONMap       `json:"config"`
}

// UpdateProviderRequest 更新 Provider 请求.
type UpdateProviderRequest struct {
	Name          *string              `json:"name,omitempty"`
	DisplayName   *string              `json:"display_name,omitempty"`
	ProviderType  *string              `json:"provider_type,omitempty"`
	ModelCategory *model.ModelCategory `json:"model_category,omitempty"`
	BaseURL       *string              `json:"base_url,omitempty"`
	APIKey        *string              `json:"api_key,omitempty"`
	APISecret     *string              `json:"api_secret,omitempty"`
	DefaultModel  *string              `json:"default_model,omitempty"`
	Config        model.JSONMap        `json:"config,omitempty"`
	IsEnabled     *bool                `json:"is_enabled,omitempty"`
}

type bizImpl struct {
	store store.Store
}

// NewBiz 创建 Provider 业务实例.
func NewBiz(s store.Store) Biz {
	return &bizImpl{store: s}
}

func (b *bizImpl) ListProviders(ctx context.Context) ([]*model.Provider, error) {
	return b.store.Providers().ListAll(ctx)
}

func (b *bizImpl) GetProvider(ctx context.Context, id string) (*model.Provider, error) {
	return b.store.Providers().Get(ctx, id)
}

func (b *bizImpl) CreateProvider(ctx context.Context, req *CreateProviderRequest) (*model.Provider, error) {
	provider := &model.Provider{
		ID:            uuid.New().String(),
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		ProviderType:  req.ProviderType,
		ModelCategory: req.ModelCategory,
		BaseURL:       req.BaseURL,
		APIKey:        req.APIKey,
		APISecret:     req.APISecret,
		DefaultModel:  req.DefaultModel,
		Config:        req.Config,
		IsEnabled:     true,
	}

	if provider.ModelCategory == "" {
		provider.ModelCategory = model.ModelCategoryChat
	}

	if err := b.store.Providers().Create(ctx, provider); err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	return provider, nil
}

func (b *bizImpl) UpdateProvider(ctx context.Context, id string, req *UpdateProviderRequest) (*model.Provider, error) {
	provider, err := b.store.Providers().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.DisplayName != nil {
		provider.DisplayName = *req.DisplayName
	}
	if req.ProviderType != nil {
		provider.ProviderType = *req.ProviderType
	}
	if req.ModelCategory != nil {
		provider.ModelCategory = *req.ModelCategory
	}
	if req.BaseURL != nil {
		provider.BaseURL = *req.BaseURL
	}
	if req.APIKey != nil {
		provider.APIKey = *req.APIKey
	}
	if req.APISecret != nil {
		provider.APISecret = *req.APISecret
	}
	if req.DefaultModel != nil {
		provider.DefaultModel = *req.DefaultModel
	}
	if req.Config != nil {
		provider.Config = req.Config
	}
	if req.IsEnabled != nil {
		provider.IsEnabled = *req.IsEnabled
	}

	if err := b.store.Providers().Update(ctx, provider); err != nil {
		return nil, fmt.Errorf("update provider: %w", err)
	}

	return provider, nil
}

func (b *bizImpl) DeleteProvider(ctx context.Context, id string) error {
	return b.store.Providers().Delete(ctx, id)
}

func (b *bizImpl) ListChatProviders(ctx context.Context) ([]*model.Provider, error) {
	return b.store.Providers().ListByCategory(ctx, model.ModelCategoryChat)
}

func (b *bizImpl) ListEmbeddingProviders(ctx context.Context) ([]*model.Provider, error) {
	return b.store.Providers().ListByCategory(ctx, model.ModelCategoryEmbedding)
}

func (b *bizImpl) ListRerankProviders(ctx context.Context) ([]*model.Provider, error) {
	return b.store.Providers().ListByCategory(ctx, model.ModelCategoryRerank)
}

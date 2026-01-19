// Package tenant 提供租户管理业务逻辑.
package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz 租户业务接口.
type Biz interface {
	// Tenant
	Create(ctx context.Context, req *CreateTenantRequest) (*model.Tenant, error)
	Get(ctx context.Context, id string) (*model.Tenant, error)
	List(ctx context.Context) ([]*model.Tenant, error)
	Update(ctx context.Context, id string, req *UpdateTenantRequest) (*model.Tenant, error)
	Delete(ctx context.Context, id string) error

	// API Key
	CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyWithSecret, error)
	GetAPIKey(ctx context.Context, id string) (*model.APIKey, error)
	ListAPIKeys(ctx context.Context, tenantID string) ([]*model.APIKey, error)
	RevokeAPIKey(ctx context.Context, id string) error
	DeleteAPIKey(ctx context.Context, id string) error
	ValidateAPIKey(ctx context.Context, key string) (*model.APIKey, error)
}

// CreateTenantRequest 创建租户请求.
type CreateTenantRequest struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Description string        `json:"description"`
	Config      model.JSONMap `json:"config"`
	Quota       model.JSONMap `json:"quota"`
}

// UpdateTenantRequest 更新租户请求.
type UpdateTenantRequest struct {
	Name        *string             `json:"name"`
	DisplayName *string             `json:"display_name"`
	Description *string             `json:"description"`
	Status      *model.TenantStatus `json:"status"`
	Config      model.JSONMap       `json:"config"`
	Quota       model.JSONMap       `json:"quota"`
}

// CreateAPIKeyRequest 创建 API Key 请求.
type CreateAPIKeyRequest struct {
	TenantID    string          `json:"tenant_id"`
	Name        string          `json:"name"`
	Permissions model.JSONSlice `json:"permissions"`
	RateLimit   int             `json:"rate_limit"`
	ExpiresAt   *string         `json:"expires_at"` // RFC3339 格式
}

// APIKeyWithSecret API Key 带密钥（仅创建时返回）.
type APIKeyWithSecret struct {
	*model.APIKey
	Secret string `json:"secret"` // 完整的 API Key，仅创建时返回一次
}

type bizImpl struct {
	store store.Store
}

// NewBiz 创建租户业务实例.
func NewBiz(s store.Store) Biz {
	return &bizImpl{store: s}
}

// Tenant 相关方法

func (b *bizImpl) Create(ctx context.Context, req *CreateTenantRequest) (*model.Tenant, error) {
	tenant := &model.Tenant{
		ID:          uuid.New().String(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Status:      model.TenantStatusActive,
		Config:      req.Config,
		Quota:       req.Quota,
	}

	if err := b.store.Tenants().Create(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (b *bizImpl) Get(ctx context.Context, id string) (*model.Tenant, error) {
	return b.store.Tenants().Get(ctx, id)
}

func (b *bizImpl) List(ctx context.Context) ([]*model.Tenant, error) {
	return b.store.Tenants().List(ctx)
}

func (b *bizImpl) Update(ctx context.Context, id string, req *UpdateTenantRequest) (*model.Tenant, error) {
	tenant, err := b.store.Tenants().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.DisplayName != nil {
		tenant.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		tenant.Description = *req.Description
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}
	if req.Config != nil {
		tenant.Config = req.Config
	}
	if req.Quota != nil {
		tenant.Quota = req.Quota
	}

	if err := b.store.Tenants().Update(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (b *bizImpl) Delete(ctx context.Context, id string) error {
	return b.store.Tenants().Delete(ctx, id)
}

// API Key 相关方法

func (b *bizImpl) CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyWithSecret, error) {
	// 验证租户存在
	if _, err := b.store.Tenants().Get(ctx, req.TenantID); err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// 生成 API Key
	key, prefix, err := model.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	apiKey := &model.APIKey{
		ID:          uuid.New().String(),
		TenantID:    req.TenantID,
		Name:        req.Name,
		Key:         key,
		KeyPrefix:   prefix,
		Permissions: req.Permissions,
		RateLimit:   req.RateLimit,
		Status:      model.APIKeyStatusActive,
	}

	if apiKey.RateLimit <= 0 {
		apiKey.RateLimit = 100
	}

	if err := b.store.Tenants().CreateAPIKey(ctx, apiKey); err != nil {
		return nil, err
	}

	return &APIKeyWithSecret{
		APIKey: apiKey,
		Secret: key,
	}, nil
}

func (b *bizImpl) GetAPIKey(ctx context.Context, id string) (*model.APIKey, error) {
	return b.store.Tenants().GetAPIKey(ctx, id)
}

func (b *bizImpl) ListAPIKeys(ctx context.Context, tenantID string) ([]*model.APIKey, error) {
	return b.store.Tenants().ListAPIKeysByTenant(ctx, tenantID)
}

func (b *bizImpl) RevokeAPIKey(ctx context.Context, id string) error {
	apiKey, err := b.store.Tenants().GetAPIKey(ctx, id)
	if err != nil {
		return err
	}

	apiKey.Status = model.APIKeyStatusRevoked
	return b.store.Tenants().UpdateAPIKey(ctx, apiKey)
}

func (b *bizImpl) DeleteAPIKey(ctx context.Context, id string) error {
	return b.store.Tenants().DeleteAPIKey(ctx, id)
}

func (b *bizImpl) ValidateAPIKey(ctx context.Context, key string) (*model.APIKey, error) {
	apiKey, err := b.store.Tenants().GetAPIKeyByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	// 更新最后使用时间
	_ = b.store.Tenants().UpdateAPIKeyLastUsed(ctx, apiKey.ID)

	return apiKey, nil
}

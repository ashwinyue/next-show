// Package store 提供数据访问层.
package store

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// TenantStore 租户存储接口.
type TenantStore interface {
	// Tenant CRUD
	Create(ctx context.Context, tenant *model.Tenant) error
	Get(ctx context.Context, id string) (*model.Tenant, error)
	GetByName(ctx context.Context, name string) (*model.Tenant, error)
	List(ctx context.Context) ([]*model.Tenant, error)
	Update(ctx context.Context, tenant *model.Tenant) error
	Delete(ctx context.Context, id string) error

	// API Key CRUD
	CreateAPIKey(ctx context.Context, apiKey *model.APIKey) error
	GetAPIKey(ctx context.Context, id string) (*model.APIKey, error)
	GetAPIKeyByKey(ctx context.Context, key string) (*model.APIKey, error)
	ListAPIKeysByTenant(ctx context.Context, tenantID string) ([]*model.APIKey, error)
	UpdateAPIKey(ctx context.Context, apiKey *model.APIKey) error
	DeleteAPIKey(ctx context.Context, id string) error
	UpdateAPIKeyLastUsed(ctx context.Context, id string) error
}

type tenantStore struct {
	db *gorm.DB
}

func newTenantStore(db *gorm.DB) TenantStore {
	return &tenantStore{db: db}
}

// Tenant CRUD

func (s *tenantStore) Create(ctx context.Context, tenant *model.Tenant) error {
	return s.db.WithContext(ctx).Create(tenant).Error
}

func (s *tenantStore) Get(ctx context.Context, id string) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *tenantStore) GetByName(ctx context.Context, name string) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *tenantStore) List(ctx context.Context) ([]*model.Tenant, error) {
	var tenants []*model.Tenant
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (s *tenantStore) Update(ctx context.Context, tenant *model.Tenant) error {
	return s.db.WithContext(ctx).Save(tenant).Error
}

func (s *tenantStore) Delete(ctx context.Context, id string) error {
	// 先删除关联的 API Keys
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", id).Delete(&model.APIKey{}).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Delete(&model.Tenant{}, "id = ?", id).Error
}

// API Key CRUD

func (s *tenantStore) CreateAPIKey(ctx context.Context, apiKey *model.APIKey) error {
	return s.db.WithContext(ctx).Create(apiKey).Error
}

func (s *tenantStore) GetAPIKey(ctx context.Context, id string) (*model.APIKey, error) {
	var apiKey model.APIKey
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&apiKey).Error; err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (s *tenantStore) GetAPIKeyByKey(ctx context.Context, key string) (*model.APIKey, error) {
	var apiKey model.APIKey
	if err := s.db.WithContext(ctx).Where("`key` = ? AND status = ?", key, model.APIKeyStatusActive).First(&apiKey).Error; err != nil {
		return nil, err
	}
	// 检查是否过期
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, gorm.ErrRecordNotFound
	}
	return &apiKey, nil
}

func (s *tenantStore) ListAPIKeysByTenant(ctx context.Context, tenantID string) ([]*model.APIKey, error) {
	var apiKeys []*model.APIKey
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&apiKeys).Error; err != nil {
		return nil, err
	}
	return apiKeys, nil
}

func (s *tenantStore) UpdateAPIKey(ctx context.Context, apiKey *model.APIKey) error {
	return s.db.WithContext(ctx).Save(apiKey).Error
}

func (s *tenantStore) DeleteAPIKey(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&model.APIKey{}, "id = ?", id).Error
}

func (s *tenantStore) UpdateAPIKeyLastUsed(ctx context.Context, id string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&model.APIKey{}).Where("id = ?", id).Update("last_used_at", now).Error
}

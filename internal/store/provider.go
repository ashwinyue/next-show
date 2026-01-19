// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/mervyn/next-show/internal/model"
)

// ProviderStore Provider 存储接口.
type ProviderStore interface {
	Create(ctx context.Context, provider *model.Provider) error
	Get(ctx context.Context, id string) (*model.Provider, error)
	GetByName(ctx context.Context, name string) (*model.Provider, error)
	Update(ctx context.Context, provider *model.Provider) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.Provider, int64, error)
	ListEnabled(ctx context.Context) ([]*model.Provider, error)
}

type providerStore struct {
	db *gorm.DB
}

func newProviderStore(db *gorm.DB) ProviderStore {
	return &providerStore{db: db}
}

func (s *providerStore) Create(ctx context.Context, provider *model.Provider) error {
	return s.db.WithContext(ctx).Create(provider).Error
}

func (s *providerStore) Get(ctx context.Context, id string) (*model.Provider, error) {
	var provider model.Provider
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (s *providerStore) GetByName(ctx context.Context, name string) (*model.Provider, error) {
	var provider model.Provider
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (s *providerStore) Update(ctx context.Context, provider *model.Provider) error {
	return s.db.WithContext(ctx).Save(provider).Error
}

func (s *providerStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Provider{}).Error
}

func (s *providerStore) List(ctx context.Context, offset, limit int) ([]*model.Provider, int64, error) {
	var providers []*model.Provider
	var total int64

	db := s.db.WithContext(ctx).Model(&model.Provider{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&providers).Error; err != nil {
		return nil, 0, err
	}
	return providers, total, nil
}

func (s *providerStore) ListEnabled(ctx context.Context) ([]*model.Provider, error) {
	var providers []*model.Provider
	if err := s.db.WithContext(ctx).Where("is_enabled = ?", true).Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

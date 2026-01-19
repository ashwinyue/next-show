// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// WebSearchStore 网络搜索配置存储接口.
type WebSearchStore interface {
	Create(ctx context.Context, config *model.WebSearchConfig) error
	Get(ctx context.Context, id string) (*model.WebSearchConfig, error)
	GetByName(ctx context.Context, name string) (*model.WebSearchConfig, error)
	GetDefault(ctx context.Context) (*model.WebSearchConfig, error)
	Update(ctx context.Context, config *model.WebSearchConfig) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*model.WebSearchConfig, error)
	ListEnabled(ctx context.Context) ([]*model.WebSearchConfig, error)
	SetDefault(ctx context.Context, id string) error
}

type webSearchStore struct {
	db *gorm.DB
}

func newWebSearchStore(db *gorm.DB) WebSearchStore {
	return &webSearchStore{db: db}
}

func (s *webSearchStore) Create(ctx context.Context, config *model.WebSearchConfig) error {
	return s.db.WithContext(ctx).Create(config).Error
}

func (s *webSearchStore) Get(ctx context.Context, id string) (*model.WebSearchConfig, error) {
	var config model.WebSearchConfig
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *webSearchStore) GetByName(ctx context.Context, name string) (*model.WebSearchConfig, error) {
	var config model.WebSearchConfig
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *webSearchStore) GetDefault(ctx context.Context) (*model.WebSearchConfig, error) {
	var config model.WebSearchConfig
	if err := s.db.WithContext(ctx).Where("is_default = ? AND is_enabled = ?", true, true).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *webSearchStore) Update(ctx context.Context, config *model.WebSearchConfig) error {
	return s.db.WithContext(ctx).Save(config).Error
}

func (s *webSearchStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.WebSearchConfig{}).Error
}

func (s *webSearchStore) List(ctx context.Context) ([]*model.WebSearchConfig, error) {
	var configs []*model.WebSearchConfig
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (s *webSearchStore) ListEnabled(ctx context.Context) ([]*model.WebSearchConfig, error) {
	var configs []*model.WebSearchConfig
	if err := s.db.WithContext(ctx).Where("is_enabled = ?", true).Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (s *webSearchStore) SetDefault(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先清除所有默认
		if err := tx.Model(&model.WebSearchConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		// 设置新默认
		return tx.Model(&model.WebSearchConfig{}).Where("id = ?", id).Update("is_default", true).Error
	})
}

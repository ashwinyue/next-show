// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// SettingsStore 系统设置存储接口.
type SettingsStore interface {
	Get(ctx context.Context, key string) (*model.SystemSettings, error)
	Set(ctx context.Context, setting *model.SystemSettings) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context) ([]*model.SystemSettings, error)
	ListByCategory(ctx context.Context, category string) ([]*model.SystemSettings, error)
	GetMultiple(ctx context.Context, keys []string) ([]*model.SystemSettings, error)
}

type settingsStore struct {
	db *gorm.DB
}

func newSettingsStore(db *gorm.DB) SettingsStore {
	return &settingsStore{db: db}
}

func (s *settingsStore) Get(ctx context.Context, key string) (*model.SystemSettings, error) {
	var setting model.SystemSettings
	if err := s.db.WithContext(ctx).Where("`key` = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (s *settingsStore) Set(ctx context.Context, setting *model.SystemSettings) error {
	// Upsert: 如果存在则更新，否则创建
	return s.db.WithContext(ctx).Save(setting).Error
}

func (s *settingsStore) Delete(ctx context.Context, key string) error {
	return s.db.WithContext(ctx).Where("`key` = ?", key).Delete(&model.SystemSettings{}).Error
}

func (s *settingsStore) List(ctx context.Context) ([]*model.SystemSettings, error) {
	var settings []*model.SystemSettings
	if err := s.db.WithContext(ctx).Order("category, `key`").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *settingsStore) ListByCategory(ctx context.Context, category string) ([]*model.SystemSettings, error) {
	var settings []*model.SystemSettings
	if err := s.db.WithContext(ctx).Where("category = ?", category).Order("`key`").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *settingsStore) GetMultiple(ctx context.Context, keys []string) ([]*model.SystemSettings, error) {
	var settings []*model.SystemSettings
	if err := s.db.WithContext(ctx).Where("`key` IN ?", keys).Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

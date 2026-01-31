// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// SkillStore Skill 存储接口.
type SkillStore interface {
	Create(ctx context.Context, skill *model.Skill) error
	Get(ctx context.Context, id string) (*model.Skill, error)
	Update(ctx context.Context, skill *model.Skill) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.Skill, int64, error)
	ListAll(ctx context.Context) ([]*model.Skill, error)
	ListByCategory(ctx context.Context, category string) ([]*model.Skill, error)
	ListEnabled(ctx context.Context) ([]*model.Skill, error)
	Search(ctx context.Context, keyword string, offset, limit int) ([]*model.Skill, int64, error)
}

type skillStore struct {
	db *gorm.DB
}

func newSkillStore(db *gorm.DB) SkillStore {
	return &skillStore{db: db}
}

func (s *skillStore) Create(ctx context.Context, skill *model.Skill) error {
	return s.db.WithContext(ctx).Create(skill).Error
}

func (s *skillStore) Get(ctx context.Context, id string) (*model.Skill, error) {
	var skill model.Skill
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&skill).Error; err != nil {
		return nil, err
	}
	return &skill, nil
}

func (s *skillStore) Update(ctx context.Context, skill *model.Skill) error {
	return s.db.WithContext(ctx).Save(skill).Error
}

func (s *skillStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Skill{}).Error
}

func (s *skillStore) List(ctx context.Context, offset, limit int) ([]*model.Skill, int64, error) {
	var skills []*model.Skill
	var total int64

	db := s.db.WithContext(ctx).Model(&model.Skill{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&skills).Error; err != nil {
		return nil, 0, err
	}
	return skills, total, nil
}

func (s *skillStore) ListAll(ctx context.Context) ([]*model.Skill, error) {
	var skills []*model.Skill
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&skills).Error; err != nil {
		return nil, err
	}
	return skills, nil
}

func (s *skillStore) ListByCategory(ctx context.Context, category string) ([]*model.Skill, error) {
	var skills []*model.Skill
	if err := s.db.WithContext(ctx).Where("category = ?", category).Order("created_at DESC").Find(&skills).Error; err != nil {
		return nil, err
	}
	return skills, nil
}

func (s *skillStore) ListEnabled(ctx context.Context) ([]*model.Skill, error) {
	var skills []*model.Skill
	if err := s.db.WithContext(ctx).Where("is_enabled = ?", true).Order("created_at DESC").Find(&skills).Error; err != nil {
		return nil, err
	}
	return skills, nil
}

func (s *skillStore) Search(ctx context.Context, keyword string, offset, limit int) ([]*model.Skill, int64, error) {
	var skills []*model.Skill
	var total int64

	searchPattern := "%" + keyword + "%"
	db := s.db.WithContext(ctx).Model(&model.Skill{}).Where(
		"name LIKE ? OR description LIKE ? OR category LIKE ?",
		searchPattern, searchPattern, searchPattern,
	)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&skills).Error; err != nil {
		return nil, 0, err
	}
	return skills, total, nil
}

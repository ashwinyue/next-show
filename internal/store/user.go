// Package store 提供数据访问层.
package store

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// UserStore 用户存储接口.
type UserStore interface {
	// User CRUD
	Create(ctx context.Context, user *model.User) error
	Get(ctx context.Context, id string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, tenantID string) ([]*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
	UpdateLastLogin(ctx context.Context, id string) error

	// Token
	CreateToken(ctx context.Context, token *model.UserToken) error
	GetToken(ctx context.Context, tokenStr string) (*model.UserToken, error)
	RevokeToken(ctx context.Context, tokenStr string) error
	CleanExpiredTokens(ctx context.Context) error
}

type userStore struct {
	db *gorm.DB
}

func newUserStore(db *gorm.DB) UserStore {
	return &userStore{db: db}
}

// User CRUD

func (s *userStore) Create(ctx context.Context, user *model.User) error {
	return s.db.WithContext(ctx).Create(user).Error
}

func (s *userStore) Get(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userStore) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userStore) List(ctx context.Context, tenantID string) ([]*model.User, error) {
	var users []*model.User
	db := s.db.WithContext(ctx)
	if tenantID != "" {
		db = db.Where("tenant_id = ?", tenantID)
	}
	if err := db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userStore) Update(ctx context.Context, user *model.User) error {
	return s.db.WithContext(ctx).Save(user).Error
}

func (s *userStore) Delete(ctx context.Context, id string) error {
	// 先删除用户的 tokens
	if err := s.db.WithContext(ctx).Where("user_id = ?", id).Delete(&model.UserToken{}).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}

func (s *userStore) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("last_login_at", now).Error
}

// Token 相关

func (s *userStore) CreateToken(ctx context.Context, token *model.UserToken) error {
	return s.db.WithContext(ctx).Create(token).Error
}

func (s *userStore) GetToken(ctx context.Context, tokenStr string) (*model.UserToken, error) {
	var token model.UserToken
	if err := s.db.WithContext(ctx).Where("token = ? AND revoked_at IS NULL AND expires_at > ?", tokenStr, time.Now()).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *userStore) RevokeToken(ctx context.Context, tokenStr string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&model.UserToken{}).Where("token = ?", tokenStr).Update("revoked_at", now).Error
}

func (s *userStore) CleanExpiredTokens(ctx context.Context) error {
	return s.db.WithContext(ctx).Where("expires_at < ? OR revoked_at IS NOT NULL", time.Now().Add(-24*time.Hour)).Delete(&model.UserToken{}).Error
}

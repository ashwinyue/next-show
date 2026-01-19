// Package model 定义数据模型.
package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserRole 用户角色.
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
	UserRoleGuest UserRole = "guest"
)

// UserStatus 用户状态.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBanned   UserStatus = "banned"
)

// User 用户.
type User struct {
	ID           string     `json:"id" gorm:"primaryKey;size:36"`
	TenantID     string     `json:"tenant_id" gorm:"size:36;index"`
	Username     string     `json:"username" gorm:"size:100;not null;uniqueIndex"`
	Email        string     `json:"email" gorm:"size:200;not null;uniqueIndex"`
	PasswordHash string     `json:"-" gorm:"size:200;not null"` // 不返回给前端
	DisplayName  string     `json:"display_name" gorm:"size:200"`
	Avatar       string     `json:"avatar" gorm:"size:500"`
	Role         UserRole   `json:"role" gorm:"size:20;default:user"`
	Status       UserStatus `json:"status" gorm:"size:20;default:active;index"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

func (User) TableName() string {
	return "users"
}

// SetPassword 设置密码（哈希）.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword 验证密码.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UserToken 用户令牌（用于存储已撤销的令牌）.
type UserToken struct {
	ID        string     `json:"id" gorm:"primaryKey;size:36"`
	UserID    string     `json:"user_id" gorm:"size:36;not null;index"`
	Token     string     `json:"-" gorm:"size:500;not null;uniqueIndex"`
	Type      string     `json:"type" gorm:"size:20;default:access"` // access, refresh
	ExpiresAt time.Time  `json:"expires_at" gorm:"index"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`

	// 关联
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserToken) TableName() string {
	return "user_tokens"
}

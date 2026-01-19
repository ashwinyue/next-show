// Package model 定义数据模型.
package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// TenantStatus 租户状态.
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
)

// Tenant 租户.
type Tenant struct {
	ID          string       `json:"id" gorm:"primaryKey;size:36"`
	Name        string       `json:"name" gorm:"size:100;not null;uniqueIndex"`
	DisplayName string       `json:"display_name" gorm:"size:200"`
	Description string       `json:"description" gorm:"size:500"`
	Status      TenantStatus `json:"status" gorm:"size:20;default:active;index"`
	Config      JSONMap      `json:"config" gorm:"type:jsonb"`
	Quota       JSONMap      `json:"quota" gorm:"type:jsonb"`       // 配额限制
	UsageStats  JSONMap      `json:"usage_stats" gorm:"type:jsonb"` // 使用统计
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (Tenant) TableName() string {
	return "tenants"
}

// APIKeyStatus API Key 状态.
type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
	APIKeyStatusExpired APIKeyStatus = "expired"
)

// APIKey API 密钥.
type APIKey struct {
	ID          string       `json:"id" gorm:"primaryKey;size:36"`
	TenantID    string       `json:"tenant_id" gorm:"size:36;not null;index"`
	Name        string       `json:"name" gorm:"size:100;not null"`
	Key         string       `json:"-" gorm:"size:64;not null;uniqueIndex"` // 不返回给前端
	KeyPrefix   string       `json:"key_prefix" gorm:"size:10"`             // 前缀用于显示，如 "sk-xxx..."
	Permissions JSONSlice    `json:"permissions" gorm:"type:jsonb"`         // 权限列表
	RateLimit   int          `json:"rate_limit" gorm:"default:100"`         // 每分钟请求限制
	Status      APIKeyStatus `json:"status" gorm:"size:20;default:active;index"`
	LastUsedAt  *time.Time   `json:"last_used_at"`
	ExpiresAt   *time.Time   `json:"expires_at" gorm:"index"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`

	// 关联
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

// GenerateAPIKey 生成 API Key.
func GenerateAPIKey() (key string, prefix string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	key = "sk-" + hex.EncodeToString(bytes)
	prefix = key[:10] + "..."
	return key, prefix, nil
}

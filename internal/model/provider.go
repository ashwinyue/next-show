// Package model 定义数据模型.
package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ModelCategory 模型类别.
type ModelCategory string

const (
	ModelCategoryChat      ModelCategory = "chat"      // 对话模型
	ModelCategoryEmbedding ModelCategory = "embedding" // Embedding 模型
	ModelCategoryRerank    ModelCategory = "rerank"    // Rerank 模型
)

// Provider 模型供应商配置.
type Provider struct {
	ID            string        `json:"id" gorm:"primaryKey;size:36"`
	Name          string        `json:"name" gorm:"uniqueIndex;size:100;not null"`
	DisplayName   string        `json:"display_name" gorm:"size:200;not null"`
	ProviderType  string        `json:"provider_type" gorm:"size:50;not null;index"` // openai/claude/ark/gemini/ollama/qianfan/deepseek
	ModelCategory ModelCategory `json:"model_category" gorm:"size:20;not null;default:chat;index"`
	BaseURL       string        `json:"base_url" gorm:"size:500"`
	APIKey        string        `json:"-" gorm:"size:500"` // 不序列化到 JSON
	APISecret     string        `json:"-" gorm:"size:500"` // 不序列化到 JSON
	DefaultModel  string        `json:"default_model" gorm:"size:200"`
	Config        JSONMap       `json:"config" gorm:"type:json"`
	IsEnabled     bool          `json:"is_enabled" gorm:"default:true;index"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

func (Provider) TableName() string {
	return "providers"
}

// JSONMap JSON Map 类型.
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// JSONSlice JSON Slice 类型（用于存储字符串数组）.
type JSONSlice []string

func (j JSONSlice) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

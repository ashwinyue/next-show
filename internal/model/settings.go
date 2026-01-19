// Package model 定义数据模型.
package model

import "time"

// SystemSettings 系统设置.
type SystemSettings struct {
	ID          string    `json:"id" gorm:"primaryKey;size:36"`
	Key         string    `json:"key" gorm:"uniqueIndex;size:100;not null"`
	Value       string    `json:"value" gorm:"type:text"`
	ValueType   string    `json:"value_type" gorm:"size:20;default:string"` // string, int, bool, json
	Category    string    `json:"category" gorm:"size:50;index"`            // general, model, feature, ui
	Label       string    `json:"label" gorm:"size:200"`
	Description string    `json:"description" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SystemSettings) TableName() string {
	return "system_settings"
}

// 系统设置 Key 常量.
const (
	// 通用设置
	SettingKeySystemName        = "system.name"
	SettingKeySystemDescription = "system.description"
	SettingKeySystemLanguage    = "system.language"

	// 默认模型设置
	SettingKeyDefaultChatProvider      = "model.default_chat_provider"
	SettingKeyDefaultEmbeddingProvider = "model.default_embedding_provider"
	SettingKeyDefaultRerankProvider    = "model.default_rerank_provider"

	// 功能开关
	SettingKeyWebSearchEnabled  = "feature.web_search_enabled"
	SettingKeyKnowledgeEnabled  = "feature.knowledge_enabled"
	SettingKeyMultiAgentEnabled = "feature.multi_agent_enabled"
	SettingKeyStreamingEnabled  = "feature.streaming_enabled"

	// Ollama 配置
	SettingKeyOllamaBaseURL = "ollama.base_url"
	SettingKeyOllamaEnabled = "ollama.enabled"
)

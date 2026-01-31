// Package model 定义数据模型.
package model

import (
	"time"

	"github.com/google/uuid"
)

// Skill 技能配置，可复用的 Agent 能力.
type Skill struct {
	ID          string `json:"id" gorm:"primaryKey;size:36"`
	Name        string `json:"name" gorm:"size:100;not null;index"`
	Description string `json:"description" gorm:"size:500"`

	// 核心配置
	SystemPrompt string         `json:"system_prompt" gorm:"type:text"`
	Instructions string         `json:"instructions" gorm:"type:text"`
	Examples     []SkillExample `json:"examples" gorm:"type:jsonb;serializer:json"`

	// 工具和知识库
	ToolIDs          []string `json:"tool_ids" gorm:"type:jsonb;serializer:json"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids" gorm:"type:jsonb;serializer:json"`

	// 模型配置
	ModelProvider string  `json:"model_provider" gorm:"size:50;default:''"`
	ModelName     string  `json:"model_name" gorm:"size:100;default:''"`
	Temperature   float64 `json:"temperature" gorm:"default:0.7"`
	MaxIterations int     `json:"max_iterations" gorm:"default:10"`

	// 元数据
	Category  string    `json:"category" gorm:"size:50;index;default:'general'"` // writing, analysis, coding, general
	Tags      []string  `json:"tags" gorm:"type:jsonb;serializer:json"`
	IsEnabled bool      `json:"is_enabled" gorm:"default:true;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SkillExample 示例，用于 few-shot learning.
type SkillExample struct {
	ID     string `json:"id"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

// BeforeCreate GORM hook，生成 UUID.
func (s *Skill) BeforeCreate() error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名.
func (Skill) TableName() string {
	return "skills"
}

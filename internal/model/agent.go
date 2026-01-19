// Package model 定义数据模型.
package model

import (
	"time"
)

// AgentType Agent 类型枚举.
type AgentType string

const (
	AgentTypeChatModel   AgentType = "chat_model"
	AgentTypeReact       AgentType = "react"
	AgentTypePlanExecute AgentType = "plan_execute"
	AgentTypeDeep        AgentType = "deep"
	AgentTypeSupervisor  AgentType = "supervisor"
	AgentTypeSequential  AgentType = "sequential"
	AgentTypeLoop        AgentType = "loop"
	AgentTypeRAG         AgentType = "rag"
	AgentTypeCustom      AgentType = "custom"
)

// Agent Agent 配置.
type Agent struct {
	ID            string    `json:"id" gorm:"primaryKey;size:36"`
	Name          string    `json:"name" gorm:"uniqueIndex;size:100;not null"`
	DisplayName   string    `json:"display_name" gorm:"size:200;not null"`
	Description   string    `json:"description" gorm:"type:text"`
	ProviderID    string    `json:"provider_id" gorm:"size:36;not null;index"`
	ModelName     string    `json:"model_name" gorm:"size:200;not null"`
	SystemPrompt  string    `json:"system_prompt" gorm:"type:text"`
	AgentType     AgentType `json:"agent_type" gorm:"size:50;not null;default:chat_model;index"`
	MaxIterations int       `json:"max_iterations" gorm:"default:10"`
	Temperature   *float64  `json:"temperature" gorm:"type:decimal(3,2)"`
	MaxTokens     *int      `json:"max_tokens"`
	Config        JSONMap   `json:"config" gorm:"type:json"`
	IsEnabled     bool      `json:"is_enabled" gorm:"default:true;index"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联
	Provider *Provider `json:"provider,omitempty" gorm:"foreignKey:ProviderID"`
}

func (Agent) TableName() string {
	return "agents"
}

// AgentRelation Agent 编排关系（组合型 Agent）.
type AgentRelation struct {
	ID            string    `json:"id" gorm:"primaryKey;size:36"`
	ParentAgentID string    `json:"parent_agent_id" gorm:"size:36;not null;index"`
	ChildAgentID  string    `json:"child_agent_id" gorm:"size:36;not null;index"`
	Role          string    `json:"role" gorm:"size:50;not null"` // planner/executor/replanner/supervisor/sub_agent/step
	SortOrder     int       `json:"sort_order" gorm:"default:0"`
	Config        JSONMap   `json:"config" gorm:"type:json"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联
	ParentAgent *Agent `json:"parent_agent,omitempty" gorm:"foreignKey:ParentAgentID"`
	ChildAgent  *Agent `json:"child_agent,omitempty" gorm:"foreignKey:ChildAgentID"`
}

func (AgentRelation) TableName() string {
	return "agent_relations"
}

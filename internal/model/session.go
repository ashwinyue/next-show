// Package model 定义数据模型.
package model

import "time"

// SessionStatus 会话状态.
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusArchived SessionStatus = "archived"
	SessionStatusDeleted  SessionStatus = "deleted"
)

// Session 会话模型.
type Session struct {
	ID        string        `json:"id" gorm:"primaryKey;size:36"`
	AgentID   string        `json:"agent_id" gorm:"size:36;not null;index"`
	UserID    string        `json:"user_id" gorm:"size:100;index"`
	Title     string        `json:"title" gorm:"size:500"`
	Status    SessionStatus `json:"status" gorm:"size:20;not null;default:active;index"`
	Metadata  JSONMap       `json:"metadata" gorm:"type:json"`
	Context   JSONMap       `json:"context" gorm:"type:json"` // SessionValues
	CreatedAt time.Time     `json:"created_at" gorm:"index"`
	UpdatedAt time.Time     `json:"updated_at"`

	// 关联
	Agent *Agent `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

// TableName 返回表名.
func (Session) TableName() string {
	return "sessions"
}

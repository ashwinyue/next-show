// Package model 定义数据模型.
package model

import "time"

// CheckpointStatus Checkpoint 状态.
type CheckpointStatus string

const (
	CheckpointStatusActive    CheckpointStatus = "active"
	CheckpointStatusResumed   CheckpointStatus = "resumed"
	CheckpointStatusExpired   CheckpointStatus = "expired"
	CheckpointStatusCompleted CheckpointStatus = "completed"
)

// Checkpoint Checkpoint 持久化（支持中断恢复）.
type Checkpoint struct {
	ID                 string           `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CheckpointID       string           `json:"checkpoint_id" gorm:"size:200;not null;uniqueIndex"`
	SessionID          *string          `json:"session_id,omitempty" gorm:"type:uuid;index"`
	AgentID            *string          `json:"agent_id,omitempty" gorm:"type:uuid;index"`
	Data               []byte           `json:"-" gorm:"type:bytea;not null"` // gob 编码的二进制数据
	DataHash           string           `json:"data_hash,omitempty" gorm:"size:64"`
	InterruptInfo      JSONMap          `json:"interrupt_info,omitempty" gorm:"type:jsonb"`
	InterruptAddresses JSONMap          `json:"interrupt_addresses,omitempty" gorm:"type:jsonb"`
	Status             CheckpointStatus `json:"status" gorm:"size:20;not null;default:active;index"`
	RunContext         JSONMap          `json:"run_context,omitempty" gorm:"type:jsonb"`
	Version            int              `json:"version" gorm:"not null;default:1"`
	Metadata           JSONMap          `json:"metadata,omitempty" gorm:"type:jsonb"`
	ExpiresAt          *time.Time       `json:"expires_at,omitempty" gorm:"index"`
	CreatedAt          time.Time        `json:"created_at" gorm:"index"`
	UpdatedAt          time.Time        `json:"updated_at"`

	// 关联
	Session *Session `json:"session,omitempty" gorm:"foreignKey:SessionID"`
	Agent   *Agent   `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

func (Checkpoint) TableName() string {
	return "checkpoints"
}

// CheckpointEventType Checkpoint 事件类型.
type CheckpointEventType string

const (
	CheckpointEventCreated         CheckpointEventType = "created"
	CheckpointEventLoaded          CheckpointEventType = "loaded"
	CheckpointEventResumed         CheckpointEventType = "resumed"
	CheckpointEventResumedWithData CheckpointEventType = "resumed_with_data"
	CheckpointEventExpired         CheckpointEventType = "expired"
	CheckpointEventDeleted         CheckpointEventType = "deleted"
)

// CheckpointEvent Checkpoint 事件日志.
type CheckpointEvent struct {
	ID           string              `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CheckpointID string              `json:"checkpoint_id" gorm:"size:200;not null;index"`
	EventType    CheckpointEventType `json:"event_type" gorm:"size:30;not null;index"`
	EventData    JSONMap             `json:"event_data,omitempty" gorm:"type:jsonb"`
	CreatedAt    time.Time           `json:"created_at" gorm:"index"`
}

func (CheckpointEvent) TableName() string {
	return "checkpoint_events"
}

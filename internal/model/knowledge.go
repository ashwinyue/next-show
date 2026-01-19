// Package model 定义数据模型.
package model

import "time"

// KnowledgeBaseStatus 知识库状态.
type KnowledgeBaseStatus string

const (
	KnowledgeBaseStatusActive   KnowledgeBaseStatus = "active"
	KnowledgeBaseStatusInactive KnowledgeBaseStatus = "inactive"
)

// KnowledgeBase 知识库.
type KnowledgeBase struct {
	ID              string              `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name            string              `json:"name" gorm:"size:255;not null"`
	Description     string              `json:"description,omitempty" gorm:"type:text"`
	ChunkingConfig  JSONMap             `json:"chunking_config,omitempty" gorm:"type:jsonb"`
	ParserConfig    JSONMap             `json:"parser_config,omitempty" gorm:"type:jsonb"`
	IndexerType     string              `json:"indexer_type,omitempty" gorm:"size:50"`
	IndexerConfig   JSONMap             `json:"indexer_config,omitempty" gorm:"type:jsonb"`
	EmbeddingConfig JSONMap             `json:"embedding_config,omitempty" gorm:"type:jsonb"`
	Status          KnowledgeBaseStatus `json:"status" gorm:"size:20;not null;default:active"`
	Metadata        JSONMap             `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`

	// 关联
	Documents []KnowledgeDocument `json:"documents,omitempty" gorm:"foreignKey:KnowledgeBaseID"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}

// DocumentSourceType 文档来源类型.
type DocumentSourceType string

const (
	DocumentSourceTypeFile DocumentSourceType = "file"
	DocumentSourceTypeURL  DocumentSourceType = "url"
	DocumentSourceTypeText DocumentSourceType = "text"
	DocumentSourceTypeS3   DocumentSourceType = "s3"
)

// DocumentParseStatus 文档解析状态.
type DocumentParseStatus string

const (
	DocumentParseStatusPending DocumentParseStatus = "pending"
	DocumentParseStatusParsed  DocumentParseStatus = "parsed"
	DocumentParseStatusFailed  DocumentParseStatus = "failed"
)

// KnowledgeDocument 知识文档.
type KnowledgeDocument struct {
	ID              string              `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	KnowledgeBaseID string              `json:"knowledge_base_id" gorm:"type:uuid;not null;index"`
	SourceType      DocumentSourceType  `json:"source_type" gorm:"size:20;not null"`
	Title           string              `json:"title,omitempty" gorm:"size:255"`
	SourceURI       string              `json:"source_uri,omitempty" gorm:"type:text"`
	FileHash        string              `json:"file_hash,omitempty" gorm:"size:64;index"`
	ContentText     string              `json:"content_text,omitempty" gorm:"type:text"`
	Metadata        JSONMap             `json:"metadata,omitempty" gorm:"type:jsonb"`
	ParseStatus     DocumentParseStatus `json:"parse_status" gorm:"size:20;not null;default:pending;index"`
	ErrorMessage    string              `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`

	// 关联
	KnowledgeBase *KnowledgeBase   `json:"knowledge_base,omitempty" gorm:"foreignKey:KnowledgeBaseID"`
	Chunks        []KnowledgeChunk `json:"chunks,omitempty" gorm:"foreignKey:DocumentID"`
}

func (KnowledgeDocument) TableName() string {
	return "knowledge_documents"
}

// KnowledgeChunk 知识分块.
type KnowledgeChunk struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	KnowledgeBaseID string    `json:"knowledge_base_id" gorm:"type:uuid;not null;index"`
	DocumentID      string    `json:"document_id" gorm:"type:uuid;not null;index"`
	ChunkIndex      int       `json:"chunk_index" gorm:"not null"`
	Content         string    `json:"content" gorm:"type:text;not null"`
	ContentHash     string    `json:"content_hash,omitempty" gorm:"size:64;index"`
	Metadata        JSONMap   `json:"metadata,omitempty" gorm:"type:jsonb"`
	IsEnabled       bool      `json:"is_enabled" gorm:"default:true;index"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// 关联
	KnowledgeBase *KnowledgeBase     `json:"knowledge_base,omitempty" gorm:"foreignKey:KnowledgeBaseID"`
	Document      *KnowledgeDocument `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	Embedding     *Embedding         `json:"embedding,omitempty" gorm:"foreignKey:ChunkID"`
}

func (KnowledgeChunk) TableName() string {
	return "knowledge_chunks"
}

// Embedding 向量嵌入（pgvector）.
type Embedding struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	KnowledgeBaseID string    `json:"knowledge_base_id" gorm:"type:uuid;not null;index"`
	ChunkID         string    `json:"chunk_id" gorm:"type:uuid;not null;uniqueIndex"`
	Embedding       []float32 `json:"-" gorm:"type:vector(1024);not null"` // pgvector
	EmbeddingDim    int       `json:"embedding_dim" gorm:"not null;default:1024"`
	EmbeddingModel  string    `json:"embedding_model" gorm:"size:128;default:dashscope/embedding-v4"`
	Metadata        JSONMap   `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// 关联
	KnowledgeBase *KnowledgeBase  `json:"knowledge_base,omitempty" gorm:"foreignKey:KnowledgeBaseID"`
	Chunk         *KnowledgeChunk `json:"chunk,omitempty" gorm:"foreignKey:ChunkID"`
}

func (Embedding) TableName() string {
	return "embeddings"
}

// Package model 定义评估相关数据模型.
package model

import (
	"time"

	"gorm.io/gorm"
)

// EvaluationDataset 评估数据集.
type EvaluationDataset struct {
	ID          string        `json:"id" gorm:"primaryKey;size:36"`
	TenantID    uint          `json:"tenant_id" gorm:"not null;index;size:20"`
	Name        string        `json:"name" gorm:"not null;size:200;index"`
	Description string        `json:"description" gorm:"type:text"`
	SourceType  DatasetSource `json:"source_type" gorm:"size:50;default:manual"`

	// Coze Loop 关联（如果使用云服务）
	CozeLoopWorkspaceID     *int64 `json:"coze_loop_workspace_id,omitempty"`
	CozeLoopEvaluationSetID *int64 `json:"coze_loop_evaluation_set_id,omitempty"`

	// 统计
	ItemCount int `json:"item_count" gorm:"default:0"`
	Version   int `json:"version" gorm:"default:1"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (EvaluationDataset) TableName() string {
	return "evaluation_datasets"
}

// DatasetItem 数据集条目.
type DatasetItem struct {
	ID        string `json:"id" gorm:"primaryKey;size:36"`
	DatasetID string `json:"dataset_id" gorm:"not null;index;size:36"`

	// Query 输入
	Query   string `json:"query" gorm:"type:text;not null"`
	QueryID string `json:"query_id,omitempty" gorm:"size:100"`

	// Ground Truth: 检索部分
	RelevantDocIDs   []string `json:"relevant_doc_ids" gorm:"type:text[]"`
	ExpectedDocCount int      `json:"expected_doc_count" gorm:"default:1"`

	// Ground Truth: 生成部分
	ExpectedAnswer   string `json:"expected_answer" gorm:"type:text"`
	ExpectedAnswerID string `json:"expected_answer_id,omitempty" gorm:"size:100"`

	// 元数据
	Metadata JSONMap `json:"metadata" gorm:"type:json"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (DatasetItem) TableName() string {
	return "dataset_items"
}

// EvaluationTask 评估任务.
type EvaluationTask struct {
	ID        string `json:"id" gorm:"primaryKey;size:36"`
	TenantID  uint   `json:"tenant_id" gorm:"not null;index;size:20"`
	DatasetID string `json:"dataset_id" gorm:"not null;index;size:36"`

	// 配置
	AgentID         string `json:"agent_id" gorm:"not null;index;size:36"`
	KnowledgeBaseID string `json:"knowledge_base_id,omitempty" gorm:"size:36"`

	// Coze Loop 关联
	CozeLoopExperimentID *int64 `json:"coze_loop_experiment_id,omitempty"`

	// 任务状态
	Status     EvaluationStatus `json:"status" gorm:"size:50;default:pending;index"`
	Progress   int              `json:"progress" gorm:"default:0"`
	TotalItems int              `json:"total_items" gorm:"default:0"`

	// 错误信息
	ErrorMessage string `json:"error_message,omitempty" gorm:"type:text"`

	// 结果汇总
	AvgRecall    *float64 `json:"avg_recall"`
	AvgPrecision *float64 `json:"avg_precision"`
	AvgMRR       *float64 `json:"avg_mrr"`
	AvgBLEU      *float64 `json:"avg_bleu"`

	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (EvaluationTask) TableName() string {
	return "evaluation_tasks"
}

// EvaluationResult 评估结果（单个测试用例）.
type EvaluationResult struct {
	ID     string `json:"id" gorm:"primaryKey;size:36"`
	TaskID string `json:"task_id" gorm:"not null;index;size:36"`
	ItemID string `json:"item_id" gorm:"not null;index;size:36"`

	// 检索结果
	RetrievedDocIDs  []string `json:"retrieved_doc_ids" gorm:"type:text[]"`
	RetrievalLatency int64    `json:"retrieval_latency_ms"` // 毫秒
	RetrievalOK      bool     `json:"retrieval_ok"`

	// 生成结果
	GeneratedAnswer    string `json:"generated_answer" gorm:"type:text"`
	GenerationLatency  int64  `json:"generation_latency_ms"` // 毫秒
	GenerationOK       bool   `json:"generation_ok"`
	PromptTokens       int    `json:"prompt_tokens"`
	CompletionTokens   int    `json:"completion_tokens"`
	TotalTokens        int    `json:"total_tokens"`

	// 评估指标
	Metrics EvaluationMetrics `json:"metrics" gorm:"embedded;embeddedPrefix:metric_"`

	CreatedAt time.Time `json:"created_at"`
}

func (EvaluationResult) TableName() string {
	return "evaluation_results"
}

// EvaluationMetrics 评估指标.
type EvaluationMetrics struct {
	// 检索指标
	Recall    float64 `json:"recall" gorm:"metric_recall"`
	Precision float64 `json:"precision" gorm:"metric_precision"`
	MRR       float64 `json:"mrr" gorm:"metric_mrr"` // Mean Reciprocal Rank

	// 生成指标
	BLEU  float64      `json:"bleu" gorm:"metric_bleu"`
	ROUGE ROUGEMetrics `json:"rouge" gorm:"embedded;embeddedPrefix:rouge_"`

	// 质量指标
	AnswerRelevance float64 `json:"answer_relevance" gorm:"metric_answer_relevance"`
	ContextCoverage float64 `json:"context_coverage" gorm:"metric_context_coverage"`
}

// ROUGEMetrics ROUGE 指标.
type ROUGEMetrics struct {
	ROUGE1 float64 `json:"rouge1" gorm:"rouge_rouge1"`
	ROUGE2 float64 `json:"rouge2" gorm:"rouge_rouge2"`
	ROUGEL float64 `json:"rougel" gorm:"rouge_rougel"`
}

// DatasetSource 数据集来源类型.
type DatasetSource string

const (
	DatasetSourceManual DatasetSource = "manual" // 手动创建
	DatasetSourceFile   DatasetSource = "file"   // 文件导入
	DatasetSourceTrace  DatasetSource = "trace"  // 从 Trace 导出
)

// EvaluationStatus 评估状态.
type EvaluationStatus string

const (
	EvaluationStatusPending   EvaluationStatus = "pending"
	EvaluationStatusRunning   EvaluationStatus = "running"
	EvaluationStatusCompleted EvaluationStatus = "completed"
	EvaluationStatusFailed    EvaluationStatus = "failed"
)

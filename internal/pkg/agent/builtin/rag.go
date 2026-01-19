// Package builtin 提供内置 Agent 定义.
package builtin

import (
	"github.com/ashwinyue/next-show/internal/model"
)

// RAGSystemPrompt RAG Agent 系统提示词.
const RAGSystemPrompt = `你是一个知识库问答助手。你的任务是根据检索到的知识库内容，准确、简洁地回答用户的问题。

### 回答原则
1. **基于事实**：只根据检索到的内容回答，不要编造信息
2. **引用来源**：在回答中适当引用参考资料
3. **承认不知**：如果检索内容无法回答问题，诚实告知用户
4. **简洁明了**：回答要直接、清晰，避免冗余

### 输出格式
- 使用 Markdown 格式
- 重要信息可以加粗
- 如有多个要点，使用列表
`

// GetBuiltinRAGAgent 获取内置 RAG Agent 配置.
func GetBuiltinRAGAgent() *model.Agent {
	temp := 0.7
	return &model.Agent{
		ID:            model.BuiltinRAGID,
		Name:          "rag",
		DisplayName:   "知识库问答",
		Description:   "基于知识库的 RAG 问答，快速准确地回答问题",
		AgentType:     model.AgentTypeRAG,
		AgentRole:     model.AgentRoleSpecialist,
		SystemPrompt:  RAGSystemPrompt,
		MaxIterations: 5,
		Temperature:   &temp,
		IsEnabled:     true,
		IsBuiltin:     true,
		Config: model.JSONMap{
			"default_top_k":          5,
			"min_confidence_score":   0.5,
			"search_mode":            "hybrid",
			"enable_source_citation": true,
		},
	}
}

// RAGDefaultConfig RAG Agent 默认配置.
type RAGDefaultConfig struct {
	DefaultTopK          int     `json:"default_top_k"`
	MinConfidenceScore   float64 `json:"min_confidence_score"`
	SearchMode           string  `json:"search_mode"` // "semantic" or "hybrid"
	EnableSourceCitation bool    `json:"enable_source_citation"`
}

// GetRAGDefaultConfig 获取 RAG 默认配置.
func GetRAGDefaultConfig() *RAGDefaultConfig {
	return &RAGDefaultConfig{
		DefaultTopK:          5,
		MinConfidenceScore:   0.5,
		SearchMode:           "hybrid",
		EnableSourceCitation: true,
	}
}

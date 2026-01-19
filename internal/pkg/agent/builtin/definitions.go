// Package builtin 提供内置工具和中间件.
package builtin

// 内置工具名称常量.
const (
	ToolThinking            = "thinking"
	ToolTodoWrite           = "todo_write"
	ToolKnowledgeSearch     = "knowledge_search"
	ToolGrepChunks          = "grep_chunks"
	ToolWebSearch           = "web_search"
	ToolWebFetch            = "web_fetch"
	ToolDataAnalysis        = "data_analysis"
	ToolDatabaseQuery       = "database_query"
	ToolListKnowledgeChunks = "list_knowledge_chunks"
)

// ToolDefinition 工具定义.
type ToolDefinition struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Category    string `json:"category"` // knowledge, web, data, utility
}

// AvailableTools 返回所有可用工具定义.
func AvailableTools() []ToolDefinition {
	return []ToolDefinition{
		{Name: ToolThinking, Label: "思考", Description: "动态和反思性的问题解决思考工具", Category: "utility"},
		{Name: ToolTodoWrite, Label: "制定计划", Description: "创建结构化的研究计划", Category: "utility"},
		{Name: ToolKnowledgeSearch, Label: "语义搜索", Description: "理解问题并查找语义相关内容", Category: "knowledge"},
		{Name: ToolGrepChunks, Label: "关键词搜索", Description: "快速定位包含特定关键词的文档", Category: "knowledge"},
		{Name: ToolListKnowledgeChunks, Label: "查看文档分块", Description: "获取文档完整分块内容", Category: "knowledge"},
		{Name: ToolWebSearch, Label: "网络搜索", Description: "搜索互联网获取实时信息", Category: "web"},
		{Name: ToolWebFetch, Label: "网页抓取", Description: "抓取网页内容", Category: "web"},
		{Name: ToolDataAnalysis, Label: "数据分析", Description: "分析数据文件", Category: "data"},
		{Name: ToolDatabaseQuery, Label: "数据库查询", Description: "查询数据库中的信息", Category: "data"},
	}
}

// DefaultTools 返回默认启用的工具列表.
func DefaultTools() []string {
	return []string{
		ToolThinking,
		ToolTodoWrite,
		ToolKnowledgeSearch,
		ToolGrepChunks,
	}
}

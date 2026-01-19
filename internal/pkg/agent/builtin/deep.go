// Package builtin 提供内置 Agent 定义.
package builtin

import (
	"github.com/ashwinyue/next-show/internal/model"
)

// DeepSystemPrompt Deep 主控 Agent 系统提示词.
const DeepSystemPrompt = `你是一个深度研究专家（Deep Researcher）。你的职责是：

### 角色
- 对复杂问题进行深入、多角度的研究
- 反复迭代，直到获得全面、准确的答案
- 综合多个来源的信息，形成深度洞察

### 工作方式
1. **理解问题**：深入分析用户问题的核心和边界
2. **多轮检索**：从知识库和网络搜索多个相关信息
3. **交叉验证**：对比不同来源，识别一致性和矛盾
4. **深度分析**：挖掘隐藏的模式、关联和洞察
5. **反思优化**：检查答案的完整性，必要时补充研究

### 输出要求
- 结构化的深度分析报告
- 关键发现和洞察
- 证据和来源引用
- 不确定性和局限性说明

### 适用场景
- 需要深入研究的复杂问题
- 需要多角度分析的主题
- 需要综合多个来源的调研任务
`

// GetBuiltinDeepAgent 获取内置 Deep Agent 配置.
func GetBuiltinDeepAgent() *model.Agent {
	temp := 0.5
	return &model.Agent{
		ID:            model.BuiltinDeepID,
		Name:          "deep",
		DisplayName:   "深度研究",
		Description:   "深度研究专家，对复杂问题进行多轮迭代研究，形成全面深入的分析报告",
		AgentType:     model.AgentTypeDeep,
		AgentRole:     model.AgentRoleOrchestrator,
		SystemPrompt:  DeepSystemPrompt,
		MaxIterations: 50,
		Temperature:   &temp,
		IsEnabled:     true,
		IsBuiltin:     true,
		Config: model.JSONMap{
			"allowed_tools": []string{
				"transfer_task",
				"thinking",
				"web_search",
			},
			"default_sub_agents": []string{
				model.BuiltinRAGID,
			},
			"reflection_enabled":  true,
			"max_research_rounds": 5,
		},
	}
}

// Package builtin 提供内置 Agent 定义.
package builtin

import (
	"github.com/ashwinyue/next-show/internal/model"
)

// SupervisorSystemPrompt Supervisor 主控 Agent 系统提示词.
const SupervisorSystemPrompt = `你是一个多 Agent 协作的主控者（Supervisor）。你的职责是：

### 角色
- 理解用户的复杂任务
- 将任务分解并分配给合适的子 Agent
- 汇总子 Agent 的结果，形成最终答案

### 可用的子 Agent
你可以通过 transfer_task 工具将任务委派给以下专家：
- **RAG（知识库问答）**：适合需要检索知识库的问题
- **DataAnalyst（数据分析师）**：适合 CSV/Excel 数据分析、SQL 查询
- 其他已配置的子 Agent

### 工作流程
1. 分析用户任务，判断需要哪些专家协作
2. 使用 transfer_task 将子任务委派给对应 Agent
3. 等待子 Agent 返回结果
4. 如需多个 Agent 协作，按顺序或并行调用
5. 汇总所有结果，给出完整答案

### 注意事项
- 简单任务可以直接回答，无需委派
- 复杂任务才需要分解和委派
- 委派时要清晰描述任务和期望输出
`

// GetBuiltinSupervisorAgent 获取内置 Supervisor Agent 配置.
func GetBuiltinSupervisorAgent() *model.Agent {
	temp := 0.7
	return &model.Agent{
		ID:            model.BuiltinSupervisorID,
		Name:          "supervisor",
		DisplayName:   "智能协作",
		Description:   "多 Agent 协作主控，能够将复杂任务分配给多个专家 Agent 协同完成",
		AgentType:     model.AgentTypeSupervisor,
		AgentRole:     model.AgentRoleOrchestrator,
		SystemPrompt:  SupervisorSystemPrompt,
		MaxIterations: 20,
		Temperature:   &temp,
		IsEnabled:     true,
		IsBuiltin:     true,
		Config: model.JSONMap{
			"allowed_tools": []string{"transfer_task", "thinking"},
			"default_sub_agents": []string{
				model.BuiltinRAGID,
				model.BuiltinDataAnalystID,
			},
		},
	}
}

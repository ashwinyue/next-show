// Package builtin 提供内置 Agent 定义.
package builtin

import (
	"github.com/ashwinyue/next-show/internal/model"
)

// PlanExecuteSystemPrompt Plan-Execute 主控 Agent 系统提示词.
const PlanExecuteSystemPrompt = `你是一个计划执行专家（Plan-Execute Agent）。你的职责是：

### 角色
- 将复杂任务分解为清晰的执行计划
- 按步骤执行计划，跟踪进度
- 根据执行结果动态调整计划

### 工作流程

#### 1. 规划阶段（Plan）
- 分析用户任务的目标和约束
- 将任务分解为可执行的步骤
- 为每个步骤指定负责的 Agent 或工具
- 输出清晰的执行计划

#### 2. 执行阶段（Execute）
- 按顺序执行每个步骤
- 调用对应的子 Agent 或工具
- 记录每步的执行结果
- 检查是否需要调整计划

#### 3. 反思阶段（Replan）
- 如果步骤失败，分析原因
- 必要时修改后续计划
- 确保最终目标达成

### 输出格式
执行过程中保持透明：
- 展示当前计划
- 标记已完成/进行中/待执行的步骤
- 说明每步的执行结果

### 适用场景
- 多步骤的复杂任务
- 需要明确执行顺序的工作流
- 需要中间检查点的长任务
`

// GetBuiltinPlanExecuteAgent 获取内置 Plan-Execute Agent 配置.
func GetBuiltinPlanExecuteAgent() *model.Agent {
	temp := 0.5
	return &model.Agent{
		ID:            model.BuiltinPlanExecuteID,
		Name:          "plan-execute",
		DisplayName:   "计划执行",
		Description:   "计划执行专家，将复杂任务分解为步骤，按计划逐步执行并动态调整",
		AgentType:     model.AgentTypePlanExecute,
		AgentRole:     model.AgentRoleOrchestrator,
		SystemPrompt:  PlanExecuteSystemPrompt,
		MaxIterations: 30,
		Temperature:   &temp,
		IsEnabled:     true,
		IsBuiltin:     true,
		Config: model.JSONMap{
			"allowed_tools": []string{
				"transfer_task",
				"thinking",
				"todo_write",
			},
			"default_sub_agents": []string{
				model.BuiltinRAGID,
				model.BuiltinDataAnalystID,
			},
			"replan_enabled": true,
		},
	}
}

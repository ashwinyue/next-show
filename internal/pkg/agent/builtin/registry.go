// Package builtin 提供内置 Agent 定义.
package builtin

import (
	"github.com/ashwinyue/next-show/internal/model"
)

// BuiltinAgentRegistry 内置 Agent 注册表.
var BuiltinAgentRegistry = map[string]func() *model.Agent{
	// 主控 Agent (Orchestrator)
	model.BuiltinSupervisorID:  GetBuiltinSupervisorAgent,
	model.BuiltinDeepID:        GetBuiltinDeepAgent,
	model.BuiltinPlanExecuteID: GetBuiltinPlanExecuteAgent,

	// 子 Agent (Specialist)
	model.BuiltinRAGID:         GetBuiltinRAGAgent,
	model.BuiltinDataAnalystID: GetBuiltinDataAnalystAgent,
}

// GetBuiltinAgent 获取内置 Agent.
func GetBuiltinAgent(id string) *model.Agent {
	if fn, ok := BuiltinAgentRegistry[id]; ok {
		return fn()
	}
	return nil
}

// ListBuiltinAgents 列出所有内置 Agent.
func ListBuiltinAgents() []*model.Agent {
	agents := make([]*model.Agent, 0, len(BuiltinAgentRegistry))
	for _, fn := range BuiltinAgentRegistry {
		agents = append(agents, fn())
	}
	return agents
}

// ListOrchestratorAgents 列出所有主控 Agent.
func ListOrchestratorAgents() []*model.Agent {
	var agents []*model.Agent
	for _, fn := range BuiltinAgentRegistry {
		agent := fn()
		if agent.AgentRole == model.AgentRoleOrchestrator {
			agents = append(agents, agent)
		}
	}
	return agents
}

// ListSpecialistAgents 列出所有子 Agent.
func ListSpecialistAgents() []*model.Agent {
	var agents []*model.Agent
	for _, fn := range BuiltinAgentRegistry {
		agent := fn()
		if agent.AgentRole == model.AgentRoleSpecialist {
			agents = append(agents, agent)
		}
	}
	return agents
}

// IsBuiltinAgent 判断是否为内置 Agent.
func IsBuiltinAgent(id string) bool {
	_, ok := BuiltinAgentRegistry[id]
	return ok
}

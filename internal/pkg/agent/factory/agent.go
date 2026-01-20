// Package factory 提供 Agent 和 Provider 工厂.
package factory

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"

	modelDef "github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/pkg/agent/middleware"
	"github.com/ashwinyue/next-show/internal/pkg/agent/rag"
	agenttools "github.com/ashwinyue/next-show/internal/pkg/agent/tools"
	"github.com/ashwinyue/next-show/internal/store"
)

// AgentFactory Agent 工厂.
type AgentFactory struct {
	store            store.Store
	chatModelFactory *ChatModelFactory
	mcpToolFactory   *MCPToolFactory
	builtinRegistry  *agenttools.ToolRegistry
	knowledgeService agenttools.KnowledgeService
	knowledgeBaseIDs []string // 默认知识库 ID
}

// AgentFactoryConfig Agent 工厂配置.
type AgentFactoryConfig struct {
	Store            store.Store
	KnowledgeService agenttools.KnowledgeService
	KnowledgeBaseIDs []string
}

// NewAgentFactoryWithConfig 使用配置创建 Agent 工厂.
func NewAgentFactoryWithConfig(cfg *AgentFactoryConfig) *AgentFactory {
	registry, _ := agenttools.DefaultRegistry()

	// 注册 web_search 和 web_fetch 工具
	_ = registry.RegisterWebSearchTool(nil)
	_ = registry.RegisterWebFetchTool(nil)

	// 注册 RAG 工具
	if cfg.KnowledgeService != nil {
		_ = registry.RegisterKnowledgeSearchTool(&agenttools.KnowledgeSearchConfig{
			Service:          cfg.KnowledgeService,
			KnowledgeBaseIDs: cfg.KnowledgeBaseIDs,
			TopK:             10,
		})
		_ = registry.RegisterGrepChunksTool(&agenttools.GrepChunksConfig{
			Service:          cfg.KnowledgeService,
			KnowledgeBaseIDs: cfg.KnowledgeBaseIDs,
			TopK:             20,
		})
		_ = registry.RegisterListKnowledgeChunksTool(&agenttools.ListKnowledgeChunksConfig{
			Service: cfg.KnowledgeService,
		})
	}

	return &AgentFactory{
		store:            cfg.Store,
		chatModelFactory: NewChatModelFactory(),
		mcpToolFactory:   NewMCPToolFactory(cfg.Store),
		builtinRegistry:  registry,
		knowledgeService: cfg.KnowledgeService,
		knowledgeBaseIDs: cfg.KnowledgeBaseIDs,
	}
}

// CreateRunner 根据 Agent ID 创建 ADK Runner.
func (f *AgentFactory) CreateRunner(ctx context.Context, agentID string) (*adk.Runner, error) {
	// 获取 Agent 配置（包含 Provider）
	agent, err := f.store.Agents().GetWithProvider(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if agent.Provider == nil {
		// 单独获取 Provider
		provider, err := f.store.Providers().Get(ctx, agent.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider: %w", err)
		}
		agent.Provider = provider
	}

	// 创建 ChatModel
	chatModel, err := f.chatModelFactory.CreateChatModel(ctx, agent.Provider, agent.ModelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	// 根据 Agent 类型创建不同的 Agent
	var adkAgent adk.Agent
	switch agent.AgentType {
	case modelDef.AgentTypeChatModel:
		adkAgent, err = f.createChatModelAgent(ctx, agent, chatModel)
	case modelDef.AgentTypeReact:
		adkAgent, err = f.createChatModelAgent(ctx, agent, chatModel)
	case modelDef.AgentTypeSupervisor:
		adkAgent, err = f.createSupervisorAgent(ctx, agent, chatModel)
	case modelDef.AgentTypeSequential:
		adkAgent, err = f.createSequentialAgent(ctx, agent)
	case modelDef.AgentTypePlanExecute:
		adkAgent, err = f.createPlanExecuteAgent(ctx, agent, chatModel)
	case modelDef.AgentTypeDeep:
		adkAgent, err = f.createDeepAgent(ctx, agent, chatModel)
	case modelDef.AgentTypeDataAnalyst:
		adkAgent, err = f.createDataAnalystAgent(ctx, agent, chatModel)
	case modelDef.AgentTypeLoop:
		adkAgent, err = f.createLoopAgent(ctx, agent, chatModel)
	case modelDef.AgentTypeRAG:
		adkAgent, err = f.createRAGAgent(ctx, agent, chatModel)
	default:
		adkAgent, err = f.createChatModelAgent(ctx, agent, chatModel)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           adkAgent,
		EnableStreaming: true,
	})
	return runner, nil
}

// createChatModelAgent 创建基础 ChatModel Agent.
func (f *AgentFactory) createChatModelAgent(ctx context.Context, agent *modelDef.Agent, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	// 加载工具
	tools, err := f.loadAgentTools(ctx, agent.ID)
	if err != nil {
		return nil, err
	}

	// 加载 MCP 工具
	mcpTools, err := f.mcpToolFactory.GetToolsForAgent(ctx, agent.ID)
	if err == nil && len(mcpTools) > 0 {
		tools = append(tools, mcpTools...)
	}

	// 构建中间件
	middlewares, err := middleware.Build(ctx, middleware.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to build middlewares: %w", err)
	}

	cfg := &adk.ChatModelAgentConfig{
		Name:          agent.Name,
		Description:   agent.Description,
		Instruction:   agent.SystemPrompt,
		Model:         chatModel,
		MaxIterations: agent.MaxIterations,
		Middlewares:   middlewares,
	}

	// 如果有工具则配置 ToolsConfig
	if len(tools) > 0 {
		cfg.ToolsConfig.Tools = tools
	}

	return adk.NewChatModelAgent(ctx, cfg)
}

// loadAgentTools 加载 Agent 关联的工具.
func (f *AgentFactory) loadAgentTools(ctx context.Context, agentID string) ([]tool.BaseTool, error) {
	agentTools, err := f.store.AgentTools().ListEnabledByAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	var tools []tool.BaseTool
	for _, at := range agentTools {
		switch at.ToolType {
		case modelDef.ToolTypeBuiltin:
			t, err := f.builtinRegistry.Get(at.BuiltinToolName)
			if err == nil && t != nil {
				tools = append(tools, t)
			}
		case modelDef.ToolTypeMCP:
			// MCP 工具在外部统一加载
			continue
		case modelDef.ToolTypeCustom:
			// TODO: 加载自定义工具
		}
	}

	return tools, nil
}

// createSupervisorAgent 创建 Supervisor 模式的 Agent.
func (f *AgentFactory) createSupervisorAgent(ctx context.Context, agent *modelDef.Agent, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	// 加载子 Agent 关系
	relations, err := f.store.AgentRelations().ListByParentWithChild(ctx, agent.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent relations: %w", err)
	}

	if len(relations) == 0 {
		return nil, fmt.Errorf("supervisor agent has no sub-agents configured")
	}

	// 创建 Supervisor Agent（协调者）
	supervisorAgent, err := f.createChatModelAgent(ctx, agent, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create supervisor agent: %w", err)
	}

	// 创建所有子 Agent
	var subAgents []adk.Agent
	for _, rel := range relations {
		if rel.ChildAgent == nil {
			continue
		}

		// 为子 Agent 创建 ChatModel
		childChatModel, err := f.chatModelFactory.CreateChatModel(ctx, rel.ChildAgent.Provider, rel.ChildAgent.ModelName)
		if err != nil {
			return nil, fmt.Errorf("failed to create chat model for sub-agent %s: %w", rel.ChildAgent.Name, err)
		}

		// 创建子 Agent
		subAgent, err := f.createChatModelAgent(ctx, rel.ChildAgent, childChatModel)
		if err != nil {
			return nil, fmt.Errorf("failed to create sub-agent %s: %w", rel.ChildAgent.Name, err)
		}
		subAgents = append(subAgents, subAgent)
	}

	// 使用 supervisor.New 组装
	return supervisor.New(ctx, &supervisor.Config{
		Supervisor: supervisorAgent,
		SubAgents:  subAgents,
	})
}

// createSequentialAgent 创建 Sequential 模式的 Agent.
func (f *AgentFactory) createSequentialAgent(ctx context.Context, agent *modelDef.Agent) (adk.Agent, error) {
	// 加载子 Agent 关系（按 sort_order 排序）
	relations, err := f.store.AgentRelations().ListByParentWithChild(ctx, agent.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent relations: %w", err)
	}

	if len(relations) == 0 {
		return nil, fmt.Errorf("sequential agent has no sub-agents configured")
	}

	// 按顺序创建所有子 Agent
	var subAgents []adk.Agent
	for _, rel := range relations {
		if rel.ChildAgent == nil {
			continue
		}

		// 为子 Agent 创建 ChatModel
		childChatModel, err := f.chatModelFactory.CreateChatModel(ctx, rel.ChildAgent.Provider, rel.ChildAgent.ModelName)
		if err != nil {
			return nil, fmt.Errorf("failed to create chat model for sub-agent %s: %w", rel.ChildAgent.Name, err)
		}

		// 创建子 Agent
		subAgent, err := f.createChatModelAgent(ctx, rel.ChildAgent, childChatModel)
		if err != nil {
			return nil, fmt.Errorf("failed to create sub-agent %s: %w", rel.ChildAgent.Name, err)
		}
		subAgents = append(subAgents, subAgent)
	}

	// 使用 SequentialAgent
	return adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
		Name:        agent.Name,
		Description: agent.Description,
		SubAgents:   subAgents,
	})
}

// createDeepAgent 创建 Deep 模式 Agent.
func (f *AgentFactory) createDeepAgent(ctx context.Context, agent *modelDef.Agent, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	// 加载工具
	tools, err := f.loadAgentTools(ctx, agent.ID)
	if err != nil {
		return nil, err
	}

	// 加载子 Agent
	subAgents, err := f.loadSubAgents(ctx, agent.ID)
	if err != nil {
		return nil, err
	}

	maxIter := agent.MaxIterations
	if maxIter <= 0 {
		maxIter = 50
	}

	var toolsConfig adk.ToolsConfig
	if len(tools) > 0 {
		toolsConfig.Tools = tools
	}

	return deep.New(ctx, &deep.Config{
		Name:         agent.Name,
		Description:  agent.Description,
		ChatModel:    chatModel,
		Instruction:  agent.SystemPrompt,
		SubAgents:    subAgents,
		ToolsConfig:  toolsConfig,
		MaxIteration: maxIter,
	})
}

// createPlanExecuteAgent 创建 Plan-Execute 模式 Agent.
func (f *AgentFactory) createPlanExecuteAgent(ctx context.Context, agent *modelDef.Agent, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	// 加载工具
	tools, err := f.loadAgentTools(ctx, agent.ID)
	if err != nil {
		return nil, err
	}

	var toolsConfig adk.ToolsConfig
	if len(tools) > 0 {
		toolsConfig.Tools = tools
	}

	maxIter := agent.MaxIterations
	if maxIter <= 0 {
		maxIter = 30
	}

	// 创建 Planner
	planner, err := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: chatModel,
	})
	if err != nil {
		return nil, fmt.Errorf("create planner: %w", err)
	}

	// 创建 Executor
	executor, err := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model:         chatModel,
		ToolsConfig:   toolsConfig,
		MaxIterations: maxIter,
	})
	if err != nil {
		return nil, fmt.Errorf("create executor: %w", err)
	}

	// 创建 Replanner
	replanner, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: chatModel,
	})
	if err != nil {
		return nil, fmt.Errorf("create replanner: %w", err)
	}

	// 组合 Plan-Execute-Replan
	return planexecute.New(ctx, &planexecute.Config{
		Planner:       planner,
		Executor:      executor,
		Replanner:     replanner,
		MaxIterations: maxIter,
	})
}

// createLoopAgent 创建 Loop 模式 Agent.
// Loop Agent 会循环执行子 Agent 列表，直到达到最大迭代次数或任务完成。
// 适用于需要重复执行的任务场景，如：
// - 数据采集与处理循环
// - 多轮分析与优化
// - 迭代式任务执行
func (f *AgentFactory) createLoopAgent(ctx context.Context, agent *modelDef.Agent, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	// 加载子 Agent 关系（按 sort_order 排序）
	relations, err := f.store.AgentRelations().ListByParentWithChild(ctx, agent.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent relations: %w", err)
	}

	if len(relations) == 0 {
		return nil, fmt.Errorf("loop agent has no sub-agents configured")
	}

	// 按顺序创建所有子 Agent
	var subAgents []adk.Agent
	for _, rel := range relations {
		if rel.ChildAgent == nil {
			continue
		}

		// 为子 Agent 创建 ChatModel
		childChatModel, err := f.chatModelFactory.CreateChatModel(ctx, rel.ChildAgent.Provider, rel.ChildAgent.ModelName)
		if err != nil {
			return nil, fmt.Errorf("failed to create chat model for sub-agent %s: %w", rel.ChildAgent.Name, err)
		}

		// 创建子 Agent
		subAgent, err := f.createChatModelAgent(ctx, rel.ChildAgent, childChatModel)
		if err != nil {
			return nil, fmt.Errorf("failed to create sub-agent %s: %w", rel.ChildAgent.Name, err)
		}
		subAgents = append(subAgents, subAgent)
	}

	// 设置最大迭代次数
	maxIterations := agent.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10 // 默认最多循环 10 次
	}

	// 使用 Loop Agent 模式
	// Loop Agent 会依次循环执行 subAgents 中的每个 Agent
	return adk.NewLoopAgent(ctx, &adk.LoopAgentConfig{
		Name:          agent.Name,
		Description:   agent.Description,
		SubAgents:     subAgents,
		MaxIterations: maxIterations,
	})
}

// createDataAnalystAgent 创建数据分析师 Agent.
func (f *AgentFactory) createDataAnalystAgent(ctx context.Context, agent *modelDef.Agent, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	// 数据分析师使用 ChatModelAgent + data_schema/data_analysis 工具
	// 工具需要在调用时动态注入（因为需要 session 和 document 上下文）
	return f.createChatModelAgent(ctx, agent, chatModel)
}

// loadSubAgents 加载子 Agent.
func (f *AgentFactory) loadSubAgents(ctx context.Context, parentAgentID string) ([]adk.Agent, error) {
	relations, err := f.store.AgentRelations().ListByParentWithChild(ctx, parentAgentID)
	if err != nil {
		return nil, fmt.Errorf("load agent relations: %w", err)
	}

	var subAgents []adk.Agent
	for _, rel := range relations {
		if rel.ChildAgent == nil {
			continue
		}

		// 为子 Agent 创建 ChatModel
		childChatModel, err := f.chatModelFactory.CreateChatModel(ctx, rel.ChildAgent.Provider, rel.ChildAgent.ModelName)
		if err != nil {
			return nil, fmt.Errorf("create chat model for sub-agent %s: %w", rel.ChildAgent.Name, err)
		}

		// 创建子 Agent
		subAgent, err := f.createChatModelAgent(ctx, rel.ChildAgent, childChatModel)
		if err != nil {
			return nil, fmt.Errorf("create sub-agent %s: %w", rel.ChildAgent.Name, err)
		}
		subAgents = append(subAgents, subAgent)
	}

	return subAgents, nil
}

// createRAGAgent 创建 RAG Agent.
func (f *AgentFactory) createRAGAgent(ctx context.Context, agent *modelDef.Agent, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	if f.knowledgeService == nil {
		return nil, fmt.Errorf("knowledge service not configured for RAG agent")
	}

	// 创建 RAG Graph
	ragGraph, err := rag.NewGraph(ctx, &rag.GraphConfig{
		ChatModel:          chatModel,
		Searcher:           rag.NewKnowledgeServiceAdapter(f.knowledgeService),
		DefaultTopK:        5,
		MinConfidenceScore: 0.5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create RAG graph: %w", err)
	}

	// 创建 RAG Agent
	return rag.NewRAGAgent(&rag.RAGAgentConfig{
		Name:             agent.Name,
		Description:      agent.Description,
		Graph:            ragGraph,
		KnowledgeBaseIDs: f.knowledgeBaseIDs,
		TopK:             5,
	}), nil
}

// Close 关闭工厂，清理资源.
func (f *AgentFactory) Close() {
	if f.mcpToolFactory != nil {
		_ = f.mcpToolFactory.Close()
	}
}

// GetBuiltinRegistry 获取内置工具注册表.
func (f *AgentFactory) GetBuiltinRegistry() *agenttools.ToolRegistry {
	return f.builtinRegistry
}

// CreateChatModel 创建 ChatModel.
func (f *AgentFactory) CreateChatModel(ctx context.Context, provider *modelDef.Provider, modelName string) (model.ToolCallingChatModel, error) {
	return f.chatModelFactory.CreateChatModel(ctx, provider, modelName)
}

// GetKnowledgeService 获取知识库服务.
func (f *AgentFactory) GetKnowledgeService() agenttools.KnowledgeService {
	return f.knowledgeService
}

// GetKnowledgeBaseIDs 获取默认知识库 ID.
func (f *AgentFactory) GetKnowledgeBaseIDs() []string {
	return f.knowledgeBaseIDs
}

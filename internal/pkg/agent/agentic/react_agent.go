// Package agentic 提供基于 AgenticModel 的 ReAct Agent 实现。
package agentic

import (
	"context"
	"io"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// AgentConfig ReAct Agent 配置。
type AgentConfig struct {
	Model            model.AgenticModel
	ToolsConfig      compose.ToolsNodeConfig
	MaxStep          int
	ToolReturnDirect map[string]struct{}
}

// Agent ReAct Agent。
type Agent struct {
	runnable compose.Runnable[[]*schema.AgenticMessage, *schema.AgenticMessage]
	graph    *compose.Graph[[]*schema.AgenticMessage, *schema.AgenticMessage]
}

// NewAgent 创建 ReAct Agent。
func NewAgent(ctx context.Context, config *AgentConfig) (_ *Agent, err error) {
	var (
		toolsNode *compose.AgenticToolsNode
		toolInfos []*schema.ToolInfo
	)

	// 生成工具信息
	toolInfos, err = genToolInfos(ctx, config.ToolsConfig)
	if err != nil {
		return nil, err
	}

	// 绑定工具到模型
	agenticModel, err := config.Model.WithTools(toolInfos)
	if err != nil {
		return nil, err
	}

	// 创建 Agentic Tools Node
	toolsNode, err = compose.NewAgenticToolsNode(ctx, &config.ToolsConfig)
	if err != nil {
		return nil, err
	}

	// 构建图
	graph := compose.NewGraph[[]*schema.AgenticMessage, *schema.AgenticMessage]()

	// 添加模型节点
	modelNode := "agentic_model"
	_ = graph.AddAgenticModelNode(modelNode, agenticModel, compose.WithNodeName("AgenticModel"))

	// 添加工具节点
	toolsNodeKey := "tools"
	_ = graph.AddAgenticToolsNode(toolsNodeKey, toolsNode, compose.WithNodeName("Tools"))

	// 添加边
	_ = graph.AddEdge(compose.START, modelNode)

	// 添加分支：模型 -> 工具 或 END
	modelPostBranch := func(ctx context.Context, sr *schema.StreamReader[*schema.AgenticMessage]) (endNode string, err error) {
		defer sr.Close()

		for {
			msg, err := sr.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", err
			}

			// 检查是否有工具调用
			for _, block := range msg.ContentBlocks {
				if block.Type == schema.ContentBlockTypeFunctionToolCall ||
					block.Type == schema.ContentBlockTypeServerToolCall ||
					block.Type == schema.ContentBlockTypeMCPToolCall {
					return toolsNodeKey, nil
				}
			}
		}
		return compose.END, nil
	}

	_ = graph.AddBranch(modelNode, compose.NewStreamGraphBranch(modelPostBranch,
		map[string]bool{
			toolsNodeKey: true,
			compose.END:  true,
		},
	))

	// 工具 -> 模型
	_ = graph.AddEdge(toolsNodeKey, modelNode)

	// 编译
	compileOpts := []compose.GraphCompileOption{
		compose.WithMaxRunSteps(config.MaxStep),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
	}

	runnable, err := graph.Compile(ctx, compileOpts...)
	if err != nil {
		return nil, err
	}

	return &Agent{
		runnable: runnable,
		graph:    graph,
	}, nil
}

// Generate 生成响应。
func (r *Agent) Generate(ctx context.Context, input []*schema.AgenticMessage, opts ...compose.Option) (*schema.AgenticMessage, error) {
	return r.runnable.Invoke(ctx, input, opts...)
}

// Stream 流式生成。
func (r *Agent) Stream(ctx context.Context, input []*schema.AgenticMessage, opts ...compose.Option) (*schema.StreamReader[*schema.AgenticMessage], error) {
	return r.runnable.Stream(ctx, input, opts...)
}

// ExportGraph 导出底层图。
func (r *Agent) ExportGraph() (compose.AnyGraph, []compose.GraphAddNodeOpt) {
	return r.graph, []compose.GraphAddNodeOpt{
		compose.WithGraphCompileOptions(
			compose.WithMaxRunSteps(12),
			compose.WithNodeTriggerMode(compose.AnyPredecessor),
		),
	}
}

func genToolInfos(ctx context.Context, config compose.ToolsNodeConfig) ([]*schema.ToolInfo, error) {
	toolInfos := make([]*schema.ToolInfo, 0, len(config.Tools))
	for _, t := range config.Tools {
		tl, err := t.Info(ctx)
		if err != nil {
			return nil, err
		}
		toolInfos = append(toolInfos, tl)
	}
	return toolInfos, nil
}

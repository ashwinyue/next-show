# RAG Agent 内部 Graph 编排设计文档

## 1. 概述

### 1.1 背景

WeKnora 项目使用插件化的 chat_pipline 实现 RAG 流程，包含查询重写、知识检索、结果重排序等步骤。next-show 项目基于 Eino 框架，使用 Agent 模式实现 AI 应用，但缺少完整的 pipeline 功能。

### 1.2 目标

在 next-show 的 RAG Agent 内部使用 Graph 编排实现类似 WeKnora 的 pipeline 功能，同时保持 Agent 接口的统一性。

### 1.3 设计原则

- **接口一致性**：RAG Agent 对外保持 Agent 接口，与其他 Agent 模式统一
- **内部灵活性**：内部使用 Graph 编排，支持灵活的节点组合
- **性能可控**：固定执行顺序，避免 Agent 的不确定性
- **易于扩展**：通过添加节点扩展功能，支持条件分支

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│              RAG Agent (对外接口)                      │
│                                                       │
│  用户查询 → RAG Agent → 最终答案                         │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│         Internal Graph (内部编排)                      │
│                                                       │
│  START                                                 │
│    │                                                   │
│    ▼                                                   │
│  ┌──────────────┐                                       │
│  │ LoadHistory  │  加载历史消息                          │
│  └──────────────┘                                       │
│    │                                                   │
│    ▼                                                   │
│  ┌──────────────┐                                       │
│  │   Rewrite   │  查询重写（可选）                      │
│  └──────────────┘                                       │
│    │                                                   │
│    ▼                                                   │
│  ┌──────────────┐                                       │
│  │    Search   │  知识检索                             │
│  └──────────────┘                                       │
│    │                                                   │
│    ▼                                                   │
│  ┌──────────────┐                                       │
│  │   Rerank    │  结果重排序（可选）                     │
│  └──────────────┘                                       │
│    │                                                   │
│    ▼                                                   │
│  ┌──────────────┐                                       │
│  │   Generate  │  LLM 生成答案                         │
│  └──────────────┘                                       │
│    │                                                   │
│    ▼                                                   │
│  END                                                  │
└─────────────────────────────────────────────────────────┘
```

### 2.2 分层架构

```
┌─────────────────────────────────────────────────────────┐
│              Handler Layer                            │
│  (HTTP Handler, SSE Writer)                        │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              Biz Layer (Agent Biz)                   │
│  - Agent 管理                                         │
│  - Session 管理                                      │
│  - Agent 创建和调度                                   │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              Agent Layer                              │
│  - RAG Agent (内部使用 Graph)                        │
│  - ReAct Agent (标准 ReAct 模式)                    │
│  - Supervisor Agent (多 Agent 协作)                 │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│         RAG Pipeline (内部 Graph 编排)                 │
│  - LoadHistory Node                                 │
│  - Rewrite Node                                     │
│  - Search Node                                      │
│  - Rerank Node                                     │
│  - Generate Node                                    │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              Service Layer                            │
│  - Knowledge Service (检索服务)                       │
│  - Model Service (LLM 服务)                         │
│  - Reranker Service (重排序服务)                    │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              Store Layer                              │
│  - Session Store                                    │
│  - Message Store                                    │
│  - Knowledge Store                                  │
└─────────────────────────────────────────────────────────┘
```

## 3. 核心模块设计

### 3.1 PipelineState (管道状态)

```go
type PipelineState struct {
    // 输入
    Query           string                 // 原始查询
    History         []*schema.Message       // 历史消息
    
    // 中间状态
    RewrittenQuery  string                 // 重写后的查询
    SearchResults  []*knowledge.ChunkResult // 检索结果
    RerankedResults []*knowledge.ChunkResult // 重排序结果
    
    // 输出
    FinalAnswer    string                 // 最终答案
    
    // 配置
    EnableRewrite  bool                   // 是否启用查询重写
    EnableRerank   bool                   // 是否启用重排序
    TopK           int                    // 检索结果数量
    VectorThreshold float64               // 向量检索阈值
}
```

### 3.2 RAGPipelineConfig (管道配置)

```go
type RAGPipelineConfig struct {
    ModelService      interfaces.ModelService    // LLM 服务
    KnowledgeService knowledge.Service        // 知识检索服务
    RerankerService interfaces.Reranker   // 重排序服务（可选）
    
    // 功能开关
    EnableRewrite    bool                   // 启用查询重写
    EnableRerank     bool                   // 启用结果重排序
    
    // 检索配置
    TopK             int                    // 检索结果数量
    VectorThreshold   float64               // 向量检索阈值
    BM25Threshold    float64               // BM25 检索阈值
    
    // 重排序配置
    RerankTopK      int                    // 重排序后保留数量
    
    // 生成配置
    MaxTokens       int                    // 最大生成 token 数
    Temperature     float64               // 生成温度
}
```

### 3.3 RAGPipeline (管道实现)

```go
type RAGPipeline struct {
    graph    compose.Graph[PipelineState, *schema.AgenticMessage]
    runnable compose.Runnable[PipelineState, *schema.AgenticMessage]
    config   *RAGPipelineConfig
}

// 核心方法
func NewRAGPipeline(ctx context.Context, cfg *RAGPipelineConfig) (*RAGPipeline, error)
func (p *RAGPipeline) Run(ctx context.Context, query string, history []*schema.Message) (*schema.AgenticMessage, error)
func (p *RAGPipeline) Stream(ctx context.Context, query string, history []*schema.Message) (*schema.StreamReader[*schema.AgenticMessage], error)
func (p *RAGPipeline) ExportGraph() (compose.AnyGraph, []compose.GraphAddNodeOpt)
```

## 4. 节点设计

### 4.1 LoadHistory Node

**功能**：加载历史对话消息

**输入**：`PipelineState`
**输出**：`PipelineState`（更新 History）

**实现**：
```go
func loadHistory(ctx context.Context, modelService interfaces.ModelService, state PipelineState) (PipelineState, error) {
    // 从数据库加载历史消息
    // 如果已传入 history，直接使用
    return state, nil
}
```

### 4.2 Rewrite Node

**功能**：根据历史消息重写查询

**输入**：`PipelineState`（Query, History）
**输出**：`PipelineState`（更新 RewrittenQuery）

**条件**：`EnableRewrite == true`

**实现**：
```go
func rewriteQuery(ctx context.Context, modelService interfaces.ModelService, state PipelineState) (PipelineState, error) {
    if !state.EnableRewrite {
        state.RewrittenQuery = state.Query
        return state, nil
    }
    
    // 构建重写提示词
    prompt := buildRewritePrompt(state.Query, state.History)
    
    // 调用 LLM 重写
    response, err := modelService.Chat(ctx, []model.Message{
        {Role: "user", Content: prompt},
    })
    if err != nil {
        return state, err
    }
    
    state.RewrittenQuery = response.Content
    return state, nil
}
```

**提示词模板**：
```
你是一个查询重写专家。根据历史对话，重写用户的查询，使其更清晰、更完整。

历史对话：
{history}

用户查询：{query}

请只输出重写后的查询，不要其他内容。
```

### 4.3 Search Node

**功能**：执行知识检索

**输入**：`PipelineState`（RewrittenQuery 或 Query）
**输出**：`PipelineState`（更新 SearchResults）

**实现**：
```go
func searchKnowledge(ctx context.Context, knowledgeService knowledge.Service, state PipelineState) (PipelineState, error) {
    // 使用重写后的查询（如果有）
    query := state.RewrittenQuery
    if query == "" {
        query = state.Query
    }
    
    // 执行混合检索
    result, err := knowledgeService.HybridSearch(ctx, &knowledge.HybridSearchRequest{
        Query:            query,
        KnowledgeBaseIDs: []string{},
        TopK:             state.TopK,
        VectorWeight:     0.7,
        BM25Weight:       0.3,
    })
    if err != nil {
        return state, err
    }
    
    state.SearchResults = result.Chunks
    return state, nil
}
```

### 4.4 Rerank Node

**功能**：对检索结果进行重排序

**输入**：`PipelineState`（Query, SearchResults）
**输出**：`PipelineState`（更新 RerankedResults）

**条件**：`EnableRerank == true && len(SearchResults) > 3`

**实现**：
```go
func rerankResults(ctx context.Context, modelService interfaces.ModelService, state PipelineState) (PipelineState, error) {
    if !state.EnableRerank || len(state.SearchResults) <= 3 {
        state.RerankedResults = state.SearchResults
        return state, nil
    }
    
    // 调用重排序模型
    reranked, err := rerankWithLLM(ctx, modelService, state.Query, state.SearchResults, state.TopK)
    if err != nil {
        // 重排序失败，使用原结果
        state.RerankedResults = state.SearchResults
        return state, nil
    }
    
    state.RerankedResults = reranked
    return state, nil
}
```

**重排序策略**：
1. **基于 LLM**：让 LLM 根据查询对结果排序
2. **基于 Reranker**：使用专门的 Reranker 模型
3. **混合策略**：结合向量分数和 LLM 评分

### 4.5 Generate Node

**功能**：基于检索结果生成答案

**输入**：`PipelineState`（Query, RerankedResults）
**输出**：`*schema.AgenticMessage`（FinalAnswer）

**实现**：
```go
func generateAnswer(ctx context.Context, modelService interfaces.ModelService, state PipelineState) (*schema.AgenticMessage, error) {
    // 构建生成提示词
    prompt := buildGeneratePrompt(state.Query, state.RerankedResults)
    
    // 调用 LLM 生成答案
    response, err := modelService.Chat(ctx, []model.Message{
        {Role: "user", Content: prompt},
    })
    if err != nil {
        return nil, err
    }
    
    return schema.UserAgenticMessage(response.Content), nil
}
```

**提示词模板**：
```
你是一个知识库问答助手。根据检索到的知识库内容，准确、简洁地回答用户的问题。

用户问题：{query}

参考资料：
{context}

### 回答原则
1. **基于事实**：只根据检索到的内容回答，不要编造信息
2. **引用来源**：在回答中适当引用参考资料
3. **承认不知**：如果检索内容无法回答问题，诚实告知用户
4. **简洁明了**：回答要直接、清晰，避免冗余

请回答：
```

## 5. 接口设计

### 5.1 RAG Agent 接口

```go
type RAGAgent interface {
    // Generate 生成答案
    Generate(ctx context.Context, query string, history []*schema.Message) (*schema.AgenticMessage, error)
    
    // Stream 流式生成
    Stream(ctx context.Context, query string, history []*schema.Message) (*schema.StreamReader[*schema.AgenticMessage], error)
    
    // GetConfig 获取配置
    GetConfig() *model.Agent
    
    // ExportGraph 导出图结构（用于调试和可视化）
    ExportGraph() (compose.AnyGraph, []compose.GraphAddNodeOpt)
}
```

### 5.2 RAGAgentAdapter (适配器)

将 RAG Agent 适配为 Agentic Agent 接口：

```go
type RAGAgentAdapter struct {
    ragAgent RAGAgent
}

func (a *RAGAgentAdapter) Generate(ctx context.Context, input []*schema.AgenticMessage, opts ...compose.Option) (*schema.AgenticMessage, error)
func (a *RAGAgentAdapter) Stream(ctx context.Context, input []*schema.AgenticMessage, opts ...compose.Option) (*schema.StreamReader[*schema.AgenticMessage], error)
```

## 6. 数据流设计

### 6.1 同步执行流程

```
用户查询
    │
    ▼
┌─────────────────────────────────────┐
│  RAGAgent.Generate()            │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  Pipeline.Run()                 │
│                                 │
│  State: {                      │
│    Query: "用户问题"            │
│    History: [...]                │
│  }                             │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  LoadHistory Node               │
│  State: {                      │
│    Query: "用户问题"            │
│    History: [...]                │
│  }                             │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  Rewrite Node                   │
│  State: {                      │
│    Query: "用户问题"            │
│    RewrittenQuery: "重写后的问题" │
│    History: [...]                │
│  }                             │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  Search Node                    │
│  State: {                      │
│    Query: "用户问题"            │
│    RewrittenQuery: "重写后的问题" │
│    SearchResults: [...]          │
│  }                             │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  Rerank Node                   │
│  State: {                      │
│    Query: "用户问题"            │
│    SearchResults: [...]          │
│    RerankedResults: [...]       │
│  }                             │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  Generate Node                  │
│  Output: "最终答案"             │
└─────────────────────────────────────┘
    │
    ▼
最终答案
```

### 6.2 流式执行流程

```
用户查询
    │
    ▼
┌─────────────────────────────────────┐
│  RAGAgent.Stream()             │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  Pipeline.Stream()              │
│                                 │
│  返回 StreamReader              │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  StreamReader.Recv()          │
│  循环接收流式消息                │
└─────────────────────────────────────┘
    │
    ▼
SSE 事件流
```

## 7. 配置说明

### 7.1 Agent 配置

```json
{
  "id": "builtin_rag",
  "name": "rag",
  "display_name": "知识库问答",
  "description": "基于知识库的 RAG 问答，使用内部 Graph 编排实现查询重写、检索、重排序和生成",
  "agent_type": "rag",
  "agent_role": "specialist",
  "max_iterations": 1,
  "temperature": 0.7,
  "is_enabled": true,
  "is_builtin": true,
  "config": {
    "default_top_k": 5,
    "min_confidence_score": 0.5,
    "search_mode": "hybrid",
    "enable_source_citation": true,
    "enable_query_rewrite": true,
    "enable_rerank": true,
    "rerank_top_k": 5,
    "vector_threshold": 0.7,
    "bm25_threshold": 0.3
  }
}
```

### 7.2 配置项说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|---------|------|
| `enable_query_rewrite` | bool | false | 是否启用查询重写 |
| `enable_rerank` | bool | false | 是否启用结果重排序 |
| `default_top_k` | int | 5 | 检索结果数量 |
| `rerank_top_k` | int | 5 | 重排序后保留数量 |
| `vector_threshold` | float64 | 0.7 | 向量检索阈值 |
| `bm25_threshold` | float64 | 0.3 | BM25 检索阈值 |
| `enable_source_citation` | bool | true | 是否在答案中引用来源 |

## 8. 扩展性设计

### 8.1 添加新节点

```go
// 1. 定义节点函数
func customNode(ctx context.Context, state PipelineState) (PipelineState, error) {
    // 自定义逻辑
    return state, nil
}

// 2. 添加到图
graph.AddChainNode("custom_node",
    compose.NewChain(customNode),
    compose.WithNodeName("CustomNode"))

// 3. 添加边
graph.AddEdge("previous_node", "custom_node")
graph.AddEdge("custom_node", "next_node")
```

### 8.2 条件分支

```go
// 添加条件分支
branchFunc := func(ctx context.Context, state PipelineState) (string, error) {
    if state.EnableRewrite {
        return "rewrite_node", nil
    }
    return "search_node", nil
}

graph.AddBranch("start_node", 
    compose.NewStateGraphBranch(branchFunc, map[string]bool{
        "rewrite_node": true,
        "search_node": true,
    }))
```

### 8.3 并行执行

```go
// 添加并行节点
graph.AddPassthroughNode("parallel_start")
graph.AddChainNode("node1", compose.NewChain(func1))
graph.AddChainNode("node2", compose.NewChain(func2))

graph.AddEdge("parallel_start", "node1")
graph.AddEdge("parallel_start", "node2")

graph.AddPassthroughNode("parallel_end")
graph.AddEdge("node1", "parallel_end")
graph.AddEdge("node2", "parallel_end")
```

## 9. 性能优化

### 9.1 并发优化

- **检索并发**：向量检索和 BM25 检索可以并行执行
- **流式生成**：使用 Stream 而非 Generate，减少首字延迟

### 9.2 缓存策略

- **Embedding 缓存**：缓存查询的向量表示
- **检索结果缓存**：缓存相同查询的检索结果
- **历史消息缓存**：缓存会话的历史消息

### 9.3 资源管理

- **连接池**：数据库和 Redis 连接池
- **并发控制**：限制并发检索请求数量
- **超时控制**：每个节点设置合理的超时时间

## 10. 错误处理

### 10.1 节点级错误处理

```go
func searchKnowledge(ctx context.Context, knowledgeService knowledge.Service, state PipelineState) (PipelineState, error) {
    result, err := knowledgeService.HybridSearch(ctx, ...)
    if err != nil {
        // 记录错误日志
        logger.Errorf(ctx, "search failed: %v", err)
        // 返回空结果，继续执行
        state.SearchResults = []*knowledge.ChunkResult{}
        return state, nil
    }
    state.SearchResults = result.Chunks
    return state, nil
}
```

### 10.2 管道级错误处理

```go
func (p *RAGPipeline) Run(ctx context.Context, query string, history []*schema.Message) (*schema.AgenticMessage, error) {
    state := PipelineState{
        Query:   query,
        History: history,
    }
    
    result, err := p.runnable.Invoke(ctx, state)
    if err != nil {
        // 返回错误消息
        return schema.UserAgenticMessage(fmt.Sprintf("抱歉，处理您的请求时出现错误：%s", err.Error())), nil
    }
    
    return result, nil
}
```

## 11. 监控和日志

### 11.1 关键指标

- **节点执行时间**：每个节点的执行耗时
- **检索准确率**：检索结果的相关性
- **重排序效果**：重排序前后的相关性对比
- **生成质量**：答案的完整性和准确性

### 11.2 日志级别

- **DEBUG**：详细的执行流程和中间状态
- **INFO**：关键节点开始和结束
- **WARN**：非关键错误（如重排序失败）
- **ERROR**：关键错误（如检索失败）

## 12. 测试策略

### 12.1 单元测试

每个节点独立测试：
```go
func TestRewriteNode(t *testing.T) {
    ctx := context.Background()
    state := PipelineState{
        Query: "天气",
        History: []*schema.Message{
            {Role: "user", Content: "北京"},
        },
        EnableRewrite: true,
    }
    
    result, err := rewriteQuery(ctx, mockModelService, state)
    assert.NoError(t, err)
    assert.Contains(t, result.RewrittenQuery, "北京")
}
```

### 12.2 集成测试

测试完整流程：
```go
func TestRAGPipeline(t *testing.T) {
    ctx := context.Background()
    pipeline, err := NewRAGPipeline(ctx, testConfig)
    assert.NoError(t, err)
    
    result, err := pipeline.Run(ctx, "天气", nil)
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Content)
}
```

### 12.3 性能测试

测试执行时间：
```go
func BenchmarkRAGPipeline(b *testing.B) {
    ctx := context.Background()
    pipeline, _ := NewRAGPipeline(ctx, testConfig)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        pipeline.Run(ctx, "天气", nil)
    }
}
```

## 13. 迁移计划

### 13.1 阶段一：基础实现

- [ ] 实现 PipelineState 和 RAGPipelineConfig
- [ ] 实现基础节点（LoadHistory, Search, Generate）
- [ ] 实现图编排和编译
- [ ] 实现 RAGAgent 和适配器

### 13.2 阶段二：功能增强

- [ ] 实现 Rewrite Node
- [ ] 实现 Rerank Node
- [ ] 添加配置项和开关
- [ ] 完善错误处理

### 13.3 阶段三：优化和扩展

- [ ] 性能优化（并发、缓存）
- [ ] 监控和日志
- [ ] 单元测试和集成测试
- [ ] 文档完善

## 14. 参考资料

- [WeKnora chat_pipline 实现](../internal/application/service/chat_pipline/)
- [CloudWeGo Eino 文档](https://github.com/cloudwego/eino)
- [RAG 论文](https://arxiv.org/abs/2005.11401)
- [ReAct 论文](https://arxiv.org/abs/2210.03629)

## 15. 附录

### 15.1 术语表

| 术语 | 说明 |
|------|------|
| RAG | Retrieval-Augmented Generation，检索增强生成 |
| ReAct | Reasoning and Acting，推理-行动模式 |
| Graph | Eino 框架中的图结构，用于编排流程 |
| Node | 图中的节点，执行特定功能 |
| Edge | 图中的边，定义节点之间的连接关系 |
| Branch | 条件分支，根据条件决定下一步 |
| SSE | Server-Sent Events，服务器推送事件 |

### 15.2 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.0 | 2025-02-02 | 初始版本 |

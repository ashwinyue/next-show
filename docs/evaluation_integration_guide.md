# 评估功能集成指南

## 📋 功能概述

next-show 现已集成完整的 RAG 评估功能，支持：

### ✅ 已实现的功能

1. **评估数据集管理**
   - 创建/查询/删除评估数据集
   - 支持手动创建、文件导入、Trace 导出三种来源
   - 灵活的 Schema 定义

2. **评估任务执行**
   - 异步任务执行
   - 并发评估多个测试用例
   - 实时进度跟踪
   - 自动指标计算和聚合

3. **评估指标**
   - **检索指标**: Recall、Precision、MRR、F1
   - **生成指标**: BLEU、ROUGE-1/2/L

4. **Eino Callback 集成**
   - 自动收集检索和生成的 Trace 数据
   - 记录延迟、错误、Token 使用等信息
   - 无侵入式集成到现有 RAG Agent

---

## 🚀 快速开始

### 1. 数据库迁移

```bash
# 执行迁移
go run cmd/server/main.go migrate up

# 或使用 golang-migrate
migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up
```

### 2. 注册路由

在 `cmd/server/main.go` 中注册评估路由：

```go
import (
    "github.com/ashwinyue/next-show/internal/handler/http"
    "github.com/ashwinyue/next-show/internal/biz/evaluation"
)

func main() {
    // ... 初始化 db 和其他依赖

    // 创建评估服务和处理器
    evaluationService := evaluation.NewService(db)
    evaluationHandler := http.NewEvaluationHandler(evaluationService)

    // 注册路由
    router := gin.Default()

    // 评估相关路由
    api := router.Group("/api/v1/evaluation")
    {
        // 数据集管理
        api.POST("/datasets", evaluationHandler.CreateDataset)
        api.GET("/datasets", evaluationHandler.ListDatasets)
        api.GET("/datasets/:id", evaluationHandler.GetDataset)
        api.GET("/datasets/:id/items", evaluationHandler.GetDatasetItems)
        api.DELETE("/datasets/:id", evaluationHandler.DeleteDataset)

        // 评估任务
        api.POST("/run", evaluationHandler.RunEvaluation)
        api.GET("/tasks", evaluationHandler.ListTasks)
        api.GET("/tasks/:id", evaluationHandler.GetTask)
        api.GET("/tasks/:id/results", evaluationHandler.GetTaskResults)
        api.DELETE("/tasks/:id", evaluationHandler.DeleteTask)
    }

    // ... 其他路由
}
```

### 3. 集成到 RAG Agent

#### 步骤 1: 修改 RAG Graph

在 `internal/pkg/agent/rag/graph.go` 中：

```go
// Run 执行 RAG 图（支持 Callback）.
func (g *Graph) Run(ctx context.Context, input *RAGInput, callbacks ...callbacks.Handler) (*RAGOutput, error) {
    // 将 Callback 传递给编译后的图
    opts := []compose.InvokeOpt{}
    for _, cb := range callbacks {
        opts = append(opts, compose.WithCallbacks(cb))
    }

    return g.compiled.Invoke(ctx, input, opts...)
}
```

#### 步骤 2: 在评估服务中使用 Callback

在 `internal/biz/evaluation/evaluation_service.go` 中：

```go
import (
    evalcallback "github.com/ashwinyue/next-show/internal/pkg/eino/callbacks/evaluation"
    "github.com/cloudwego/eino/compose"
)

// evaluateItem 评估单个条目.
func (s *Service) evaluateItem(ctx context.Context, task *model.EvaluationTask, item model.DatasetItem) (*model.EvaluationResult, error) {
    // 1. 创建 Callback Handler
    handler := evalcallback.NewEvaluationCallbackHandler()

    // 2. 准备 RAG 输入
    ragInput := &rag.RAGInput{
        Query:            item.Query,
        KnowledgeBaseIDs: []string{task.KnowledgeBaseID},
        TopK:             5,
    }

    // 3. 获取 RAG Agent 并执行（传入 Callback）
    // ragAgent := s.agentFactory.GetRAGAgent(task.AgentID)
    // output, err := ragAgent.Run(ctx, ragInput, handler)

    // TODO: 实际调用 RAG Agent
    // 目前先使用模拟数据
    output := &rag.RAGOutput{
        Answer:  "模拟生成的答案",
        Sources: []*rag.SourceChunk{},
    }

    // 4. 从 Callback 收集数据
    evalData := handler.GetData()

    // 5. 构建评估结果
    result := &model.EvaluationResult{
        ID:              uuid.New().String(),
        TaskID:          task.ID,
        ItemID:          item.ID,
        RetrievedDocIDs:  evalData.RetrievedDocIDs,
        GeneratedAnswer:  evalData.GeneratedAnswer,
        RetrievalOK:     evalData.RetrievalError == nil,
        GenerationOK:    evalData.GenerationError == nil,
    }

    // 6. 计算指标
    metricInput := &metrics.MetricInput{
        RetrievedIDs:  result.RetrievedDocIDs,
        RelevantIDs:   item.RelevantDocIDs,
        GeneratedText: result.GeneratedAnswer,
        ExpectedText:  item.ExpectedAnswer,
    }

    result.Metrics.Recall = metrics.NewRecallMetric().Compute(metricInput)
    result.Metrics.Precision = metrics.NewPrecisionMetric().Compute(metricInput)
    result.Metrics.MRR = metrics.NewMRRMetric().Compute(metricInput)
    result.Metrics.BLEU = metrics.NewBLEUMetric(4).Compute(metricInput)
    result.Metrics.ROUGE.ROUGE1 = metrics.NewROUGEMetric(metrics.ROUGE1).Compute(metricInput)
    result.Metrics.ROUGE.ROUGE2 = metrics.NewROUGEMetric(metrics.ROUGE2).Compute(metricInput)
    result.Metrics.ROUGE.ROUGEL = metrics.NewROUGEMetric(metrics.ROUGEL).Compute(metricInput)

    // 7. 保存延迟信息
    if evalData.RetrievalLatency > 0 {
        result.RetrievalLatency = evalData.RetrievalLatency.Milliseconds()
    }
    if evalData.GenerationLatency > 0 {
        result.GenerationLatency = evalData.GenerationLatency.Milliseconds()
    }

    return result, nil
}
```

---

## 📊 API 使用示例

### 示例 1: 创建评估数据集

```bash
curl -X POST http://localhost:8080/api/v1/evaluation/datasets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "RAG 知识库质量评估",
    "description": "评估 RAG 系统的检索和生成质量",
    "items": [
      {
        "query": "什么是机器学习？",
        "relevant_doc_ids": ["doc_ml_001", "doc_ml_002"],
        "expected_answer": "机器学习是人工智能的一个分支，它使计算机能够从数据中学习并改进，而无需明确编程。"
      },
      {
        "query": "深度学习和机器学习的区别？",
        "relevant_doc_ids": ["doc_dl_001", "doc_ml_001"],
        "expected_answer": "深度学习是机器学习的子集，使用多层神经网络。主要区别在于特征提取方式和数据需求。"
      }
    ]
  }'
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "RAG 知识库质量评估",
    "item_count": 2,
    "created_at": "2025-01-20T10:00:00Z"
  }
}
```

### 示例 2: 运行评估任务

```bash
curl -X POST http://localhost:8080/api/v1/evaluation/run \
  -H "Content-Type: application/json" \
  -d '{
    "dataset_id": "550e8400-e29b-41d4-a716-446655440000",
    "agent_id": "builtin-rag",
    "knowledge_base_id": "kb_001"
  }'
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "total_items": 2,
    "progress": 0
  }
}
```

### 示例 3: 查询评估结果

```bash
# 查询任务状态
curl http://localhost:8080/api/v1/evaluation/tasks/660e8400-e29b-41d4-a716-446655440000

# 查询详细结果（分页）
curl "http://localhost:8080/api/v1/evaluation/tasks/660e8400-e29b-41d4-a716-446655440000/results?page=1&page_size=10"
```

**响应：**
```json
{
  "success": true,
  "data": {
    "results": [
      {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "task_id": "660e8400-e29b-41d4-a716-446655440000",
        "metrics": {
          "recall": 0.8,
          "precision": 0.66,
          "mrr": 0.5,
          "bleu": 0.75,
          "rouge": {
            "rouge1": 0.72,
            "rouge2": 0.65,
            "rougel": 0.68
          }
        }
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 10
  }
}
```

---

## 🔧 高级用法

### 1. 使用 Eino Callback 收集自定义数据

```go
package main

import (
    "context"
    "github.com/ashwinyue/next-show/internal/pkg/eino/callbacks/evaluation"
    "github.com/cloudwego/eino/compose"
)

func main() {
    // 创建 Callback Handler
    handler := evaluation.NewEvaluationCallbackHandler()

    // 在 RAG Graph 中使用
    ctx := context.Background()
    ragInput := &RAGInput{
        Query: "用户问题",
        KnowledgeBaseIDs: []string{"kb_001"},
    }

    // 传入 Callback
    output, err := ragGraph.Run(ctx, ragInput, handler)
    if err != nil {
        panic(err)
    }

    // 获取收集的数据
    data := handler.GetData()

    // 打印结果
    fmt.Printf("检索到的文档 ID: %v\n", data.RetrievedDocIDs)
    fmt.Printf("生成的答案: %s\n", data.GeneratedAnswer)
    fmt.Printf("检索延迟: %v\n", data.RetrievalLatency)
    fmt.Printf("生成延迟: %v\n", data.GenerationLatency)
}
```

### 2. 自定义评估指标

```go
package metrics

// CustomMetric 自定义评估指标.
type CustomMetric struct{}

func (m *CustomMetric) Name() string {
    return "custom_metric"
}

func (m *CustomMetric) Compute(input *MetricInput) float64 {
    // 实现自定义的计算逻辑
    return 0.85
}

func (m *CustomMetric) Validate(input *MetricInput) error {
    return nil
}

// 使用自定义指标
customMetric := &CustomMetric{}
score := customMetric.Compute(&metrics.MetricInput{
    RetrievedIDs:  []string{"doc_1"},
    RelevantIDs:   []string{"doc_1", "doc_2"},
})
```

### 3. 从 Trace 导出评估数据集

```go
// 获取最近的会话 Trace
sessions, _ := sessionService.ListRecentSessions(ctx, 100)

items := make([]model.DatasetItem, 0)
for _, session := range sessions {
    // 获取会话的消息
    messages, _ := messageService.GetBySessionID(ctx, session.ID)

    // 提取 QA 对作为评估用例
    for _, msg := range messages {
        if msg.Role == "user" {
            item := model.DatasetItem{
                Query:   msg.Content,
                Source:  model.DatasetSourceTrace,
                TraceID: session.ID,
            }
            items = append(items, item)
        }
    }
}

// 创建数据集
dataset := &model.EvaluationDataset{
    Name:       "从 Trace 导出的评估集",
    SourceType: model.DatasetSourceTrace,
    Items:      items,
}
```

---

## 📈 性能优化建议

### 1. 并发控制

```go
// 在 evaluation_service.go 中调整并发数
const maxConcurrentEvaluations = 10

semaphore := make(chan struct{}, maxConcurrentEvaluations)

for _, item := range items {
    semaphore <- struct{}{} // 获取信号量

    go func(item model.DatasetItem) {
        defer func() { <-semaphore }() // 释放信号量

        result, err := s.evaluateItem(ctx, task, item)
        // ...
    }(item)
}
```

### 2. 批量插入优化

```go
// 使用 GORM 的 CreateInBatches
results := make([]*model.EvaluationResult, 0, len(items))
// ... 填充 results

if err := s.db.CreateInBatches(results, 100).Error; err != nil {
    return fmt.Errorf("failed to save results: %w", err)
}
```

---

## 🎯 最佳实践

### 1. 数据集设计

- **查询多样性**: 覆盖不同类型的问题（事实型、概念型、比较型等）
- **Ground Truth 质量**: 确保标准答案准确
- **文档标注**: 明确哪些文档是相关的
- **平衡性**: 每个难度级别的用例数量均衡

### 2. 评估频率

- **开发阶段**: 每次重大改动后运行评估
- **测试阶段**: 每日定时评估
- **生产阶段**: 每周评估，跟踪性能趋势

### 3. 指标解读

- **Recall < 0.5**: 检索遗漏严重，需要优化检索策略
- **Precision < 0.5**: 检索噪音多，需要调整 TopK 或 reranker
- **BLEU < 0.3**: 生成质量差，需要优化 Prompt 或模型
- **MRR < 0.3**: 相关文档排名靠后，需要改进排序算法

---

## 🔗 相关资源

- [CloudWeGo Eino 文档](https://www.cloudwego.io/docs/eino/)
- [WeKnora 评估实现](https://github.com/Tencent/WeKnora)
- [coze-loop 评估架构](https://github.com/coze-dev/coze-loop)
- [RAG 评估最佳实践](https://arxiv.org/abs/2309.01431)

---

## 💬 反馈与支持

如有问题或建议，请提交 Issue 或 Pull Request！

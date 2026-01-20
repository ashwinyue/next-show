// Package evaluation_demo 提供评估功能集成示例.
package evaluation_demo

import (
	"context"
	"fmt"
	"strings"

	"github.com/ashwinyue/next-show/internal/biz/evaluation"
	evalmetrics "github.com/ashwinyue/next-show/internal/biz/evaluation/metrics"
	"github.com/ashwinyue/next-show/internal/model"
	evalcallback "github.com/ashwinyue/next-show/internal/pkg/eino/callbacks/evaluation"
	"gorm.io/gorm"
)

// DemoService 评估功能演示服务.
type DemoService struct {
	db                *gorm.DB
	evaluationService *evaluation.Service
}

// NewDemoService 创建演示服务.
func NewDemoService(db *gorm.DB) *DemoService {
	return &DemoService{
		db:                db,
		evaluationService: evaluation.NewService(db),
	}
}

// CreateSampleDataset 创建示例数据集.
func (s *DemoService) CreateSampleDataset(ctx context.Context) (*model.EvaluationDataset, error) {
	req := &evaluation.CreateDatasetRequest{
		TenantID:    1,
		Name:        "RAG 知识库评估示例",
		Description: "用于测试 RAG Agent 检索和生成质量的示例数据集",
		Items: []evaluation.CreateDatasetItemRequest{
			{
				Query:          "什么是机器学习？",
				RelevantDocIDs: []string{"doc_ml_001", "doc_ml_002"},
				ExpectedAnswer: "机器学习（Machine Learning）是人工智能的一个分支，它使计算机能够从数据中学习并改进，而无需明确编程。主要类型包括监督学习、无监督学习和强化学习。",
			},
			{
				Query:          "深度学习和机器学习的区别是什么？",
				RelevantDocIDs: []string{"doc_dl_001", "doc_ml_001"},
				ExpectedAnswer: "深度学习是机器学习的一个子集，使用多层神经网络。主要区别在于：深度学习可以自动提取特征，而传统机器学习需要手动特征工程；深度学习通常需要更多数据和计算资源。",
			},
			{
				Query:          "什么是自然语言处理？",
				RelevantDocIDs: []string{"doc_nlp_001"},
				ExpectedAnswer: "自然语言处理（NLP）是人工智能领域的一个重要分支，专注于使计算机能够理解、解释和生成人类语言。主要任务包括文本分类、情感分析、机器翻译和问答系统等。",
			},
		},
	}

	return s.evaluationService.CreateDataset(ctx, req)
}

// RunEvaluationWithCallback 使用 Callback 运行评估示例.
func (s *DemoService) RunEvaluationWithCallback(ctx context.Context) (*model.EvaluationTask, error) {
	// 1. 创建示例数据集
	dataset, err := s.CreateSampleDataset(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create sample dataset: %w", err)
	}

	fmt.Printf("✅ 创建数据集成功: %s (包含 %d 个测试用例)\n", dataset.Name, dataset.ItemCount)

	// 2. 创建评估任务
	req := &evaluation.RunEvaluationRequest{
		TenantID:        1,
		DatasetID:       dataset.ID,
		AgentID:         "builtin-rag",
		KnowledgeBaseID: "kb_001",
	}

	task, err := s.evaluationService.RunEvaluation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to run evaluation: %w", err)
	}

	fmt.Printf("✅ 评估任务已创建: %s\n", task.ID)
	fmt.Printf("   状态: %s\n", task.Status)
	fmt.Printf("   总测试用例数: %d\n", task.TotalItems)

	return task, nil
}

// EvaluateSingleItem 示例：评估单个条目（使用 Callback）.
func (s *DemoService) EvaluateSingleItem(ctx context.Context) error {
	// 创建测试用例
	item := &model.DatasetItem{
		Query:          "什么是机器学习？",
		RelevantDocIDs: []string{"doc_ml_001", "doc_ml_002"},
		ExpectedAnswer: "机器学习是人工智能的一个分支...",
	}

	// 创建 Callback Handler
	handler := evalcallback.NewEvaluationCallbackHandler()

	// TODO: 在这里调用你的 RAG Agent，并传入 Callback
	// result, err := ragAgent.RunWithCallback(ctx, item.Query, handler)
	// if err != nil {
	//     return fmt.Errorf("failed to run agent: %w", err)
	// }

	// 模拟：假设 RAG Agent 返回了结果
	_ = handler // 避免未使用警告

	// 模拟的检索和生成结果
	mockRetrievedIDs := []string{"doc_ml_001", "doc_ml_003", "doc_ai_001"}
	mockGeneratedAnswer := "机器学习是人工智能的一个分支，它使计算机能够从数据中学习。"

	// 计算指标
	metricInput := &evalmetrics.MetricInput{
		RetrievedIDs:  mockRetrievedIDs,
		RelevantIDs:   item.RelevantDocIDs,
		GeneratedText: mockGeneratedAnswer,
		ExpectedText:  item.ExpectedAnswer,
	}

	// 计算各项指标
	recall := evalmetrics.NewRecallMetric().Compute(metricInput)
	precision := evalmetrics.NewPrecisionMetric().Compute(metricInput)
	mrr := evalmetrics.NewMRRMetric().Compute(metricInput)
	bleu := evalmetrics.NewBLEUMetric(4).Compute(metricInput)

	fmt.Println("\n📊 单个条目评估结果：")
	fmt.Printf("   Query: %s\n", item.Query)
	fmt.Printf("   检索到的文档: %v\n", mockRetrievedIDs)
	fmt.Printf("   相关的文档: %v\n", item.RelevantDocIDs)
	fmt.Printf("   生成的答案: %s\n", mockGeneratedAnswer)
	fmt.Println("\n   评估指标:")
	fmt.Printf("   - Recall (召回率): %.2f%%\n", recall*100)
	fmt.Printf("   - Precision (精确率): %.2f%%\n", precision*100)
	fmt.Printf("   - MRR (平均倒数排名): %.4f\n", mrr)
	fmt.Printf("   - BLEU (翻译质量): %.4f\n", bleu)

	return nil
}

// PrintExampleUsage 打印使用示例.
func (s *DemoService) PrintExampleUsage() {
	fmt.Println(`
🎯 next-show 评估功能使用指南
=====================================

1️⃣  创建评估数据集
   POST /api/v1/evaluation/datasets
   {
     "name": "RAG 评估数据集",
     "description": "测试检索和生成质量",
     "items": [
       {
         "query": "什么是机器学习？",
         "relevant_doc_ids": ["doc_1", "doc_2"],
         "expected_answer": "机器学习是..."
       }
     ]
   }

2️⃣  运行评估任务
   POST /api/v1/evaluation/run
   {
     "dataset_id": "dataset-uuid",
     "agent_id": "builtin-rag",
     "knowledge_base_id": "kb-001"
   }

3️⃣  查询评估结果
   GET /api/v1/evaluation/tasks/{task_id}/results

4️⃣  在代码中使用 Eino Callback 收集数据
   import "github.com/ashwinyue/next-show/internal/pkg/eino/callbacks/evaluation"

   handler := evaluation.NewEvaluationCallbackHandler()

   // 在 RAG Graph 中使用
   graph.Invoke(ctx, input, compose.WithCallbacks(handler))

   // 获取收集的数据
   data := handler.GetData()
   fmt.Printf("检索到的文档: %v\n", data.RetrievedDocIDs)
   fmt.Printf("生成的答案: %s\n", data.GeneratedAnswer)
   fmt.Printf("检索延迟: %v\n", data.RetrievalLatency)
   fmt.Printf("生成延迟: %v\n", data.GenerationLatency)

📊 支持的评估指标
-------------------
检索指标:
  ✅ Recall (召回率)     = |Retrieved ∩ Relevant| / |Relevant|
  ✅ Precision (精确率)  = |Retrieved ∩ Relevant| / |Retrieved|
  ✅ MRR (平均倒数排名)   = 1 / rank_of_first_relevant_doc
  ✅ F1 Score            = 2 * (Precision * Recall) / (Precision + Recall)

生成指标:
  ✅ BLEU  (机器翻译质量)
  ✅ ROUGE-1/2/L (摘要质量)

🔧 下一步：集成到 RAG Agent
---------------------------
在 internal/pkg/agent/rag/graph.go 中：

1. 导入 Callback Handler
   import evalcb "github.com/ashwinyue/next-show/internal/pkg/eino/callbacks/evaluation"

2. 在 Run 方法中接收可选的 Callback
   func (g *Graph) Run(ctx context.Context, input *RAGInput, callbacks ...callbacks.Handler) (*RAGOutput, error)

3. 传递 Callback 到 Graph
   return g.compiled.Invoke(ctx, input, compose.WithCallbacks(callbacks...))

4. 在评估服务中使用
   handler := evaluation.NewEvaluationCallbackHandler()
   output, err := ragGraph.Run(ctx, ragInput, handler)
   data := handler.GetData()

💡 提示
-------
- 评估任务是异步执行的，不会阻塞 API 响应
- 可以通过轮询 GET /api/v1/evaluation/tasks/{id} 查询任务状态
- 所有结果都会持久化到数据库，支持历史查询
- 支持并发评估多个数据集
`)
}

// Demo 运行完整演示.
func (s *DemoService) Demo(ctx context.Context) error {
	fmt.Println("🚀 next-show 评估功能演示")
	fmt.Println("========================\n")

	// 打印使用指南
	s.PrintExampleUsage()

	// 创建示例数据集并运行评估
	fmt.Println("\n开始执行演示...\n")

	task, err := s.RunEvaluationWithCallback(ctx)
	if err != nil {
		return fmt.Errorf("evaluation failed: %w", err)
	}

	fmt.Printf("\n✅ 评估任务已启动！\n")
	fmt.Printf("任务 ID: %s\n", task.ID)
	fmt.Printf("查询状态: GET /api/v1/evaluation/tasks/%s\n", task.ID)

	// 演示单个条目评估
	fmt.Println("\n" + strings.Repeat("=", 60))
	return s.EvaluateSingleItem(ctx)
}

// 辅助函数：重复字符串
func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Package metrics 提供评估指标接口和实现.
package metrics

// MetricInput 指标计算输入.
type MetricInput struct {
	// 检索相关
	RetrievedIDs []string // 检索到的文档 ID 列表
	RelevantIDs  []string // 相关文档 ID 列表 (Ground Truth)
	Ranks        []int    // 检索排名（可选，用于 NDCG）

	// 生成相关
	GeneratedText string // 生成的文本
	ExpectedText  string // 期望文本 (Ground Truth)

	// 额外信息
	Metadata map[string]interface{}
}

// Metric 指标接口.
type Metric interface {
	// Name 返回指标名称
	Name() string

	// Compute 计算指标值 (0.0 - 1.0)
	Compute(input *MetricInput) float64

	// Validate 验证输入数据
	Validate(input *MetricInput) error
}

// RetrievalMetric 检索指标接口.
type RetrievalMetric interface {
	Metric
	IsRetrievalMetric()
}

// GenerationMetric 生成指标接口.
type GenerationMetric interface {
	Metric
	IsGenerationMetric()
}

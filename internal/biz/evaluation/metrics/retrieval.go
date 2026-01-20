// Package metrics 提供检索质量评估指标.
package metrics

import (
	"fmt"
)

// RecallMetric 召回率指标.
// Recall = |Retrieved ∩ Relevant| / |Relevant|
type RecallMetric struct{}

// NewRecallMetric 创建 Recall 指标.
func NewRecallMetric() *RecallMetric {
	return &RecallMetric{}
}

func (m *RecallMetric) Name() string {
	return "recall"
}

func (m *RecallMetric) Compute(input *MetricInput) float64 {
	if len(input.RelevantIDs) == 0 {
		return 0.0
	}

	// 转换为 set 用于高效查找
	relevantSet := toSet(input.RelevantIDs)

	// 计算命中的数量
	hitCount := 0
	for _, retrievedID := range input.RetrievedIDs {
		if _, exists := relevantSet[retrievedID]; exists {
			hitCount++
		}
	}

	return float64(hitCount) / float64(len(input.RelevantIDs))
}

func (m *RecallMetric) Validate(input *MetricInput) error {
	if len(input.RetrievedIDs) == 0 || len(input.RelevantIDs) == 0 {
		return fmt.Errorf("retrieved_ids and relevant_ids are required")
	}
	return nil
}

func (m *RecallMetric) IsRetrievalMetric() {}

// PrecisionMetric 精确率指标.
// Precision = |Retrieved ∩ Relevant| / |Retrieved|
type PrecisionMetric struct{}

// NewPrecisionMetric 创建 Precision 指标.
func NewPrecisionMetric() *PrecisionMetric {
	return &PrecisionMetric{}
}

func (m *PrecisionMetric) Name() string {
	return "precision"
}

func (m *PrecisionMetric) Compute(input *MetricInput) float64 {
	if len(input.RetrievedIDs) == 0 {
		return 0.0
	}

	relevantSet := toSet(input.RelevantIDs)

	// 计算命中的数量
	hitCount := 0
	for _, retrievedID := range input.RetrievedIDs {
		if _, exists := relevantSet[retrievedID]; exists {
			hitCount++
		}
	}

	return float64(hitCount) / float64(len(input.RetrievedIDs))
}

func (m *PrecisionMetric) Validate(input *MetricInput) error {
	if len(input.RetrievedIDs) == 0 || len(input.RelevantIDs) == 0 {
		return fmt.Errorf("retrieved_ids and relevant_ids are required")
	}
	return nil
}

func (m *PrecisionMetric) IsRetrievalMetric() {}

// MRRMetric Mean Reciprocal Rank 指标.
// MRR = 1 / rank_of_first_relevant_doc
type MRRMetric struct{}

// NewMRRMetric 创建 MRR 指标.
func NewMRRMetric() *MRRMetric {
	return &MRRMetric{}
}

func (m *MRRMetric) Name() string {
	return "mrr"
}

func (m *MRRMetric) Compute(input *MetricInput) float64 {
	if len(input.RelevantIDs) == 0 {
		return 0.0
	}

	relevantSet := toSet(input.RelevantIDs)

	// 找到第一个相关文档的排名
	for rank, retrievedID := range input.RetrievedIDs {
		if _, exists := relevantSet[retrievedID]; exists {
			return 1.0 / float64(rank+1)
		}
	}

	return 0.0
}

func (m *MRRMetric) Validate(input *MetricInput) error {
	return nil
}

func (m *MRRMetric) IsRetrievalMetric() {}

// F1Metric F1 分数指标（精确率和召回率的调和平均）.
type F1Metric struct{}

// NewF1Metric 创建 F1 指标.
func NewF1Metric() *F1Metric {
	return &F1Metric{}
}

func (m *F1Metric) Name() string {
	return "f1"
}

func (m *F1Metric) Compute(input *MetricInput) float64 {
	recallMetric := NewRecallMetric()
	precisionMetric := NewPrecisionMetric()

	recall := recallMetric.Compute(input)
	precision := precisionMetric.Compute(input)

	if recall+precision == 0 {
		return 0.0
	}

	return 2 * (recall * precision) / (recall + precision)
}

func (m *F1Metric) Validate(input *MetricInput) error {
	if len(input.RetrievedIDs) == 0 || len(input.RelevantIDs) == 0 {
		return fmt.Errorf("retrieved_ids and relevant_ids are required")
	}
	return nil
}

func (m *F1Metric) IsRetrievalMetric() {}

// toSet 将字符串切片转换为 set.
func toSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

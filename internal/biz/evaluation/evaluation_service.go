// Package evaluation 提供评估业务逻辑.
package evaluation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ashwinyue/next-show/internal/biz/evaluation/metrics"
	"github.com/ashwinyue/next-show/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 评估服务.
type Service struct {
	db *gorm.DB
}

// NewService 创建评估服务.
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateDatasetRequest 创建数据集请求.
type CreateDatasetRequest struct {
	TenantID    uint                       `json:"tenant_id"`
	Name        string                     `json:"name" binding:"required"`
	Description string                     `json:"description"`
	Items       []CreateDatasetItemRequest `json:"items"`
}

// CreateDatasetItemRequest 创建数据集条目请求.
type CreateDatasetItemRequest struct {
	Query          string        `json:"query" binding:"required"`
	RelevantDocIDs []string      `json:"relevant_doc_ids"`
	ExpectedAnswer string        `json:"expected_answer"`
	Metadata       model.JSONMap `json:"metadata"`
}

// CreateDataset 创建评估数据集.
func (s *Service) CreateDataset(ctx context.Context, req *CreateDatasetRequest) (*model.EvaluationDataset, error) {
	dataset := &model.EvaluationDataset{
		ID:          uuid.New().String(),
		TenantID:    req.TenantID,
		Name:        req.Name,
		Description: req.Description,
		SourceType:  model.DatasetSourceManual,
		ItemCount:   len(req.Items),
		Version:     1,
	}

	// 使用事务创建数据集和条目
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(dataset).Error; err != nil {
			return fmt.Errorf("failed to create dataset: %w", err)
		}

		// 创建条目
		for _, itemReq := range req.Items {
			item := &model.DatasetItem{
				ID:             uuid.New().String(),
				DatasetID:      dataset.ID,
				Query:          itemReq.Query,
				RelevantDocIDs: itemReq.RelevantDocIDs,
				ExpectedAnswer: itemReq.ExpectedAnswer,
				Metadata:       itemReq.Metadata,
			}
			if err := tx.Create(item).Error; err != nil {
				return fmt.Errorf("failed to create dataset item: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return dataset, nil
}

// GetDataset 获取数据集详情.
func (s *Service) GetDataset(ctx context.Context, tenantID uint, datasetID string) (*model.EvaluationDataset, error) {
	var dataset model.EvaluationDataset
	err := s.db.Where("tenant_id = ? AND id = ?", tenantID, datasetID).First(&dataset).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	return &dataset, nil
}

// ListDatasets 列出数据集.
func (s *Service) ListDatasets(ctx context.Context, tenantID uint) ([]model.EvaluationDataset, error) {
	var datasets []model.EvaluationDataset
	err := s.db.Where("tenant_id = ?", tenantID).Find(&datasets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}

	return datasets, nil
}

// GetDatasetItems 获取数据集的条目.
func (s *Service) GetDatasetItems(ctx context.Context, tenantID uint, datasetID string) ([]model.DatasetItem, error) {
	// 先验证权限
	var dataset model.EvaluationDataset
	err := s.db.Where("tenant_id = ? AND id = ?", tenantID, datasetID).First(&dataset).Error
	if err != nil {
		return nil, fmt.Errorf("dataset not found: %w", err)
	}

	var items []model.DatasetItem
	err = s.db.Where("dataset_id = ?", datasetID).Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset items: %w", err)
	}

	return items, nil
}

// RunEvaluationRequest 运行评估请求.
type RunEvaluationRequest struct {
	TenantID        uint   `json:"tenant_id"`
	DatasetID       string `json:"dataset_id" binding:"required"`
	AgentID         string `json:"agent_id" binding:"required"`
	KnowledgeBaseID string `json:"knowledge_base_id"`
}

// RunEvaluation 运行评估任务（异步）.
func (s *Service) RunEvaluation(ctx context.Context, req *RunEvaluationRequest) (*model.EvaluationTask, error) {
	// 1. 验证数据集存在
	_, err := s.GetDataset(ctx, req.TenantID, req.DatasetID)
	if err != nil {
		return nil, fmt.Errorf("dataset not found: %w", err)
	}

	// 2. 获取数据集条目
	items, err := s.GetDatasetItems(ctx, req.TenantID, req.DatasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset items: %w", err)
	}

	// 3. 创建评估任务
	now := time.Now()
	task := &model.EvaluationTask{
		ID:              uuid.New().String(),
		TenantID:        req.TenantID,
		DatasetID:       req.DatasetID,
		AgentID:         req.AgentID,
		KnowledgeBaseID: req.KnowledgeBaseID,
		Status:          model.EvaluationStatusPending,
		TotalItems:      len(items),
		StartedAt:       &now,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 4. 异步执行评估
	go s.executeEvaluation(context.Background(), task, items)

	return task, nil
}

// executeEvaluation 执行评估任务.
func (s *Service) executeEvaluation(ctx context.Context, task *model.EvaluationTask, items []model.DatasetItem) {
	// 更新任务状态为运行中
	task.Status = model.EvaluationStatusRunning
	s.db.Save(task)

	var wg sync.WaitGroup
	resultsChan := make(chan *model.EvaluationResult, len(items))
	errorsChan := make(chan error, len(items))

	// 并发执行评估（可配置并发数）
	for _, item := range items {
		wg.Add(1)
		go func(item model.DatasetItem) {
			defer wg.Done()

			result, err := s.evaluateItem(ctx, task, item)
			if err != nil {
				errorsChan <- err
				return
			}
			resultsChan <- result
		}(item)
	}

	// 等待所有评估完成
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// 收集结果
	var results []*model.EvaluationResult
	var errorCount int

	for result := range resultsChan {
		results = append(results, result)

		// 保存结果到数据库
		if err := s.db.Create(result).Error; err != nil {
			errorCount++
			continue
		}

		// 更新进度
		task.Progress = int(float64(len(results)) / float64(task.TotalItems) * 100)
		s.db.Save(task)
	}

	// 处理错误
	for err := range errorsChan {
		errorCount++
		_ = err // 记录错误日志
	}

	// 计算平均指标并更新任务
	s.aggregateResults(task, results)

	task.Status = model.EvaluationStatusCompleted
	if errorCount > 0 {
		task.Status = model.EvaluationStatusFailed
		task.ErrorMessage = fmt.Sprintf("%d items failed", errorCount)
	}
	now := time.Now()
	task.CompletedAt = &now
	s.db.Save(task)
}

// evaluateItem 评估单个条目.
func (s *Service) evaluateItem(ctx context.Context, task *model.EvaluationTask, item model.DatasetItem) (*model.EvaluationResult, error) {
	// TODO: 这里需要调用 RAG Agent 并传入 Callback Handler
	// 目前先创建一个模拟结果

	result := &model.EvaluationResult{
		ID:     uuid.New().String(),
		TaskID: task.ID,
		ItemID: item.ID,
		// RetrievedDocIDs:  <从 RAG Agent 获取>
		// GeneratedAnswer:  <从 RAG Agent 获取>
		RetrievalOK:  true,
		GenerationOK: true,
	}

	// 计算指标
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

	return result, nil
}

// aggregateResults 聚合评估结果.
func (s *Service) aggregateResults(task *model.EvaluationTask, results []*model.EvaluationResult) {
	if len(results) == 0 {
		return
	}

	var totalRecall, totalPrecision, totalMRR, totalBLEU float64

	for _, result := range results {
		totalRecall += result.Metrics.Recall
		totalPrecision += result.Metrics.Precision
		totalMRR += result.Metrics.MRR
		totalBLEU += result.Metrics.BLEU
	}

	count := float64(len(results))
	task.AvgRecall = &[]float64{totalRecall / count}[0]
	task.AvgPrecision = &[]float64{totalPrecision / count}[0]
	task.AvgMRR = &[]float64{totalMRR / count}[0]
	task.AvgBLEU = &[]float64{totalBLEU / count}[0]
}

// GetTask 获取评估任务.
func (s *Service) GetTask(ctx context.Context, tenantID uint, taskID string) (*model.EvaluationTask, error) {
	var task model.EvaluationTask
	err := s.db.Where("tenant_id = ? AND id = ?", tenantID, taskID).First(&task).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// ListTasks 列出评估任务.
func (s *Service) ListTasks(ctx context.Context, tenantID uint, datasetID string) ([]model.EvaluationTask, error) {
	var tasks []model.EvaluationTask
	query := s.db.Where("tenant_id = ?", tenantID)

	if datasetID != "" {
		query = query.Where("dataset_id = ?", datasetID)
	}

	err := query.Order("created_at DESC").Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, nil
}

// GetTaskResults 获取评估任务的结果.
func (s *Service) GetTaskResults(ctx context.Context, tenantID uint, taskID string) ([]model.EvaluationResult, error) {
	// 验证任务存在
	var task model.EvaluationTask
	err := s.db.Where("tenant_id = ? AND id = ?", tenantID, taskID).First(&task).Error
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	var results []model.EvaluationResult
	err = s.db.Where("task_id = ?", taskID).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get task results: %w", err)
	}

	return results, nil
}

// DeleteDataset 删除数据集.
func (s *Service) DeleteDataset(ctx context.Context, tenantID uint, datasetID string) error {
	// 使用事务删除数据集及其关联数据
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除数据集条目
		if err := tx.Where("dataset_id = ?", datasetID).Delete(&model.DatasetItem{}).Error; err != nil {
			return err
		}

		// 删除数据集
		if err := tx.Where("tenant_id = ? AND id = ?", tenantID, datasetID).Delete(&model.EvaluationDataset{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteTask 删除评估任务.
func (s *Service) DeleteTask(ctx context.Context, tenantID uint, taskID string) error {
	// 使用事务删除任务及其关联数据
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除评估结果
		if err := tx.Where("task_id = ?", taskID).Delete(&model.EvaluationResult{}).Error; err != nil {
			return err
		}

		// 删除任务
		if err := tx.Where("tenant_id = ? AND id = ?", tenantID, taskID).Delete(&model.EvaluationTask{}).Error; err != nil {
			return err
		}

		return nil
	})
}

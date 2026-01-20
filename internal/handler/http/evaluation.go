// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"
	"strconv"

	"github.com/ashwinyue/next-show/internal/biz/evaluation"
	"github.com/ashwinyue/next-show/internal/model"
	"github.com/gin-gonic/gin"
)

// EvaluationHandler 评估处理器.
type EvaluationHandler struct {
	evaluationService *evaluation.Service
}

// NewEvaluationHandler 创建评估处理器.
func NewEvaluationHandler(evaluationService *evaluation.Service) *EvaluationHandler {
	return &EvaluationHandler{
		evaluationService: evaluationService,
	}
}

// CreateDatasetRequest 创建数据集请求.
type CreateDatasetRequest struct {
	Name        string                         `json:"name" binding:"required"`
	Description string                         `json:"description"`
	Items       []CreateDatasetItemRequestHTTP `json:"items" binding:"required"`
}

// CreateDatasetItemRequestHTTP 创建数据集条目请求（HTTP 层）.
type CreateDatasetItemRequestHTTP struct {
	Query          string        `json:"query" binding:"required"`
	RelevantDocIDs []string      `json:"relevant_doc_ids"`
	ExpectedAnswer string        `json:"expected_answer"`
	Metadata       model.JSONMap `json:"metadata"`
}

// CreateDataset 创建评估数据集.
// @Summary 创建评估数据集
// @Description 创建新的评估数据集，包含多个测试用例
// @Tags 评估
// @Accept json
// @Produce json
// @Param request body CreateDatasetRequest true "创建数据集请求"
// @Success 200 {object} map[string]interface{} "数据集"
// @Failure 400 {object} map[string]string "错误信息"
// @Router /api/v1/evaluation/datasets [post]
func (h *EvaluationHandler) CreateDataset(c *gin.Context) {
	var req CreateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取租户 ID（从中间件或 JWT 中获取）
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		// TODO: 从 JWT 或其他方式获取
		tenantID = uint(1) // 临时默认值
	}

	// 转换为服务层请求
	items := make([]evaluation.CreateDatasetItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = evaluation.CreateDatasetItemRequest{
			Query:          item.Query,
			RelevantDocIDs: item.RelevantDocIDs,
			ExpectedAnswer: item.ExpectedAnswer,
			Metadata:       item.Metadata,
		}
	}

	serviceReq := &evaluation.CreateDatasetRequest{
		TenantID:    tenantID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Items:       items,
	}

	dataset, err := h.evaluationService.CreateDataset(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataset,
	})
}

// GetDataset 获取数据集详情.
// @Summary 获取数据集详情
// @Description 根据 ID 获取评估数据集的详细信息
// @Tags 评估
// @Accept json
// @Produce json
// @Param id path string true "数据集 ID"
// @Success 200 {object} map[string]interface{} "数据集详情"
// @Failure 404 {object} map[string]string "错误信息"
// @Router /api/v1/evaluation/datasets/{id} [get]
func (h *EvaluationHandler) GetDataset(c *gin.Context) {
	id := c.Param("id")

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		tenantID = uint(1)
	}

	dataset, err := h.evaluationService.GetDataset(c.Request.Context(), tenantID.(uint), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataset,
	})
}

// ListDatasets 列出数据集.
// @Summary 列出数据集
// @Description 获取当前租户的所有评估数据集
// @Tags 评估
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "数据集列表"
// @Router /api/v1/evaluation/datasets [get]
func (h *EvaluationHandler) ListDatasets(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		tenantID = uint(1)
	}

	datasets, err := h.evaluationService.ListDatasets(c.Request.Context(), tenantID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    datasets,
	})
}

// GetDatasetItems 获取数据集条目.
// @Summary 获取数据集条目
// @Description 获取指定数据集的所有测试用例
// @Tags 评估
// @Accept json
// @Produce json
// @Param id path string true "数据集 ID"
// @Success 200 {object} map[string]interface{} "数据集条目列表"
// @Router /api/v1/evaluation/datasets/{id}/items [get]
func (h *EvaluationHandler) GetDatasetItems(c *gin.Context) {
	id := c.Param("id")

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		tenantID = uint(1)
	}

	items, err := h.evaluationService.GetDatasetItems(c.Request.Context(), tenantID.(uint), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})
}

// DeleteDataset 删除数据集.
// @Summary 删除数据集
// @Description 删除指定的评估数据集及其所有测试用例
// @Tags 评估
// @Accept json
// @Produce json
// @Param id path string true "数据集 ID"
// @Success 200 {object} map[string]string "成功信息"
// @Router /api/v1/evaluation/datasets/{id} [delete]
func (h *EvaluationHandler) DeleteDataset(c *gin.Context) {
	id := c.Param("id")

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		tenantID = uint(1)
	}

	if err := h.evaluationService.DeleteDataset(c.Request.Context(), tenantID.(uint), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "dataset deleted",
	})
}

// RunEvaluationRequest 运行评估请求.
type RunEvaluationRequest struct {
	DatasetID       string `json:"dataset_id" binding:"required"`
	AgentID         string `json:"agent_id" binding:"required"`
	KnowledgeBaseID string `json:"knowledge_base_id"`
}

// RunEvaluation 运行评估任务.
// @Summary 运行评估任务
// @Description 使用指定 Agent 对数据集进行评估
// @Tags 评估
// @Accept json
// @Produce json
// @Param request body RunEvaluationRequest true "运行评估请求"
// @Success 200 {object} map[string]interface{} "评估任务"
// @Failure 400 {object} map[string]string "错误信息"
// @Router /api/v1/evaluation/run [post]
func (h *EvaluationHandler) RunEvaluation(c *gin.Context) {
	var req RunEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		tenantID = uint(1)
	}

	serviceReq := &evaluation.RunEvaluationRequest{
		TenantID:        tenantID.(uint),
		DatasetID:       req.DatasetID,
		AgentID:         req.AgentID,
		KnowledgeBaseID: req.KnowledgeBaseID,
	}

	task, err := h.evaluationService.RunEvaluation(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

// GetTask 获取评估任务.
// @Summary 获取评估任务
// @Description 根据 ID 获取评估任务的详细信息
// @Tags 评估
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} map[string]interface{} "任务详情"
// @Failure 404 {object} map[string]string "错误信息"
// @Router /api/v1/evaluation/tasks/{id} [get]
func (h *EvaluationHandler) GetTask(c *gin.Context) {
	id := c.Param("id")

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		tenantID = uint(1)
	}

	task, err := h.evaluationService.GetTask(c.Request.Context(), tenantID.(uint), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

// ListTasks 列出评估任务.
// @Summary 列出评估任务
// @Description 获取评估任务列表
// @Tags 评估
// @Accept json
// @Produce json
// @Param dataset_id query string false "数据集 ID（可选）"
// @Success 200 {object} map[string]interface{} "任务列表"
// @Router /api/v1/evaluation/tasks [get]
func (h *EvaluationHandler) ListTasks(c *gin.Context) {
	datasetID := c.Query("dataset_id")

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		tenantID = uint(1)
	}

	tasks, err := h.evaluationService.ListTasks(c.Request.Context(), tenantID.(uint), datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tasks,
	})
}

// GetTaskResults 获取评估任务的结果.
// @Summary 获取评估结果
// @Description 获取指定评估任务的所有结果
// @Tags 评估
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "评估结果"
// @Failure 404 {object} map[string]string "错误信息"
// @Router /api/v1/evaluation/tasks/{id}/results [get]
func (h *EvaluationHandler) GetTaskResults(c *gin.Context) {
	id := c.Param("id")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		tenantID = uint(1)
	}

	results, err := h.evaluationService.GetTaskResults(c.Request.Context(), tenantID.(uint), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 简单分页
	total := len(results)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		results = []model.EvaluationResult{}
	} else if end > total {
		results = results[start:]
	} else {
		results = results[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"results":     results,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// DeleteTask 删除评估任务.
// @Summary 删除评估任务
// @Description 删除指定的评估任务及其所有结果
// @Tags 评估
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} map[string]string "成功信息"
// @Router /api/v1/evaluation/tasks/{id} [delete]
func (h *EvaluationHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		tenantID = uint(1)
	}

	if err := h.evaluationService.DeleteTask(c.Request.Context(), tenantID.(uint), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "task deleted",
	})
}

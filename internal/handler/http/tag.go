// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/model"
)

// ListTags 列出知识库的标签.
func (h *Handler) ListTags(c *gin.Context) {
	kbID := c.Param("kb_id")
	tags, err := h.biz.Knowledge().ListTags(c.Request.Context(), kbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// GetTag 获取标签详情.
func (h *Handler) GetTag(c *gin.Context) {
	tagID := c.Param("tag_id")
	tag, err := h.biz.Knowledge().GetTag(c.Request.Context(), tagID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tag)
}

// CreateTagRequest 创建标签请求.
type CreateTagRequest struct {
	Name        string `json:"name" binding:"required"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// CreateTag 创建标签.
func (h *Handler) CreateTag(c *gin.Context) {
	kbID := c.Param("kb_id")
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag := &model.KnowledgeTag{
		KnowledgeBaseID: kbID,
		Name:            req.Name,
		Color:           req.Color,
		Description:     req.Description,
	}

	if err := h.biz.Knowledge().CreateTag(c.Request.Context(), tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// UpdateTagRequest 更新标签请求.
type UpdateTagRequest struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
}

// UpdateTag 更新标签.
func (h *Handler) UpdateTag(c *gin.Context) {
	tagID := c.Param("tag_id")
	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.biz.Knowledge().GetTag(c.Request.Context(), tagID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		tag.Name = *req.Name
	}
	if req.Color != nil {
		tag.Color = *req.Color
	}
	if req.Description != nil {
		tag.Description = *req.Description
	}

	if err := h.biz.Knowledge().UpdateTag(c.Request.Context(), tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// DeleteTag 删除标签.
func (h *Handler) DeleteTag(c *gin.Context) {
	tagID := c.Param("tag_id")
	if err := h.biz.Knowledge().DeleteTag(c.Request.Context(), tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListChunksByTag 列出标签关联的分块.
func (h *Handler) ListChunksByTag(c *gin.Context) {
	tagID := c.Param("tag_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	chunks, total, err := h.biz.Knowledge().ListChunksByTag(c.Request.Context(), tagID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chunks":    chunks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// === Chunk 管理 Handler ===

// ListChunksByKB 列出知识库的分块.
func (h *Handler) ListChunksByKB(c *gin.Context) {
	kbID := c.Param("kb_id")
	docID := c.Query("document_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var chunks []*model.KnowledgeChunk
	var total int64
	var err error

	if docID != "" {
		chunks, total, err = h.biz.Knowledge().ListChunks(c.Request.Context(), docID, pageSize, offset)
	} else {
		chunks, total, err = h.biz.Knowledge().ListChunksByKnowledgeBase(c.Request.Context(), kbID, pageSize, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chunks":    chunks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetChunkDetail 获取分块详情.
func (h *Handler) GetChunkDetail(c *gin.Context) {
	chunkID := c.Param("chunk_id")
	chunk, err := h.biz.Knowledge().GetChunk(c.Request.Context(), chunkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	tags, _ := h.biz.Knowledge().ListTagsByChunk(c.Request.Context(), chunkID)

	c.JSON(http.StatusOK, gin.H{
		"chunk": chunk,
		"tags":  tags,
	})
}

// UpdateChunkRequest 更新分块请求.
type UpdateChunkRequest struct {
	Content   *string       `json:"content"`
	Metadata  model.JSONMap `json:"metadata"`
	IsEnabled *bool         `json:"is_enabled"`
}

// UpdateChunkHandler 更新分块.
func (h *Handler) UpdateChunkHandler(c *gin.Context) {
	chunkID := c.Param("chunk_id")
	var req UpdateChunkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chunk, err := h.biz.Knowledge().GetChunk(c.Request.Context(), chunkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.Content != nil {
		chunk.Content = *req.Content
	}
	if req.Metadata != nil {
		chunk.Metadata = req.Metadata
	}
	if req.IsEnabled != nil {
		chunk.IsEnabled = *req.IsEnabled
	}

	if err := h.biz.Knowledge().UpdateChunk(c.Request.Context(), chunk); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chunk)
}

// DeleteChunkHandler 删除分块.
func (h *Handler) DeleteChunkHandler(c *gin.Context) {
	chunkID := c.Param("chunk_id")
	if err := h.biz.Knowledge().DeleteChunk(c.Request.Context(), chunkID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// AddChunkTagRequest 添加标签到分块请求.
type AddChunkTagRequest struct {
	TagID string `json:"tag_id" binding:"required"`
}

// AddChunkTag 添加标签到分块.
func (h *Handler) AddChunkTag(c *gin.Context) {
	chunkID := c.Param("chunk_id")
	var req AddChunkTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.biz.Knowledge().AddTagToChunk(c.Request.Context(), chunkID, req.TagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "added"})
}

// RemoveChunkTag 从分块移除标签.
func (h *Handler) RemoveChunkTag(c *gin.Context) {
	chunkID := c.Param("chunk_id")
	tagID := c.Param("tag_id")

	if err := h.biz.Knowledge().RemoveTagFromChunk(c.Request.Context(), chunkID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}

// ListChunkTags 列出分块的标签.
func (h *Handler) ListChunkTagsHandler(c *gin.Context) {
	chunkID := c.Param("chunk_id")
	tags, err := h.biz.Knowledge().ListTagsByChunk(c.Request.Context(), chunkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

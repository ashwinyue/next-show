// Package http 提供 HTTP Handler.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/model"
)

// CreateKnowledgeBase 创建知识库.
func (h *Handler) CreateKnowledgeBase(c *gin.Context) {
	var req model.KnowledgeBase
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.biz.Knowledge().CreateKnowledgeBase(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetKnowledgeBase 获取知识库.
func (h *Handler) GetKnowledgeBase(c *gin.Context) {
	id := c.Param("id")
	kb, err := h.biz.Knowledge().GetKnowledgeBase(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, kb)
}

// ListKnowledgeBases 列出知识库.
func (h *Handler) ListKnowledgeBases(c *gin.Context) {
	kbs, err := h.biz.Knowledge().ListKnowledgeBases(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": kbs, "total": len(kbs)})
}

// UpdateKnowledgeBase 更新知识库.
func (h *Handler) UpdateKnowledgeBase(c *gin.Context) {
	id := c.Param("id")
	var req model.KnowledgeBase
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id

	if err := h.biz.Knowledge().UpdateKnowledgeBase(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, req)
}

// DeleteKnowledgeBase 删除知识库.
func (h *Handler) DeleteKnowledgeBase(c *gin.Context) {
	id := c.Param("id")
	if err := h.biz.Knowledge().DeleteKnowledgeBase(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// CreateDocument 创建文档.
func (h *Handler) CreateDocument(c *gin.Context) {
	kbID := c.Param("id")
	var req model.KnowledgeDocument
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.KnowledgeBaseID = kbID

	if err := h.biz.Knowledge().CreateDocument(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetDocument 获取文档.
func (h *Handler) GetDocument(c *gin.Context) {
	id := c.Param("id")
	doc, err := h.biz.Knowledge().GetDocument(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, doc)
}

// ListDocuments 列出文档.
func (h *Handler) ListDocuments(c *gin.Context) {
	kbID := c.Param("id")
	docs, err := h.biz.Knowledge().ListDocuments(c.Request.Context(), kbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": docs, "total": len(docs)})
}

// DeleteDocument 删除文档.
func (h *Handler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	if err := h.biz.Knowledge().DeleteDocument(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ListChunksRequest 列出分块请求.
type ListChunksRequest struct {
	Limit  int `form:"limit,default=20"`
	Offset int `form:"offset,default=0"`
}

// ListChunks 列出分块.
func (h *Handler) ListChunks(c *gin.Context) {
	docID := c.Param("id")
	var req ListChunksRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	chunks, total, err := h.biz.Knowledge().ListChunks(c.Request.Context(), docID, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  chunks,
		"total":  total,
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

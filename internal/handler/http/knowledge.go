// Package http 提供 HTTP Handler.
package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/next-show/internal/biz/knowledge"
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

// ImportDocument 导入文档到知识库（JSON 请求）.
func (h *Handler) ImportDocument(c *gin.Context) {
	kbID := c.Param("id")
	var req knowledge.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.KnowledgeBaseID = kbID

	result, err := h.biz.Knowledge().ImportDocument(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// HybridSearchRequest 混合检索请求.
type HybridSearchRequest struct {
	Query        string  `json:"query" binding:"required"`
	TopK         int     `json:"top_k,omitempty"`
	VectorWeight float64 `json:"vector_weight,omitempty"`
	BM25Weight   float64 `json:"bm25_weight,omitempty"`
}

// HybridSearch 混合检索（向量 + BM25）.
func (h *Handler) HybridSearch(c *gin.Context) {
	kbID := c.Param("id")
	var req HybridSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用 knowledge service 的混合检索
	result, err := h.biz.Knowledge().Search(c.Request.Context(), kbID, req.Query, req.TopK, req.VectorWeight, req.BM25Weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UploadDocument 上传文件到知识库（multipart/form-data）.
func (h *Handler) UploadDocument(c *gin.Context) {
	kbID := c.Param("id")

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required: " + err.Error()})
		return
	}
	defer file.Close()

	// 获取可选参数
	title := c.PostForm("title")
	if title == "" {
		title = header.Filename
	}

	chunkSize := 512
	chunkOverlap := 50
	if cs := c.PostForm("chunk_size"); cs != "" {
		if _, err := fmt.Sscanf(cs, "%d", &chunkSize); err != nil {
			chunkSize = 512
		}
	}
	if co := c.PostForm("chunk_overlap"); co != "" {
		if _, err := fmt.Sscanf(co, "%d", &chunkOverlap); err != nil {
			chunkOverlap = 50
		}
	}

	req := &knowledge.ImportRequest{
		KnowledgeBaseID: kbID,
		Title:           title,
		SourceType:      "file",
		FileName:        header.Filename,
		FileReader:      file,
		ChunkSize:       chunkSize,
		ChunkOverlap:    chunkOverlap,
	}

	result, err := h.biz.Knowledge().ImportDocument(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// SearchKnowledgeBaseRequest 搜索知识库请求.
type SearchKnowledgeBaseRequest struct {
	Query        string  `json:"query" binding:"required"`
	TopK         int     `json:"top_k"`
	VectorWeight float64 `json:"vector_weight"`
	BM25Weight   float64 `json:"bm25_weight"`
}

// SearchKnowledgeBase 搜索知识库.
func (h *Handler) SearchKnowledgeBase(c *gin.Context) {
	kbID := c.Param("id")
	var req SearchKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	searchResult, err := h.biz.Knowledge().Search(c.Request.Context(), kbID, req.Query, req.TopK, req.VectorWeight, req.BM25Weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, searchResult)
}

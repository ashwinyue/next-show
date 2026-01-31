// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateSessionRequest 创建会话请求.
type CreateSessionRequest struct {
	AgentID string `json:"agent_id"`
}

// CreateSession 创建会话.
func (h *Handler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 从认证中获取 userID
	userID := "default_user"

	session, err := h.biz.Sessions().Create(c.Request.Context(), userID, req.AgentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ListSessions 列出会话.
func (h *Handler) ListSessions(c *gin.Context) {
	userID := "default_user"
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	sessions, total, err := h.biz.Sessions().List(c.Request.Context(), userID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": sessions,
		"total": total,
	})
}

// GetSession 获取会话详情.
func (h *Handler) GetSession(c *gin.Context) {
	id := c.Param("id")
	session, err := h.biz.Sessions().Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, session)
}

// DeleteSession 删除会话.
func (h *Handler) DeleteSession(c *gin.Context) {
	id := c.Param("id")
	if err := h.biz.Sessions().Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetMessages 获取会话消息（对齐 WeKnora: /api/v1/messages/:id/load）.
func (h *Handler) GetMessages(c *gin.Context) {
	sessionID := c.Param("session_id")
	beforeTime := c.Query("before_time")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	messages, err := h.biz.Sessions().GetMessages(c.Request.Context(), sessionID, beforeTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": messages})
}

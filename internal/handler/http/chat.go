// Package http 提供 HTTP Handler 层.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChatRequest Chat 请求.
type ChatRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Stream    bool   `json:"stream"`
}

// Chat 非流式对话.
func (h *Handler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现非流式对话
	c.JSON(http.StatusOK, gin.H{
		"message": "chat endpoint - not implemented yet",
		"request": req,
	})
}

// ChatStream 流式对话（SSE）.
func (h *Handler) ChatStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现流式对话，使用 EventProcessor + SSE Writer
	c.JSON(http.StatusOK, gin.H{
		"message": "chat stream endpoint - not implemented yet",
		"request": req,
	})
}

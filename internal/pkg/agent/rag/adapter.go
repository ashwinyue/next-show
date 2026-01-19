// Package rag 提供 RAG 图编排能力.
package rag

import (
	"context"

	agenttools "github.com/ashwinyue/next-show/internal/pkg/agent/tools"
)

// KnowledgeServiceAdapter 将 KnowledgeService 适配为 KnowledgeSearcher.
type KnowledgeServiceAdapter struct {
	service agenttools.KnowledgeService
}

// NewKnowledgeServiceAdapter 创建适配器.
func NewKnowledgeServiceAdapter(service agenttools.KnowledgeService) *KnowledgeServiceAdapter {
	return &KnowledgeServiceAdapter{service: service}
}

// SemanticSearch 执行语义搜索.
func (a *KnowledgeServiceAdapter) SemanticSearch(ctx context.Context, query string, kbIDs []string, topK int) ([]*SourceChunk, error) {
	result, err := a.service.SemanticSearch(ctx, &agenttools.SemanticSearchRequest{
		Queries:          []string{query},
		KnowledgeBaseIDs: kbIDs,
		TopK:             topK,
	})
	if err != nil {
		return nil, err
	}

	chunks := make([]*SourceChunk, 0, len(result.Chunks))
	for _, r := range result.Chunks {
		chunks = append(chunks, &SourceChunk{
			ChunkID:       r.ID,
			DocumentID:    r.DocumentID,
			Content:       r.Content,
			Score:         r.Score,
			DocumentTitle: r.DocumentTitle,
		})
	}

	return chunks, nil
}

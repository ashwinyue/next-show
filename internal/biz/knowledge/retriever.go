// Package knowledge 提供知识库业务逻辑.
package knowledge

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
	"github.com/cloudwego/eino/components/embedding"
)

// RetrieverService 基于 pgvector 的检索服务
type RetrieverService struct {
	embedder embedding.Embedder
	store    store.KnowledgeStore
}

// NewRetrieverService 创建新的检索服务
func NewRetrieverService(store store.KnowledgeStore, embedder embedding.Embedder) *RetrieverService {
	return &RetrieverService{
		embedder: embedder,
		store:    store,
	}
}

// SearchRequest 检索请求
type SearchRequest struct {
	KnowledgeBaseIDs []string `json:"knowledge_base_ids"`
	Query            string   `json:"query"`
	Limit            int      `json:"limit,omitempty"`
	ScoreThreshold   *float64 `json:"score_threshold,omitempty"`
}

// VectorSearchResult 向量检索结果
type VectorSearchResult struct {
	Chunk  *model.KnowledgeChunk `json:"chunk"`
	Score  float64               `json:"score"`
	Source string                `json:"source,omitempty"`
}

// Search 执行向量检索
func (r *RetrieverService) Search(ctx context.Context, req *SearchRequest) ([]*VectorSearchResult, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// 1. 将查询文本转换为向量
	queryVectors, err := r.embedder.EmbedStrings(ctx, []string{req.Query})
	if err != nil {
		return nil, fmt.Errorf("embed query failed: %w", err)
	}

	if len(queryVectors) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}

	// 转换为 float32 切片
	queryVector := make([]float32, len(queryVectors[0]))
	for i, v := range queryVectors[0] {
		queryVector[i] = float32(v)
	}

	// 2. 执行向量搜索
	results, err := r.store.SearchChunksByVector(ctx, req.KnowledgeBaseIDs, queryVector, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	var searchResults []*VectorSearchResult
	for _, result := range results {
		searchResults = append(searchResults, &VectorSearchResult{
			Chunk:  result.Chunk,
			Score:  result.Score,
			Source: "vector",
		})
	}

	return searchResults, nil
}

// SearchHybrid 执行混合检索（向量 + 全文）
func (r *RetrieverService) SearchHybrid(ctx context.Context, req *SearchRequest) ([]*VectorSearchResult, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// 1. 将查询文本转换为向量
	queryVectors, err := r.embedder.EmbedStrings(ctx, []string{req.Query})
	if err != nil {
		return nil, fmt.Errorf("embed query failed: %w", err)
	}

	if len(queryVectors) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}

	// 转换为 float32 切片
	queryVector := make([]float32, len(queryVectors[0]))
	for i, v := range queryVectors[0] {
		queryVector[i] = float32(v)
	}

	// 2. 执行混合搜索
	results, err := r.store.HybridSearch(ctx, req.KnowledgeBaseIDs, queryVector, req.Query, req.Limit, 0.7, 0.3)
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}

	var searchResults []*VectorSearchResult
	for _, result := range results {
		searchResults = append(searchResults, &VectorSearchResult{
			Chunk:  result.Chunk,
			Score:  result.Score,
			Source: "hybrid",
		})
	}

	return searchResults, nil
}

// SearchFullText 执行全文检索
func (r *RetrieverService) SearchFullText(ctx context.Context, req *SearchRequest) ([]*VectorSearchResult, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// 执行全文搜索
	results, err := r.store.SearchChunksByFullText(ctx, req.KnowledgeBaseIDs, req.Query, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("full text search failed: %w", err)
	}

	var searchResults []*VectorSearchResult
	for _, result := range results {
		searchResults = append(searchResults, &VectorSearchResult{
			Chunk:  result.Chunk,
			Score:  result.Score,
			Source: "fulltext",
		})
	}

	return searchResults, nil
}

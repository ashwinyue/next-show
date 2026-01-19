// Package knowledge 提供知识库业务逻辑.
package knowledge

import (
	"context"

	"github.com/cloudwego/eino/components/embedding"

	"github.com/ashwinyue/next-show/internal/pkg/agent/tools"
	"github.com/ashwinyue/next-show/internal/store"
)

// Service 知识库服务实现.
type Service struct {
	store          store.Store
	embeddingModel embedding.Embedder
}

// Config 知识库服务配置.
type Config struct {
	Store          store.Store
	EmbeddingModel embedding.Embedder
}

// NewService 创建知识库服务.
func NewService(cfg *Config) *Service {
	return &Service{
		store:          cfg.Store,
		embeddingModel: cfg.EmbeddingModel,
	}
}

// SemanticSearch 语义搜索.
func (s *Service) SemanticSearch(ctx context.Context, req *tools.SemanticSearchRequest) (*tools.SemanticSearchResult, error) {
	if s.embeddingModel == nil {
		return &tools.SemanticSearchResult{
			Chunks:     []*tools.ChunkResult{},
			TotalCount: 0,
		}, nil
	}

	// 合并所有查询为一个查询向量
	queryText := ""
	for i, q := range req.Queries {
		if i > 0 {
			queryText += " "
		}
		queryText += q
	}

	// 生成查询向量
	embeddings, err := s.embeddingModel.EmbedStrings(ctx, []string{queryText})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return &tools.SemanticSearchResult{
			Chunks:     []*tools.ChunkResult{},
			TotalCount: 0,
		}, nil
	}

	// 转换 float64 到 float32
	queryVector := make([]float32, len(embeddings[0]))
	for i, v := range embeddings[0] {
		queryVector[i] = float32(v)
	}

	// 执行向量搜索
	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}

	results, err := s.store.Knowledge().SearchChunksByVector(ctx, req.KnowledgeBaseIDs, queryVector, topK)
	if err != nil {
		return nil, err
	}

	// 转换结果
	chunks := make([]*tools.ChunkResult, 0, len(results))
	for _, r := range results {
		// 获取文档标题
		docTitle := ""
		if doc, err := s.store.Knowledge().GetDocument(ctx, r.Chunk.DocumentID); err == nil && doc != nil {
			docTitle = doc.Title
		}

		chunks = append(chunks, &tools.ChunkResult{
			ID:              r.Chunk.ID,
			DocumentID:      r.Chunk.DocumentID,
			DocumentTitle:   docTitle,
			KnowledgeBaseID: r.Chunk.KnowledgeBaseID,
			ChunkIndex:      r.Chunk.ChunkIndex,
			Content:         r.Chunk.Content,
			Score:           r.Score,
		})
	}

	return &tools.SemanticSearchResult{
		Chunks:     chunks,
		TotalCount: len(chunks),
	}, nil
}

// KeywordSearch 关键词搜索.
func (s *Service) KeywordSearch(ctx context.Context, req *tools.KeywordSearchRequest) (*tools.KeywordSearchResult, error) {
	topK := req.TopK
	if topK <= 0 {
		topK = 20
	}

	results, err := s.store.Knowledge().SearchChunksByKeyword(ctx, req.KnowledgeBaseIDs, req.Keywords, topK)
	if err != nil {
		return nil, err
	}

	// 转换结果
	chunks := make([]*tools.ChunkResult, 0, len(results))
	for _, r := range results {
		// 获取文档标题
		docTitle := ""
		if doc, err := s.store.Knowledge().GetDocument(ctx, r.DocumentID); err == nil && doc != nil {
			docTitle = doc.Title
		}

		chunks = append(chunks, &tools.ChunkResult{
			ID:              r.ID,
			DocumentID:      r.DocumentID,
			DocumentTitle:   docTitle,
			KnowledgeBaseID: r.KnowledgeBaseID,
			ChunkIndex:      r.ChunkIndex,
			Content:         r.Content,
		})
	}

	return &tools.KeywordSearchResult{
		Chunks:     chunks,
		TotalCount: len(chunks),
	}, nil
}

// ListChunks 列出文档分块.
func (s *Service) ListChunks(ctx context.Context, req *tools.ListChunksRequest) (*tools.ListChunksResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	results, total, err := s.store.Knowledge().ListChunksByDocument(ctx, req.DocumentID, limit, offset)
	if err != nil {
		return nil, err
	}

	// 获取文档标题
	docTitle := ""
	if doc, err := s.store.Knowledge().GetDocument(ctx, req.DocumentID); err == nil && doc != nil {
		docTitle = doc.Title
	}

	// 转换结果
	chunks := make([]*tools.ChunkResult, 0, len(results))
	for _, r := range results {
		chunks = append(chunks, &tools.ChunkResult{
			ID:              r.ID,
			DocumentID:      r.DocumentID,
			DocumentTitle:   docTitle,
			KnowledgeBaseID: r.KnowledgeBaseID,
			ChunkIndex:      r.ChunkIndex,
			Content:         r.Content,
		})
	}

	return &tools.ListChunksResult{
		Chunks:     chunks,
		TotalCount: int(total),
	}, nil
}

// Ensure interface is implemented
var _ tools.KnowledgeService = (*Service)(nil)

// Package knowledge 提供知识库业务逻辑.
package knowledge

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/reranker/score"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"

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

// HybridSearch 混合检索（向量 + BM25）.
func (s *Service) HybridSearch(ctx context.Context, req *tools.HybridSearchRequest) (*tools.HybridSearchResult, error) {
	if s.embeddingModel == nil {
		return &tools.HybridSearchResult{
			Chunks:     []*tools.ChunkResult{},
			TotalCount: 0,
		}, nil
	}

	// 生成查询向量
	embeddings, err := s.embeddingModel.EmbedStrings(ctx, []string{req.Query})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return &tools.HybridSearchResult{
			Chunks:     []*tools.ChunkResult{},
			TotalCount: 0,
		}, nil
	}

	// 转换 float64 到 float32
	queryVector := make([]float32, len(embeddings[0]))
	for i, v := range embeddings[0] {
		queryVector[i] = float32(v)
	}

	// 设置默认权重
	vectorWeight := req.VectorWeight
	if vectorWeight <= 0 {
		vectorWeight = 0.7
	}
	bm25Weight := req.BM25Weight
	if bm25Weight <= 0 {
		bm25Weight = 0.3
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}

	// 执行混合检索
	results, err := s.store.Knowledge().HybridSearch(ctx, req.KnowledgeBaseIDs, queryVector, req.Query, topK, vectorWeight, bm25Weight)
	if err != nil {
		return nil, err
	}

	// 转换结果
	chunks := make([]*tools.ChunkResult, 0, len(results))
	for _, r := range results {
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

	return &tools.HybridSearchResult{
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

// RerankedSearch 带重排序的混合检索.
func (s *Service) RerankedSearch(ctx context.Context, req *tools.HybridSearchRequest) (*tools.HybridSearchResult, error) {
	// 先执行混合检索
	result, err := s.HybridSearch(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(result.Chunks) <= 1 {
		return result, nil
	}

	// 转换为 schema.Document 用于重排序
	docs := make([]*schema.Document, len(result.Chunks))
	for i, chunk := range result.Chunks {
		docs[i] = &schema.Document{
			ID:      chunk.ID,
			Content: chunk.Content,
			MetaData: map[string]any{
				"document_id":       chunk.DocumentID,
				"document_title":    chunk.DocumentTitle,
				"knowledge_base_id": chunk.KnowledgeBaseID,
				"chunk_index":       chunk.ChunkIndex,
				"original_score":    chunk.Score,
			},
		}
		docs[i].WithScore(chunk.Score)
	}

	// 使用 score reranker 重排序（高分放首尾，利用 LLM 首尾效应）
	reranker, err := score.NewReranker(ctx, &score.Config{})
	if err != nil {
		return result, nil // 重排序失败，返回原结果
	}

	rerankedDocs, err := reranker.Transform(ctx, docs)
	if err != nil {
		return result, nil
	}

	// 转换回 ChunkResult
	rerankedChunks := make([]*tools.ChunkResult, len(rerankedDocs))
	for i, doc := range rerankedDocs {
		rerankedChunks[i] = &tools.ChunkResult{
			ID:              doc.ID,
			Content:         doc.Content,
			DocumentID:      doc.MetaData["document_id"].(string),
			DocumentTitle:   doc.MetaData["document_title"].(string),
			KnowledgeBaseID: doc.MetaData["knowledge_base_id"].(string),
			ChunkIndex:      doc.MetaData["chunk_index"].(int),
			Score:           doc.Score(),
		}
	}

	return &tools.HybridSearchResult{
		Chunks:     rerankedChunks,
		TotalCount: len(rerankedChunks),
	}, nil
}

// Ensure interface is implemented
var _ tools.KnowledgeService = (*Service)(nil)

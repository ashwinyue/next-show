// Package knowledge 提供知识库业务逻辑.
package knowledge

import (
	"context"

	"github.com/cloudwego/eino/components/embedding"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz 知识库业务接口.
type Biz interface {
	// KnowledgeBase
	CreateKnowledgeBase(ctx context.Context, kb *model.KnowledgeBase) error
	GetKnowledgeBase(ctx context.Context, id string) (*model.KnowledgeBase, error)
	ListKnowledgeBases(ctx context.Context) ([]*model.KnowledgeBase, error)
	UpdateKnowledgeBase(ctx context.Context, kb *model.KnowledgeBase) error
	DeleteKnowledgeBase(ctx context.Context, id string) error

	// Document
	CreateDocument(ctx context.Context, doc *model.KnowledgeDocument) error
	GetDocument(ctx context.Context, id string) (*model.KnowledgeDocument, error)
	ListDocuments(ctx context.Context, kbID string) ([]*model.KnowledgeDocument, error)
	DeleteDocument(ctx context.Context, id string) error

	// Chunk
	GetChunk(ctx context.Context, id string) (*model.KnowledgeChunk, error)
	ListChunks(ctx context.Context, docID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error)
	ListChunksByKnowledgeBase(ctx context.Context, kbID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error)
	UpdateChunk(ctx context.Context, chunk *model.KnowledgeChunk) error
	DeleteChunk(ctx context.Context, id string) error

	// Tag
	CreateTag(ctx context.Context, tag *model.KnowledgeTag) error
	GetTag(ctx context.Context, id string) (*model.KnowledgeTag, error)
	ListTags(ctx context.Context, kbID string) ([]*model.KnowledgeTag, error)
	UpdateTag(ctx context.Context, tag *model.KnowledgeTag) error
	DeleteTag(ctx context.Context, id string) error

	// ChunkTag
	AddTagToChunk(ctx context.Context, chunkID, tagID string) error
	RemoveTagFromChunk(ctx context.Context, chunkID, tagID string) error
	ListTagsByChunk(ctx context.Context, chunkID string) ([]*model.KnowledgeTag, error)
	ListChunksByTag(ctx context.Context, tagID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error)

	// Import
	ImportDocument(ctx context.Context, req *ImportRequest) (*ImportResult, error)

	// Search
	Search(ctx context.Context, kbID, query string, topK int, vectorWeight, bm25Weight float64) (*SearchResult, error)
}

// bizImpl 知识库业务实现.
type bizImpl struct {
	store    store.Store
	embedder embedding.Embedder
}

// NewBiz 创建知识库业务实例.
func NewBiz(s store.Store, embedder embedding.Embedder) Biz {
	return &bizImpl{store: s, embedder: embedder}
}

func (b *bizImpl) CreateKnowledgeBase(ctx context.Context, kb *model.KnowledgeBase) error {
	return b.store.Knowledge().CreateKnowledgeBase(ctx, kb)
}

func (b *bizImpl) GetKnowledgeBase(ctx context.Context, id string) (*model.KnowledgeBase, error) {
	return b.store.Knowledge().GetKnowledgeBase(ctx, id)
}

func (b *bizImpl) ListKnowledgeBases(ctx context.Context) ([]*model.KnowledgeBase, error) {
	return b.store.Knowledge().ListKnowledgeBases(ctx)
}

func (b *bizImpl) UpdateKnowledgeBase(ctx context.Context, kb *model.KnowledgeBase) error {
	return b.store.Knowledge().UpdateKnowledgeBase(ctx, kb)
}

func (b *bizImpl) DeleteKnowledgeBase(ctx context.Context, id string) error {
	return b.store.Knowledge().DeleteKnowledgeBase(ctx, id)
}

func (b *bizImpl) CreateDocument(ctx context.Context, doc *model.KnowledgeDocument) error {
	return b.store.Knowledge().CreateDocument(ctx, doc)
}

func (b *bizImpl) GetDocument(ctx context.Context, id string) (*model.KnowledgeDocument, error) {
	return b.store.Knowledge().GetDocument(ctx, id)
}

func (b *bizImpl) ListDocuments(ctx context.Context, kbID string) ([]*model.KnowledgeDocument, error) {
	return b.store.Knowledge().ListDocumentsByKnowledgeBase(ctx, kbID)
}

func (b *bizImpl) DeleteDocument(ctx context.Context, id string) error {
	return b.store.Knowledge().DeleteDocument(ctx, id)
}

func (b *bizImpl) GetChunk(ctx context.Context, id string) (*model.KnowledgeChunk, error) {
	return b.store.Knowledge().GetChunk(ctx, id)
}

func (b *bizImpl) ListChunks(ctx context.Context, docID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error) {
	return b.store.Knowledge().ListChunksByDocument(ctx, docID, limit, offset)
}

func (b *bizImpl) ListChunksByKnowledgeBase(ctx context.Context, kbID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error) {
	return b.store.Knowledge().ListChunksByKnowledgeBase(ctx, kbID, limit, offset)
}

func (b *bizImpl) UpdateChunk(ctx context.Context, chunk *model.KnowledgeChunk) error {
	return b.store.Knowledge().UpdateChunk(ctx, chunk)
}

func (b *bizImpl) DeleteChunk(ctx context.Context, id string) error {
	return b.store.Knowledge().DeleteChunk(ctx, id)
}

// Tag 相关方法

func (b *bizImpl) CreateTag(ctx context.Context, tag *model.KnowledgeTag) error {
	return b.store.Knowledge().CreateTag(ctx, tag)
}

func (b *bizImpl) GetTag(ctx context.Context, id string) (*model.KnowledgeTag, error) {
	return b.store.Knowledge().GetTag(ctx, id)
}

func (b *bizImpl) ListTags(ctx context.Context, kbID string) ([]*model.KnowledgeTag, error) {
	return b.store.Knowledge().ListTagsByKnowledgeBase(ctx, kbID)
}

func (b *bizImpl) UpdateTag(ctx context.Context, tag *model.KnowledgeTag) error {
	return b.store.Knowledge().UpdateTag(ctx, tag)
}

func (b *bizImpl) DeleteTag(ctx context.Context, id string) error {
	return b.store.Knowledge().DeleteTag(ctx, id)
}

// ChunkTag 相关方法

func (b *bizImpl) AddTagToChunk(ctx context.Context, chunkID, tagID string) error {
	return b.store.Knowledge().AddTagToChunk(ctx, chunkID, tagID)
}

func (b *bizImpl) RemoveTagFromChunk(ctx context.Context, chunkID, tagID string) error {
	return b.store.Knowledge().RemoveTagFromChunk(ctx, chunkID, tagID)
}

func (b *bizImpl) ListTagsByChunk(ctx context.Context, chunkID string) ([]*model.KnowledgeTag, error) {
	return b.store.Knowledge().ListTagsByChunk(ctx, chunkID)
}

func (b *bizImpl) ListChunksByTag(ctx context.Context, tagID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error) {
	return b.store.Knowledge().ListChunksByTag(ctx, tagID, limit, offset)
}

// SearchResult 检索结果.
type SearchResult struct {
	Chunks     []*ChunkSearchResult `json:"chunks"`
	TotalCount int                  `json:"total_count"`
}

// ChunkSearchResult 分块检索结果.
type ChunkSearchResult struct {
	ID              string  `json:"id"`
	DocumentID      string  `json:"document_id"`
	DocumentTitle   string  `json:"document_title"`
	KnowledgeBaseID string  `json:"knowledge_base_id"`
	ChunkIndex      int     `json:"chunk_index"`
	Content         string  `json:"content"`
	Score           float64 `json:"score"`
}

// Search 混合检索.
func (b *bizImpl) Search(ctx context.Context, kbID, query string, topK int, vectorWeight, bm25Weight float64) (*SearchResult, error) {
	if b.embedder == nil {
		return &SearchResult{Chunks: []*ChunkSearchResult{}, TotalCount: 0}, nil
	}

	if topK <= 0 {
		topK = 10
	}
	if vectorWeight <= 0 {
		vectorWeight = 0.7
	}
	if bm25Weight <= 0 {
		bm25Weight = 0.3
	}

	// 生成查询向量
	embeddings, err := b.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return &SearchResult{Chunks: []*ChunkSearchResult{}, TotalCount: 0}, nil
	}

	queryVector := make([]float32, len(embeddings[0]))
	for i, v := range embeddings[0] {
		queryVector[i] = float32(v)
	}

	// 执行混合检索
	kbIDs := []string{kbID}
	results, err := b.store.Knowledge().HybridSearch(ctx, kbIDs, queryVector, query, topK, vectorWeight, bm25Weight)
	if err != nil {
		return nil, err
	}

	// 转换结果
	chunks := make([]*ChunkSearchResult, 0, len(results))
	for _, r := range results {
		docTitle := ""
		if doc, err := b.store.Knowledge().GetDocument(ctx, r.Chunk.DocumentID); err == nil && doc != nil {
			docTitle = doc.Title
		}

		chunks = append(chunks, &ChunkSearchResult{
			ID:              r.Chunk.ID,
			DocumentID:      r.Chunk.DocumentID,
			DocumentTitle:   docTitle,
			KnowledgeBaseID: r.Chunk.KnowledgeBaseID,
			ChunkIndex:      r.Chunk.ChunkIndex,
			Content:         r.Chunk.Content,
			Score:           r.Score,
		})
	}

	return &SearchResult{
		Chunks:     chunks,
		TotalCount: len(chunks),
	}, nil
}

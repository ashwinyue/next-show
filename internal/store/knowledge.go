// Package store 提供数据访问层.
package store

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// KnowledgeStore 知识库存储接口.
type KnowledgeStore interface {
	// KnowledgeBase CRUD
	CreateKnowledgeBase(ctx context.Context, kb *model.KnowledgeBase) error
	GetKnowledgeBase(ctx context.Context, id string) (*model.KnowledgeBase, error)
	ListKnowledgeBases(ctx context.Context) ([]*model.KnowledgeBase, error)
	UpdateKnowledgeBase(ctx context.Context, kb *model.KnowledgeBase) error
	DeleteKnowledgeBase(ctx context.Context, id string) error

	// Document CRUD
	CreateDocument(ctx context.Context, doc *model.KnowledgeDocument) error
	GetDocument(ctx context.Context, id string) (*model.KnowledgeDocument, error)
	ListDocumentsByKnowledgeBase(ctx context.Context, kbID string) ([]*model.KnowledgeDocument, error)
	DeleteDocument(ctx context.Context, id string) error

	// Chunk
	GetChunk(ctx context.Context, id string) (*model.KnowledgeChunk, error)
	ListChunksByDocument(ctx context.Context, docID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error)
	SearchChunksByKeyword(ctx context.Context, kbIDs []string, keywords []string, limit int) ([]*model.KnowledgeChunk, error)

	// Vector Search (pgvector)
	SearchChunksByVector(ctx context.Context, kbIDs []string, embedding []float32, limit int) ([]*ChunkWithScore, error)
}

// ChunkWithScore 带分数的分块结果.
type ChunkWithScore struct {
	Chunk *model.KnowledgeChunk
	Score float64
}

type knowledgeStore struct {
	db *gorm.DB
}

func newKnowledgeStore(db *gorm.DB) KnowledgeStore {
	return &knowledgeStore{db: db}
}

func (s *knowledgeStore) CreateKnowledgeBase(ctx context.Context, kb *model.KnowledgeBase) error {
	return s.db.WithContext(ctx).Create(kb).Error
}

func (s *knowledgeStore) GetKnowledgeBase(ctx context.Context, id string) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&kb).Error; err != nil {
		return nil, err
	}
	return &kb, nil
}

func (s *knowledgeStore) ListKnowledgeBases(ctx context.Context) ([]*model.KnowledgeBase, error) {
	var kbs []*model.KnowledgeBase
	if err := s.db.WithContext(ctx).Where("status = ?", model.KnowledgeBaseStatusActive).Find(&kbs).Error; err != nil {
		return nil, err
	}
	return kbs, nil
}

func (s *knowledgeStore) UpdateKnowledgeBase(ctx context.Context, kb *model.KnowledgeBase) error {
	return s.db.WithContext(ctx).Save(kb).Error
}

func (s *knowledgeStore) DeleteKnowledgeBase(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&model.KnowledgeBase{}, "id = ?", id).Error
}

func (s *knowledgeStore) CreateDocument(ctx context.Context, doc *model.KnowledgeDocument) error {
	return s.db.WithContext(ctx).Create(doc).Error
}

func (s *knowledgeStore) GetDocument(ctx context.Context, id string) (*model.KnowledgeDocument, error) {
	var doc model.KnowledgeDocument
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *knowledgeStore) ListDocumentsByKnowledgeBase(ctx context.Context, kbID string) ([]*model.KnowledgeDocument, error) {
	var docs []*model.KnowledgeDocument
	if err := s.db.WithContext(ctx).Where("knowledge_base_id = ?", kbID).Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func (s *knowledgeStore) DeleteDocument(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&model.KnowledgeDocument{}, "id = ?", id).Error
}

func (s *knowledgeStore) GetChunk(ctx context.Context, id string) (*model.KnowledgeChunk, error) {
	var chunk model.KnowledgeChunk
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&chunk).Error; err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (s *knowledgeStore) ListChunksByDocument(ctx context.Context, docID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error) {
	var chunks []*model.KnowledgeChunk
	var total int64

	db := s.db.WithContext(ctx).Model(&model.KnowledgeChunk{}).Where("document_id = ? AND is_enabled = ?", docID, true)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("chunk_index ASC").Offset(offset).Limit(limit).Find(&chunks).Error; err != nil {
		return nil, 0, err
	}

	return chunks, total, nil
}

func (s *knowledgeStore) SearchChunksByKeyword(ctx context.Context, kbIDs []string, keywords []string, limit int) ([]*model.KnowledgeChunk, error) {
	if len(keywords) == 0 {
		return nil, nil
	}

	db := s.db.WithContext(ctx).Model(&model.KnowledgeChunk{}).Where("is_enabled = ?", true)

	if len(kbIDs) > 0 {
		db = db.Where("knowledge_base_id IN ?", kbIDs)
	}

	// 构建关键词搜索条件 (ILIKE for case-insensitive)
	for _, kw := range keywords {
		db = db.Where("content ILIKE ?", "%"+kw+"%")
	}

	var chunks []*model.KnowledgeChunk
	if err := db.Limit(limit).Find(&chunks).Error; err != nil {
		return nil, err
	}

	return chunks, nil
}

func (s *knowledgeStore) SearchChunksByVector(ctx context.Context, kbIDs []string, embedding []float32, limit int) ([]*ChunkWithScore, error) {
	// 使用 pgvector 的余弦相似度搜索
	// SQL: SELECT c.*, 1 - (e.embedding <=> $1) as score FROM knowledge_chunks c
	//      JOIN embeddings e ON e.chunk_id = c.id
	//      WHERE c.knowledge_base_id IN ($2) AND c.is_enabled = true
	//      ORDER BY e.embedding <=> $1 LIMIT $3

	query := `
		SELECT c.id, c.knowledge_base_id, c.document_id, c.chunk_index, c.content, 
		       c.content_hash, c.metadata, c.is_enabled, c.created_at, c.updated_at,
		       1 - (e.embedding <=> $1::vector) as score
		FROM knowledge_chunks c
		JOIN embeddings e ON e.chunk_id = c.id
		WHERE c.is_enabled = true
	`

	args := []interface{}{vectorToString(embedding)}
	argIdx := 2

	if len(kbIDs) > 0 {
		query += " AND c.knowledge_base_id = ANY($" + itoa(argIdx) + ")"
		args = append(args, kbIDs)
		argIdx++
	}

	query += " ORDER BY e.embedding <=> $1::vector LIMIT $" + itoa(argIdx)
	args = append(args, limit)

	rows, err := s.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*ChunkWithScore
	for rows.Next() {
		var chunk model.KnowledgeChunk
		var score float64
		if err := rows.Scan(
			&chunk.ID, &chunk.KnowledgeBaseID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content,
			&chunk.ContentHash, &chunk.Metadata, &chunk.IsEnabled, &chunk.CreatedAt, &chunk.UpdatedAt,
			&score,
		); err != nil {
			return nil, err
		}
		results = append(results, &ChunkWithScore{
			Chunk: &chunk,
			Score: score,
		})
	}

	return results, nil
}

// vectorToString 将 float32 切片转换为 pgvector 格式字符串.
func vectorToString(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	s := "["
	for i, f := range v {
		if i > 0 {
			s += ","
		}
		s += ftoa(f)
	}
	s += "]"
	return s
}

func itoa(i int) string {
	return string(rune('0'+i%10)) + ""
}

func ftoa(f float32) string {
	return fmt.Sprintf("%f", f)
}

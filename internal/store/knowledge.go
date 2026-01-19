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
	UpdateDocument(ctx context.Context, doc *model.KnowledgeDocument) error
	DeleteDocument(ctx context.Context, id string) error

	// Chunk CRUD
	GetChunk(ctx context.Context, id string) (*model.KnowledgeChunk, error)
	ListChunksByDocument(ctx context.Context, docID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error)
	ListChunksByKnowledgeBase(ctx context.Context, kbID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error)
	UpdateChunk(ctx context.Context, chunk *model.KnowledgeChunk) error
	DeleteChunk(ctx context.Context, id string) error
	SearchChunksByKeyword(ctx context.Context, kbIDs []string, keywords []string, limit int) ([]*model.KnowledgeChunk, error)

	// Chunk & Embedding Write
	CreateChunks(ctx context.Context, chunks []*model.KnowledgeChunk) error
	CreateEmbeddings(ctx context.Context, embeddings []*model.Embedding) error

	// Vector Search (pgvector)
	SearchChunksByVector(ctx context.Context, kbIDs []string, embedding []float32, limit int) ([]*ChunkWithScore, error)

	// BM25 Full-Text Search
	SearchChunksByFullText(ctx context.Context, kbIDs []string, query string, limit int) ([]*ChunkWithScore, error)

	// Hybrid Search (Vector + BM25)
	HybridSearch(ctx context.Context, kbIDs []string, embedding []float32, query string, limit int, vectorWeight, bm25Weight float64) ([]*ChunkWithScore, error)

	// Tag CRUD
	CreateTag(ctx context.Context, tag *model.KnowledgeTag) error
	GetTag(ctx context.Context, id string) (*model.KnowledgeTag, error)
	ListTagsByKnowledgeBase(ctx context.Context, kbID string) ([]*model.KnowledgeTag, error)
	UpdateTag(ctx context.Context, tag *model.KnowledgeTag) error
	DeleteTag(ctx context.Context, id string) error

	// ChunkTag
	AddTagToChunk(ctx context.Context, chunkID, tagID string) error
	RemoveTagFromChunk(ctx context.Context, chunkID, tagID string) error
	ListTagsByChunk(ctx context.Context, chunkID string) ([]*model.KnowledgeTag, error)
	ListChunksByTag(ctx context.Context, tagID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error)
}

func (s *knowledgeStore) CreateChunks(ctx context.Context, chunks []*model.KnowledgeChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Create(&chunks).Error
}

func (s *knowledgeStore) CreateEmbeddings(ctx context.Context, embeddings []*model.Embedding) error {
	if len(embeddings) == 0 {
		return nil
	}

	// 使用 Raw SQL 写入 embedding，确保 pgvector cast 正确
	query := `INSERT INTO embeddings (knowledge_base_id, chunk_id, embedding, embedding_dim, embedding_model, metadata)
			VALUES ($1, $2, $3::vector, $4, $5, $6)
			ON CONFLICT (chunk_id) DO UPDATE SET embedding = EXCLUDED.embedding, embedding_dim = EXCLUDED.embedding_dim, embedding_model = EXCLUDED.embedding_model, metadata = EXCLUDED.metadata`

	for _, e := range embeddings {
		if e == nil {
			continue
		}
		if err := s.db.WithContext(ctx).Exec(query,
			e.KnowledgeBaseID,
			e.ChunkID,
			vectorToString(e.Embedding),
			e.EmbeddingDim,
			e.EmbeddingModel,
			e.Metadata,
		).Error; err != nil {
			return err
		}
	}
	return nil
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

func (s *knowledgeStore) UpdateDocument(ctx context.Context, doc *model.KnowledgeDocument) error {
	return s.db.WithContext(ctx).Save(doc).Error
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

// SearchChunksByFullText 使用 PostgreSQL 全文搜索 (BM25-like).
func (s *knowledgeStore) SearchChunksByFullText(ctx context.Context, kbIDs []string, query string, limit int) ([]*ChunkWithScore, error) {
	if query == "" {
		return nil, nil
	}

	// 使用 PostgreSQL 全文搜索，ts_rank 提供类似 BM25 的排名
	// plainto_tsquery 自动处理查询词
	sqlQuery := `
		SELECT c.id, c.knowledge_base_id, c.document_id, c.chunk_index, c.content, 
		       c.content_hash, c.metadata, c.is_enabled, c.created_at, c.updated_at,
		       ts_rank_cd(to_tsvector('simple', c.content), plainto_tsquery('simple', $1)) as score
		FROM knowledge_chunks c
		WHERE c.is_enabled = true
		  AND to_tsvector('simple', c.content) @@ plainto_tsquery('simple', $1)
	`

	args := []interface{}{query}
	argIdx := 2

	if len(kbIDs) > 0 {
		sqlQuery += " AND c.knowledge_base_id = ANY($" + fmt.Sprintf("%d", argIdx) + ")"
		args = append(args, kbIDs)
		argIdx++
	}

	sqlQuery += " ORDER BY score DESC LIMIT $" + fmt.Sprintf("%d", argIdx)
	args = append(args, limit)

	rows, err := s.db.WithContext(ctx).Raw(sqlQuery, args...).Rows()
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

// HybridSearch 混合检索（向量 + 全文搜索）.
func (s *knowledgeStore) HybridSearch(ctx context.Context, kbIDs []string, embedding []float32, query string, limit int, vectorWeight, bm25Weight float64) ([]*ChunkWithScore, error) {
	// 使用 RRF (Reciprocal Rank Fusion) 合并向量搜索和全文搜索结果
	// hybrid_score = vectorWeight * vector_score + bm25Weight * bm25_score
	sqlQuery := `
		WITH vector_results AS (
			SELECT c.id, c.knowledge_base_id, c.document_id, c.chunk_index, c.content, 
			       c.content_hash, c.metadata, c.is_enabled, c.created_at, c.updated_at,
			       1 - (e.embedding <=> $1::vector) as vector_score,
			       0::float as bm25_score
			FROM knowledge_chunks c
			JOIN embeddings e ON e.chunk_id = c.id
			WHERE c.is_enabled = true
	`

	args := []interface{}{vectorToString(embedding)}
	argIdx := 2

	if len(kbIDs) > 0 {
		sqlQuery += " AND c.knowledge_base_id = ANY($" + fmt.Sprintf("%d", argIdx) + ")"
		args = append(args, kbIDs)
		argIdx++
	}

	sqlQuery += " ORDER BY e.embedding <=> $1::vector LIMIT $" + fmt.Sprintf("%d", argIdx)
	args = append(args, limit*2) // 获取更多结果用于合并
	argIdx++

	sqlQuery += `
		),
		bm25_results AS (
			SELECT c.id, c.knowledge_base_id, c.document_id, c.chunk_index, c.content, 
			       c.content_hash, c.metadata, c.is_enabled, c.created_at, c.updated_at,
			       0::float as vector_score,
			       ts_rank_cd(to_tsvector('simple', c.content), plainto_tsquery('simple', $` + fmt.Sprintf("%d", argIdx) + `)) as bm25_score
			FROM knowledge_chunks c
			WHERE c.is_enabled = true
			  AND to_tsvector('simple', c.content) @@ plainto_tsquery('simple', $` + fmt.Sprintf("%d", argIdx) + `)
	`
	args = append(args, query)
	argIdx++

	if len(kbIDs) > 0 {
		sqlQuery += " AND c.knowledge_base_id = ANY($" + fmt.Sprintf("%d", argIdx) + ")"
		args = append(args, kbIDs)
		argIdx++
	}

	sqlQuery += " ORDER BY bm25_score DESC LIMIT $" + fmt.Sprintf("%d", argIdx)
	args = append(args, limit*2)
	argIdx++

	// 合并结果并计算混合分数
	sqlQuery += `
		),
		combined AS (
			SELECT id, knowledge_base_id, document_id, chunk_index, content, 
			       content_hash, metadata, is_enabled, created_at, updated_at,
			       MAX(vector_score) as vector_score,
			       MAX(bm25_score) as bm25_score
			FROM (
				SELECT * FROM vector_results
				UNION ALL
				SELECT * FROM bm25_results
			) all_results
			GROUP BY id, knowledge_base_id, document_id, chunk_index, content, 
			         content_hash, metadata, is_enabled, created_at, updated_at
		)
		SELECT id, knowledge_base_id, document_id, chunk_index, content, 
		       content_hash, metadata, is_enabled, created_at, updated_at,
		       ($` + fmt.Sprintf("%d", argIdx) + ` * vector_score + $` + fmt.Sprintf("%d", argIdx+1) + ` * bm25_score) as score
		FROM combined
		ORDER BY score DESC
		LIMIT $` + fmt.Sprintf("%d", argIdx+2)

	args = append(args, vectorWeight, bm25Weight, limit)

	rows, err := s.db.WithContext(ctx).Raw(sqlQuery, args...).Rows()
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

// Chunk 扩展方法

func (s *knowledgeStore) ListChunksByKnowledgeBase(ctx context.Context, kbID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error) {
	var chunks []*model.KnowledgeChunk
	var total int64

	db := s.db.WithContext(ctx).Model(&model.KnowledgeChunk{}).Where("knowledge_base_id = ?", kbID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&chunks).Error; err != nil {
		return nil, 0, err
	}

	return chunks, total, nil
}

func (s *knowledgeStore) UpdateChunk(ctx context.Context, chunk *model.KnowledgeChunk) error {
	return s.db.WithContext(ctx).Save(chunk).Error
}

func (s *knowledgeStore) DeleteChunk(ctx context.Context, id string) error {
	// 先删除关联的 embedding 和 chunk_tags
	if err := s.db.WithContext(ctx).Where("chunk_id = ?", id).Delete(&model.Embedding{}).Error; err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Where("chunk_id = ?", id).Delete(&model.ChunkTag{}).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Delete(&model.KnowledgeChunk{}, "id = ?", id).Error
}

// Tag CRUD

func (s *knowledgeStore) CreateTag(ctx context.Context, tag *model.KnowledgeTag) error {
	return s.db.WithContext(ctx).Create(tag).Error
}

func (s *knowledgeStore) GetTag(ctx context.Context, id string) (*model.KnowledgeTag, error) {
	var tag model.KnowledgeTag
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (s *knowledgeStore) ListTagsByKnowledgeBase(ctx context.Context, kbID string) ([]*model.KnowledgeTag, error) {
	var tags []*model.KnowledgeTag
	if err := s.db.WithContext(ctx).Where("knowledge_base_id = ?", kbID).Order("name ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (s *knowledgeStore) UpdateTag(ctx context.Context, tag *model.KnowledgeTag) error {
	return s.db.WithContext(ctx).Save(tag).Error
}

func (s *knowledgeStore) DeleteTag(ctx context.Context, id string) error {
	// 先删除关联的 chunk_tags
	if err := s.db.WithContext(ctx).Where("tag_id = ?", id).Delete(&model.ChunkTag{}).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Delete(&model.KnowledgeTag{}, "id = ?", id).Error
}

// ChunkTag 关联方法

func (s *knowledgeStore) AddTagToChunk(ctx context.Context, chunkID, tagID string) error {
	chunkTag := &model.ChunkTag{
		ChunkID: chunkID,
		TagID:   tagID,
	}
	// 使用 upsert 避免重复
	return s.db.WithContext(ctx).Where("chunk_id = ? AND tag_id = ?", chunkID, tagID).
		FirstOrCreate(chunkTag).Error
}

func (s *knowledgeStore) RemoveTagFromChunk(ctx context.Context, chunkID, tagID string) error {
	return s.db.WithContext(ctx).Where("chunk_id = ? AND tag_id = ?", chunkID, tagID).
		Delete(&model.ChunkTag{}).Error
}

func (s *knowledgeStore) ListTagsByChunk(ctx context.Context, chunkID string) ([]*model.KnowledgeTag, error) {
	var tags []*model.KnowledgeTag
	err := s.db.WithContext(ctx).
		Joins("JOIN chunk_tags ON chunk_tags.tag_id = knowledge_tags.id").
		Where("chunk_tags.chunk_id = ?", chunkID).
		Find(&tags).Error
	return tags, err
}

func (s *knowledgeStore) ListChunksByTag(ctx context.Context, tagID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error) {
	var chunks []*model.KnowledgeChunk
	var total int64

	db := s.db.WithContext(ctx).Model(&model.KnowledgeChunk{}).
		Joins("JOIN chunk_tags ON chunk_tags.chunk_id = knowledge_chunks.id").
		Where("chunk_tags.tag_id = ?", tagID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("knowledge_chunks.created_at DESC").Offset(offset).Limit(limit).Find(&chunks).Error; err != nil {
		return nil, 0, err
	}

	return chunks, total, nil
}

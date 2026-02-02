// Package store 提供数据访问层.
package store

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// DistanceFunction represents the distance function for vector similarity search.
type DistanceFunction string

const (
	// DistanceCosine uses cosine distance for similarity.
	DistanceCosine DistanceFunction = "cosine"
	// DistanceL2 uses Euclidean (L2) distance.
	DistanceL2 DistanceFunction = "l2"
	// DistanceIP uses inner product distance.
	DistanceIP DistanceFunction = "ip"
)

// String returns the string representation of the distance function.
func (d DistanceFunction) String() string {
	return string(d)
}

// Operator returns the SQL operator for the distance function.
func (d DistanceFunction) Operator() string {
	switch d {
	case DistanceCosine:
		return "<=>"
	case DistanceL2:
		return "<->"
	case DistanceIP:
		return "<#>"
	default:
		return "<=>"
	}
}

// Validate checks if the distance function is valid.
func (d DistanceFunction) Validate() error {
	switch d {
	case DistanceCosine, DistanceL2, DistanceIP:
		return nil
	default:
		return fmt.Errorf("invalid distance function: %s", d)
	}
}

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

	// Vector Search (pgvector) - 保留原有方法以兼容
	SearchChunksByVector(ctx context.Context, kbIDs []string, embedding []float32, limit int) ([]*ChunkWithScore, error)
	// Vector Search (pgvector) - 改进版本，支持更多选项
	SearchChunksByVectorWithOptions(ctx context.Context, kbIDs []string, embedding []float32, limit int, options ...SearchOptions) ([]*ChunkWithScore, error)

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

// SearchOptions 搜索选项
type SearchOptions struct {
	// DistanceFunction 距离函数，默认余弦相似度
	DistanceFunction DistanceFunction
	// ScoreThreshold 分数阈值，只返回分数 >= threshold的结果
	ScoreThreshold *float64
	// WhereClause 自定义 WHERE 条件，例如 "metadata->>'category' = 'tech'"
	WhereClause string
}

// SearchChunksByVector 保留原有签名以兼容现有代码
func (s *knowledgeStore) SearchChunksByVector(ctx context.Context, kbIDs []string, embedding []float32, limit int) ([]*ChunkWithScore, error) {
	return s.SearchChunksByVectorWithOptions(ctx, kbIDs, embedding, limit)
}

// SearchChunksByVectorWithOptions 改进的向量搜索（参考 pgvector 最佳实践）
func (s *knowledgeStore) SearchChunksByVectorWithOptions(ctx context.Context, kbIDs []string, embedding []float32, limit int, options ...SearchOptions) ([]*ChunkWithScore, error) {
	if len(embedding) == 0 {
		return nil, fmt.Errorf("embedding vector is empty")
	}

	if limit <= 0 {
		limit = 5
	}

	// 获取搜索选项
	opts := SearchOptions{
		DistanceFunction: DistanceCosine,
	}
	if len(options) > 0 {
		opts = options[0]
	}

	// 验证距离函数
	if err := opts.DistanceFunction.Validate(); err != nil {
		return nil, fmt.Errorf("invalid distance function: %w", err)
	}

	// 构建查询
	query, args, err := s.buildVectorSearchQuery(kbIDs, embedding, limit, &opts)
	if err != nil {
		return nil, fmt.Errorf("build search query: %w", err)
	}

	// 执行查询
	rows, err := s.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("execute search query: %w", err)
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
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, &ChunkWithScore{
			Chunk: &chunk,
			Score: s.calculateScore(score, opts.DistanceFunction),
		})
	}

	return results, nil
}

// buildVectorSearchQuery 构建向量搜索 SQL
func (s *knowledgeStore) buildVectorSearchQuery(kbIDs []string, queryVector []float32, limit int, opts *SearchOptions) (string, []interface{}, error) {
	// 验证表名（防止 SQL 注入）
	if err := validateIdentifier("knowledge_chunks"); err != nil {
		return "", nil, fmt.Errorf("invalid table name: %w", err)
	}

	op := opts.DistanceFunction.Operator()

	// 构建基础查询
	args := []interface{}{vectorToString(queryVector)}
	query := fmt.Sprintf(`
		SELECT c.id, c.knowledge_base_id, c.document_id, c.chunk_index, c.content,
		       c.content_hash, c.metadata, c.is_enabled, c.created_at, c.updated_at,
		       (e.embedding %s $1::vector) as distance
		FROM knowledge_chunks c
		JOIN embeddings e ON e.chunk_id = c.id
		WHERE c.is_enabled = true`, op)

	// 添加知识库过滤
	if len(kbIDs) > 0 {
		query += " AND c.knowledge_base_id = ANY($2)"
		args = append(args, kbIDs)
	}

	// 添加自定义 WHERE 条件
	if opts.WhereClause != "" {
		// 简单的 SQL 注入检测
		if !isSafeSQL(opts.WhereClause) {
			return "", nil, fmt.Errorf("unsafe where clause: %s", opts.WhereClause)
		}
		query += " AND " + opts.WhereClause
	}

	// 添加分数阈值过滤
	if opts.ScoreThreshold != nil && *opts.ScoreThreshold > 0 {
		thresholdDistance := s.calculateThresholdDistance(*opts.ScoreThreshold, opts.DistanceFunction)
		query += fmt.Sprintf(" AND (e.embedding %s $%d) < %f", op, len(args)+1, thresholdDistance)
		args = append(args, thresholdDistance)
	}

	// 添加排序和限制
	query += fmt.Sprintf(" ORDER BY e.embedding %s $1::vector LIMIT $%d", op, len(args)+1)
	args = append(args, limit)

	return query, args, nil
}

// calculateScore 计算相似度分数
func (s *knowledgeStore) calculateScore(distance float64, distanceFunc DistanceFunction) float64 {
	switch distanceFunc {
	case DistanceCosine:
		// 余弦距离：分数 = 1 - 距离
		return 1 - distance
	case DistanceL2, DistanceIP:
		// L2 和 IP：使用倒数作为分数
		if distance == 0 {
			return 1.0
		}
		return 1.0 / (1.0 + distance)
	default:
		return 1 - distance
	}
}

// calculateThresholdDistance 计算阈值对应的距离值
func (s *knowledgeStore) calculateThresholdDistance(scoreThreshold float64, distanceFunc DistanceFunction) float64 {
	switch distanceFunc {
	case DistanceCosine:
		// 余弦：距离 = 1 - 分数
		return 1 - scoreThreshold
	case DistanceL2, DistanceIP:
		// L2 和 IP：距离已经是正确尺度
		return scoreThreshold
	default:
		return scoreThreshold
	}
}

// validateIdentifier validates SQL identifiers to prevent SQL injection.
// PostgreSQL identifiers must start with a letter or underscore, and contain only letters, digits, and underscores.
func validateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	// Check PostgreSQL naming rules for unquoted identifiers
	for i, c := range name {
		isLetter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		isDigit := c >= '0' && c <= '9'
		isUnderscore := c == '_'

		if i == 0 && !isLetter && c != '_' {
			return fmt.Errorf("identifier must start with a letter or underscore: %s", name)
		}

		if !isLetter && !isDigit && !isUnderscore {
			return fmt.Errorf("identifier contains invalid character: %s", name)
		}
	}

	return nil
}

// quoteIdentifier quotes a PostgreSQL identifier.
func quoteIdentifier(name string) string {
	// Wrap in double quotes to safely use any valid identifier
	return "\"" + name + "\""
}

// isSafeSQL 检查 SQL 片段是否安全（简单的安全检查）
func isSafeSQL(sql string) bool {
	// 检查危险关键字
	dangerous := []string{"DROP", "DELETE", "UPDATE", "INSERT", "ALTER", "CREATE", "EXEC", "EXECUTE"}
	upperSQL := fmt.Sprintf(" %s ", strings.ToUpper(sql))

	for _, keyword := range dangerous {
		if strings.Contains(upperSQL, fmt.Sprintf(" %s ", keyword)) {
			return false
		}
	}
	return true
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

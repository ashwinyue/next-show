// Package knowledge 提供知识库业务逻辑.
package knowledge

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino-ext/components/document/loader/url"
	"github.com/cloudwego/eino-ext/components/document/parser/docx"
	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino-ext/components/document/parser/xlsx"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/model"
)

// DataFilesBaseDir 数据文件存储基础目录.
const DataFilesBaseDir = "data/files"

// SplitterType 分块器类型.
type SplitterType string

const (
	SplitterTypeRecursive SplitterType = "recursive" // 递归分块（按字符/分隔符）
	SplitterTypeSemantic  SplitterType = "semantic"  // 语义分块（按语义相似度）
)

// ImportRequest 文档导入请求.
type ImportRequest struct {
	KnowledgeBaseID string    `json:"knowledge_base_id"`
	Title           string    `json:"title"`
	SourceType      string    `json:"source_type"` // "url", "text", "file"
	SourceURI       string    `json:"source_uri,omitempty"`
	Content         string    `json:"content,omitempty"`
	FileName        string    `json:"-"` // 文件名（用于判断文件类型）
	FileReader      io.Reader `json:"-"` // 文件内容读取器

	// Splitter options
	SplitterType SplitterType `json:"splitter_type,omitempty"` // 分块类型：recursive（默认）或 semantic
	ChunkSize    int          `json:"chunk_size,omitempty"`    // 递归分块的块大小
	ChunkOverlap int          `json:"chunk_overlap,omitempty"` // 递归分块的重叠大小
	Percentile   float64      `json:"percentile,omitempty"`    // 语义分块的百分位阈值（0-1，默认0.9）
}

// ImportResult 文档导入结果.
type ImportResult struct {
	DocumentID string `json:"document_id"`
	ChunkCount int    `json:"chunk_count"`
}

// ImportDocument 导入文档到知识库.
func (b *bizImpl) ImportDocument(ctx context.Context, req *ImportRequest) (*ImportResult, error) {
	docID := uuid.New().String()
	var fileData []byte
	var fileHash string
	var sourceURI string

	// 1. 获取文档内容
	var docs []*schema.Document
	var err error

	switch req.SourceType {
	case "url":
		docs, err = b.loadFromURL(ctx, req.SourceURI)
		sourceURI = req.SourceURI
	case "text":
		docs = []*schema.Document{{Content: req.Content}}
	case "file":
		// 先读取文件内容到内存
		fileData, err = io.ReadAll(req.FileReader)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}
		fileHash = md5HashBytes(fileData)

		// 保存原始文件到本地
		sourceURI, err = b.saveFileToLocal(req.KnowledgeBaseID, docID, req.FileName, fileData)
		if err != nil {
			return nil, fmt.Errorf("save file: %w", err)
		}

		// 解析文件内容
		docs, err = b.parseFile(ctx, req.FileName, bytes.NewReader(fileData))
	default:
		return nil, fmt.Errorf("unsupported source type: %s", req.SourceType)
	}
	if err != nil {
		return nil, fmt.Errorf("load document: %w", err)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no content loaded")
	}

	// 2. 合并所有文档内容
	var contentBuilder strings.Builder
	for _, doc := range docs {
		contentBuilder.WriteString(doc.Content)
		contentBuilder.WriteString("\n")
	}
	fullContent := contentBuilder.String()

	// 3. 创建文档记录
	docModel := &model.KnowledgeDocument{
		ID:              docID,
		KnowledgeBaseID: req.KnowledgeBaseID,
		Title:           req.Title,
		SourceType:      model.DocumentSourceType(req.SourceType),
		SourceURI:       sourceURI,
		FileHash:        fileHash,
		ParseStatus:     model.DocumentParseStatusPending,
	}
	if err := b.store.Knowledge().CreateDocument(ctx, docModel); err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	// 4. 分块
	var chunks []*schema.Document
	switch req.SplitterType {
	case SplitterTypeSemantic:
		// 语义分块
		percentile := req.Percentile
		if percentile <= 0 || percentile > 1 {
			percentile = 0.9
		}
		chunks, err = b.splitDocumentSemantic(ctx, fullContent, percentile)
	default:
		// 递归分块（默认）
		chunkSize := req.ChunkSize
		if chunkSize <= 0 {
			chunkSize = 512
		}
		chunkOverlap := req.ChunkOverlap
		if chunkOverlap <= 0 {
			chunkOverlap = 50
		}
		chunks, err = b.splitDocumentRecursive(ctx, fullContent, chunkSize, chunkOverlap)
	}
	if err != nil {
		return nil, fmt.Errorf("split document: %w", err)
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks after splitting")
	}

	// 5. 生成 embedding
	var chunkContents []string
	for _, c := range chunks {
		chunkContents = append(chunkContents, c.Content)
	}

	embeddingVectors, err := b.embedder.EmbedStrings(ctx, chunkContents)
	if err != nil {
		return nil, fmt.Errorf("embed chunks: %w", err)
	}

	// 6. 创建 chunk 和 embedding 记录
	var chunkModels []*model.KnowledgeChunk
	var embeddingModels []*model.Embedding

	for i, c := range chunks {
		chunkID := uuid.New().String()
		contentHash := md5Hash(c.Content)

		chunkModels = append(chunkModels, &model.KnowledgeChunk{
			ID:              chunkID,
			KnowledgeBaseID: req.KnowledgeBaseID,
			DocumentID:      docID,
			ChunkIndex:      i,
			Content:         c.Content,
			ContentHash:     contentHash,
			IsEnabled:       true,
		})

		if i < len(embeddingVectors) {
			vec32 := make([]float32, len(embeddingVectors[i]))
			for j, v := range embeddingVectors[i] {
				vec32[j] = float32(v)
			}
			embeddingModels = append(embeddingModels, &model.Embedding{
				KnowledgeBaseID: req.KnowledgeBaseID,
				ChunkID:         chunkID,
				Embedding:       vec32,
				EmbeddingDim:    len(vec32),
				EmbeddingModel:  "default",
			})
		}
	}

	// 7. 批量写入
	if err := b.store.Knowledge().CreateChunks(ctx, chunkModels); err != nil {
		return nil, fmt.Errorf("create chunks: %w", err)
	}

	if err := b.store.Knowledge().CreateEmbeddings(ctx, embeddingModels); err != nil {
		return nil, fmt.Errorf("create embeddings: %w", err)
	}

	// 8. 更新文档解析状态
	docModel.ParseStatus = model.DocumentParseStatusParsed
	if err := b.store.Knowledge().UpdateDocument(ctx, docModel); err != nil {
		return nil, fmt.Errorf("update document status: %w", err)
	}

	return &ImportResult{
		DocumentID: docID,
		ChunkCount: len(chunkModels),
	}, nil
}

// loadFromURL 从 URL 加载文档.
func (b *bizImpl) loadFromURL(ctx context.Context, uri string) ([]*schema.Document, error) {
	loader, err := url.NewLoader(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("create url loader: %w", err)
	}

	docs, err := loader.Load(ctx, document.Source{URI: uri})
	if err != nil {
		return nil, fmt.Errorf("load from url: %w", err)
	}

	return docs, nil
}

// splitDocumentRecursive 递归分块文档.
func (b *bizImpl) splitDocumentRecursive(ctx context.Context, content string, chunkSize, chunkOverlap int) ([]*schema.Document, error) {
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   chunkSize,
		OverlapSize: chunkOverlap,
		Separators:  []string{"\n\n", "\n", "。", ".", " ", ""},
	})
	if err != nil {
		return nil, fmt.Errorf("create splitter: %w", err)
	}

	docs := []*schema.Document{{Content: content}}
	chunks, err := splitter.Transform(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("split: %w", err)
	}

	return chunks, nil
}

// splitDocumentSemantic 语义分块文档.
func (b *bizImpl) splitDocumentSemantic(ctx context.Context, content string, percentile float64) ([]*schema.Document, error) {
	if b.embedder == nil {
		return nil, fmt.Errorf("embedder is required for semantic splitting")
	}

	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    b.embedder,
		Percentile:   percentile,
		BufferSize:   1,
		MinChunkSize: 100,
		Separators:   []string{"\n\n", "\n", "。", ".", "?", "!", " "},
	})
	if err != nil {
		return nil, fmt.Errorf("create semantic splitter: %w", err)
	}

	docs := []*schema.Document{{Content: content}}
	chunks, err := splitter.Transform(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("semantic split: %w", err)
	}

	return chunks, nil
}

// parseFile 解析文件内容.
func (b *bizImpl) parseFile(ctx context.Context, fileName string, reader io.Reader) ([]*schema.Document, error) {
	ext := strings.ToLower(filepath.Ext(fileName))

	var p parser.Parser
	var err error

	switch ext {
	case ".pdf":
		p, err = pdf.NewPDFParser(ctx, nil)
	case ".docx":
		p, err = docx.NewDocxParser(ctx, nil)
	case ".xlsx", ".xls":
		p, err = xlsx.NewXlsxParser(ctx, nil)
	case ".csv":
		// CSV 解析为表格文本
		return b.parseCSV(reader)
	case ".txt", ".md":
		// 纯文本直接读取
		content, readErr := io.ReadAll(reader)
		if readErr != nil {
			return nil, fmt.Errorf("read text file: %w", readErr)
		}
		return []*schema.Document{{Content: string(content)}}, nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("create parser for %s: %w", ext, err)
	}

	docs, err := p.Parse(ctx, reader)
	if err != nil {
		return nil, fmt.Errorf("parse %s file: %w", ext, err)
	}

	return docs, nil
}

// parseCSV 解析 CSV 文件为文档.
func (b *bizImpl) parseCSV(reader io.Reader) ([]*schema.Document, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty csv file")
	}

	// 将 CSV 转换为表格文本格式
	var sb strings.Builder
	for i, row := range records {
		if i == 0 {
			sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
			sb.WriteString("|" + strings.Repeat("---|", len(row)) + "\n")
		} else {
			sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
		}
	}

	return []*schema.Document{{Content: sb.String()}}, nil
}

// saveFileToLocal 保存文件到本地存储.
func (b *bizImpl) saveFileToLocal(kbID, docID, fileName string, data []byte) (string, error) {
	// 构建存储路径: data/files/<kbID>/<docID>/<filename>
	dir := filepath.Join(DataFilesBaseDir, kbID, docID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create directory: %w", err)
	}

	filePath := filepath.Join(dir, fileName)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return filePath, nil
}

func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func md5HashBytes(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}

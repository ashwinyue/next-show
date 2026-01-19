// Package knowledge 提供知识库业务逻辑.
package knowledge

import (
	"context"

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
	ListChunks(ctx context.Context, docID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error)
}

// bizImpl 知识库业务实现.
type bizImpl struct {
	store store.Store
}

// NewBiz 创建知识库业务实例.
func NewBiz(s store.Store) Biz {
	return &bizImpl{store: s}
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

func (b *bizImpl) ListChunks(ctx context.Context, docID string, limit, offset int) ([]*model.KnowledgeChunk, int64, error) {
	return b.store.Knowledge().ListChunksByDocument(ctx, docID, limit, offset)
}

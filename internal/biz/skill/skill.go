// Package skill 提供 Skill 业务逻辑层.
package skill

import (
	"context"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz Skill 业务逻辑接口.
type Biz interface {
	// CRUD 操作
	Create(ctx context.Context, skill *model.Skill) (*model.Skill, error)
	Get(ctx context.Context, id string) (*model.Skill, error)
	Update(ctx context.Context, skill *model.Skill) (*model.Skill, error)
	Delete(ctx context.Context, id string) error

	// 列表和搜索
	List(ctx context.Context, page, pageSize int) ([]*model.Skill, int64, error)
	ListAll(ctx context.Context) ([]*model.Skill, error)
	ListByCategory(ctx context.Context, category string) ([]*model.Skill, error)
	ListEnabled(ctx context.Context) ([]*model.Skill, error)
	Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Skill, int64, error)
}

type biz struct {
	store store.Store
}

// NewBiz 创建 Skill 业务层实例.
func NewBiz(store store.Store) Biz {
	return &biz{store: store}
}

func (b *biz) Create(ctx context.Context, skill *model.Skill) (*model.Skill, error) {
	if err := b.store.Skills().Create(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

func (b *biz) Get(ctx context.Context, id string) (*model.Skill, error) {
	return b.store.Skills().Get(ctx, id)
}

func (b *biz) Update(ctx context.Context, skill *model.Skill) (*model.Skill, error) {
	if err := b.store.Skills().Update(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

func (b *biz) Delete(ctx context.Context, id string) error {
	return b.store.Skills().Delete(ctx, id)
}

func (b *biz) List(ctx context.Context, page, pageSize int) ([]*model.Skill, int64, error) {
	offset := (page - 1) * pageSize
	return b.store.Skills().List(ctx, offset, pageSize)
}

func (b *biz) ListAll(ctx context.Context) ([]*model.Skill, error) {
	return b.store.Skills().ListAll(ctx)
}

func (b *biz) ListByCategory(ctx context.Context, category string) ([]*model.Skill, error) {
	return b.store.Skills().ListByCategory(ctx, category)
}

func (b *biz) ListEnabled(ctx context.Context) ([]*model.Skill, error) {
	return b.store.Skills().ListEnabled(ctx)
}

func (b *biz) Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Skill, int64, error) {
	offset := (page - 1) * pageSize
	return b.store.Skills().Search(ctx, keyword, offset, pageSize)
}

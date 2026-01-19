// Package store 提供数据访问层.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-show/internal/model"
)

// CheckpointStore Checkpoint 存储接口（实现 compose.CheckPointStore）.
type CheckpointStore interface {
	Get(ctx context.Context, checkpointID string) (*model.Checkpoint, error)
	Set(ctx context.Context, checkpoint *model.Checkpoint) error
	Delete(ctx context.Context, checkpointID string) error
	UpdateStatus(ctx context.Context, checkpointID string, status model.CheckpointStatus) error
	ListBySession(ctx context.Context, sessionID string) ([]*model.Checkpoint, error)
	ListActive(ctx context.Context) ([]*model.Checkpoint, error)
	CleanExpired(ctx context.Context) (int64, error)
}

type checkpointStore struct {
	db *gorm.DB
}

func newCheckpointStore(db *gorm.DB) CheckpointStore {
	return &checkpointStore{db: db}
}

func (s *checkpointStore) Get(ctx context.Context, checkpointID string) (*model.Checkpoint, error) {
	var checkpoint model.Checkpoint
	if err := s.db.WithContext(ctx).Where("checkpoint_id = ?", checkpointID).First(&checkpoint).Error; err != nil {
		return nil, err
	}
	return &checkpoint, nil
}

func (s *checkpointStore) Set(ctx context.Context, checkpoint *model.Checkpoint) error {
	return s.db.WithContext(ctx).Save(checkpoint).Error
}

func (s *checkpointStore) Delete(ctx context.Context, checkpointID string) error {
	return s.db.WithContext(ctx).Where("checkpoint_id = ?", checkpointID).Delete(&model.Checkpoint{}).Error
}

func (s *checkpointStore) UpdateStatus(ctx context.Context, checkpointID string, status model.CheckpointStatus) error {
	return s.db.WithContext(ctx).Model(&model.Checkpoint{}).Where("checkpoint_id = ?", checkpointID).Update("status", status).Error
}

func (s *checkpointStore) ListBySession(ctx context.Context, sessionID string) ([]*model.Checkpoint, error) {
	var checkpoints []*model.Checkpoint
	if err := s.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at DESC").Find(&checkpoints).Error; err != nil {
		return nil, err
	}
	return checkpoints, nil
}

func (s *checkpointStore) ListActive(ctx context.Context) ([]*model.Checkpoint, error) {
	var checkpoints []*model.Checkpoint
	if err := s.db.WithContext(ctx).Where("status = ?", model.CheckpointStatusActive).Find(&checkpoints).Error; err != nil {
		return nil, err
	}
	return checkpoints, nil
}

func (s *checkpointStore) CleanExpired(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).Where("status = ? AND expires_at < NOW()", model.CheckpointStatusActive).Delete(&model.Checkpoint{})
	return result.RowsAffected, result.Error
}

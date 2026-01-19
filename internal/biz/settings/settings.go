// Package settings 提供系统设置业务逻辑.
package settings

import (
	"context"
	"runtime"

	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// 编译时注入的版本信息
var (
	Version   = "dev"
	CommitID  = "unknown"
	BuildTime = "unknown"
)

// SystemInfo 系统信息.
type SystemInfo struct {
	Version   string `json:"version"`
	CommitID  string `json:"commit_id"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

// Biz 系统设置业务接口.
type Biz interface {
	// GetSystemInfo 获取系统信息.
	GetSystemInfo(ctx context.Context) *SystemInfo
	// List 列出所有设置.
	List(ctx context.Context) ([]*model.SystemSettings, error)
	// ListByCategory 按类别列出设置.
	ListByCategory(ctx context.Context, category string) ([]*model.SystemSettings, error)
	// Get 获取单个设置.
	Get(ctx context.Context, key string) (*model.SystemSettings, error)
	// Set 设置配置项.
	Set(ctx context.Context, req *SetRequest) (*model.SystemSettings, error)
	// Delete 删除设置.
	Delete(ctx context.Context, key string) error
	// GetMultiple 批量获取设置.
	GetMultiple(ctx context.Context, keys []string) (map[string]string, error)
	// SetMultiple 批量设置.
	SetMultiple(ctx context.Context, settings map[string]string) error
}

// SetRequest 设置请求.
type SetRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	ValueType   string `json:"value_type"`
	Category    string `json:"category"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type bizImpl struct {
	store store.Store
}

// NewBiz 创建系统设置业务实例.
func NewBiz(s store.Store) Biz {
	return &bizImpl{store: s}
}

func (b *bizImpl) GetSystemInfo(ctx context.Context) *SystemInfo {
	return &SystemInfo{
		Version:   Version,
		CommitID:  CommitID,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
	}
}

func (b *bizImpl) List(ctx context.Context) ([]*model.SystemSettings, error) {
	return b.store.Settings().List(ctx)
}

func (b *bizImpl) ListByCategory(ctx context.Context, category string) ([]*model.SystemSettings, error) {
	return b.store.Settings().ListByCategory(ctx, category)
}

func (b *bizImpl) Get(ctx context.Context, key string) (*model.SystemSettings, error) {
	return b.store.Settings().Get(ctx, key)
}

func (b *bizImpl) Set(ctx context.Context, req *SetRequest) (*model.SystemSettings, error) {
	// 尝试获取现有设置
	existing, _ := b.store.Settings().Get(ctx, req.Key)

	setting := &model.SystemSettings{
		Key:         req.Key,
		Value:       req.Value,
		ValueType:   req.ValueType,
		Category:    req.Category,
		Label:       req.Label,
		Description: req.Description,
	}

	if existing != nil {
		setting.ID = existing.ID
	} else {
		setting.ID = uuid.New().String()
	}

	if setting.ValueType == "" {
		setting.ValueType = "string"
	}

	if err := b.store.Settings().Set(ctx, setting); err != nil {
		return nil, err
	}

	return setting, nil
}

func (b *bizImpl) Delete(ctx context.Context, key string) error {
	return b.store.Settings().Delete(ctx, key)
}

func (b *bizImpl) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	settings, err := b.store.Settings().GetMultiple(ctx, keys)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

func (b *bizImpl) SetMultiple(ctx context.Context, settings map[string]string) error {
	for key, value := range settings {
		_, err := b.Set(ctx, &SetRequest{
			Key:   key,
			Value: value,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

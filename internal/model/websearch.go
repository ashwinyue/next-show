// Package model 定义数据模型.
package model

import "time"

// WebSearchProvider 网络搜索提供商类型.
type WebSearchProvider string

const (
	WebSearchProviderTavily     WebSearchProvider = "tavily"
	WebSearchProviderBing       WebSearchProvider = "bing"
	WebSearchProviderGoogle     WebSearchProvider = "google"
	WebSearchProviderDuckDuckGo WebSearchProvider = "duckduckgo"
	WebSearchProviderSerper     WebSearchProvider = "serper"
)

// WebSearchConfig 网络搜索配置.
type WebSearchConfig struct {
	ID             string            `json:"id" gorm:"primaryKey;size:36"`
	Name           string            `json:"name" gorm:"uniqueIndex;size:100;not null"`
	DisplayName    string            `json:"display_name" gorm:"size:200;not null"`
	Provider       WebSearchProvider `json:"provider" gorm:"size:50;not null;index"`
	APIKey         string            `json:"-" gorm:"size:500"`
	BaseURL        string            `json:"base_url" gorm:"size:500"`
	MaxResults     int               `json:"max_results" gorm:"default:10"`
	SearchDepth    string            `json:"search_depth" gorm:"size:20;default:basic"` // basic, advanced
	IncludeDomains JSONSlice         `json:"include_domains" gorm:"type:jsonb"`
	ExcludeDomains JSONSlice         `json:"exclude_domains" gorm:"type:jsonb"`
	Config         JSONMap           `json:"config" gorm:"type:jsonb"`
	IsEnabled      bool              `json:"is_enabled" gorm:"default:true;index"`
	IsDefault      bool              `json:"is_default" gorm:"default:false;index"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func (WebSearchConfig) TableName() string {
	return "web_search_configs"
}

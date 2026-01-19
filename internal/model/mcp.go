// Package model 定义数据模型.
package model

import "time"

// TransportType MCP 传输类型.
type TransportType string

const (
	TransportTypeStdio     TransportType = "stdio"
	TransportTypeSSE       TransportType = "sse"
	TransportTypeWebSocket TransportType = "websocket"
)

// MCPServer MCP 服务器配置.
type MCPServer struct {
	ID             string        `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name           string        `json:"name" gorm:"size:100;not null;uniqueIndex"`
	DisplayName    string        `json:"display_name" gorm:"size:200;not null"`
	Description    string        `json:"description,omitempty" gorm:"type:text"`
	TransportType  TransportType `json:"transport_type" gorm:"size:20;not null;default:stdio;index"`
	Command        string        `json:"command,omitempty" gorm:"size:500"`
	Args           JSONMap       `json:"args,omitempty" gorm:"type:jsonb"`
	Env            JSONMap       `json:"env,omitempty" gorm:"type:jsonb"`
	ServerURL      string        `json:"server_url,omitempty" gorm:"size:500"`
	CustomHeaders  JSONMap       `json:"custom_headers,omitempty" gorm:"type:jsonb"`
	TimeoutSeconds int           `json:"timeout_seconds" gorm:"default:30"`
	IsEnabled      bool          `json:"is_enabled" gorm:"default:true;index"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`

	// 关联
	Tools []MCPTool `json:"tools,omitempty" gorm:"foreignKey:MCPServerID"`
}

func (MCPServer) TableName() string {
	return "mcp_servers"
}

// MCPTool MCP 工具配置.
type MCPTool struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MCPServerID    string    `json:"mcp_server_id" gorm:"type:uuid;not null;index"`
	Name           string    `json:"name" gorm:"size:200;not null"`
	DisplayName    string    `json:"display_name,omitempty" gorm:"size:300"`
	Description    string    `json:"description,omitempty" gorm:"type:text"`
	InputSchema    JSONMap   `json:"input_schema,omitempty" gorm:"type:jsonb"`
	ReturnDirectly bool      `json:"return_directly" gorm:"default:false"`
	IsEnabled      bool      `json:"is_enabled" gorm:"default:true;index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联
	MCPServer *MCPServer `json:"mcp_server,omitempty" gorm:"foreignKey:MCPServerID"`
}

func (MCPTool) TableName() string {
	return "mcp_tools"
}

// ToolType 工具类型.
type ToolType string

const (
	ToolTypeMCP     ToolType = "mcp"
	ToolTypeBuiltin ToolType = "builtin"
	ToolTypeCustom  ToolType = "custom"
)

// AgentTool Agent 工具关联.
type AgentTool struct {
	ID               string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AgentID          string    `json:"agent_id" gorm:"type:uuid;not null;index"`
	ToolType         ToolType  `json:"tool_type" gorm:"size:20;not null;index"`
	MCPToolID        *string   `json:"mcp_tool_id,omitempty" gorm:"type:uuid"`
	BuiltinToolName  string    `json:"builtin_tool_name,omitempty" gorm:"size:200"`
	CustomToolConfig JSONMap   `json:"custom_tool_config,omitempty" gorm:"type:jsonb"`
	ReturnDirectly   bool      `json:"return_directly" gorm:"default:false"`
	IsEnabled        bool      `json:"is_enabled" gorm:"default:true;index"`
	Priority         int       `json:"priority" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// 关联
	Agent   *Agent   `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
	MCPTool *MCPTool `json:"mcp_tool,omitempty" gorm:"foreignKey:MCPToolID"`
}

func (AgentTool) TableName() string {
	return "agent_tools"
}

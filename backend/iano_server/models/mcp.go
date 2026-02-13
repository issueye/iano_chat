package models

// MCPTransportType MCP 传输类型
type MCPTransportType string

const (
	MCPTransportStdio MCPTransportType = "stdio" // 标准输入输出
	MCPTransportSSE   MCPTransportType = "sse"   // Server-Sent Events
	MCPTransportHTTP  MCPTransportType = "http"  // HTTP
)

// MCPServerStatus MCP 服务器状态
type MCPServerStatus string

const (
	MCPServerStatusConnected    MCPServerStatus = "connected"    // 已连接
	MCPServerStatusDisconnected MCPServerStatus = "disconnected" // 未连接
	MCPServerStatusError        MCPServerStatus = "error"        // 错误
)

// MCPServer MCPServer 模型，用于存储 MCP 服务器配置信息
type MCPServer struct {
	BaseModel
	Name         string           `gorm:"column:name;size:255;not null" json:"name"`                  // 服务器名称
	Desc         string           `gorm:"column:desc;type:text" json:"desc"`                          // 描述
	Transport    MCPTransportType `gorm:"column:transport;size:20;not null" json:"transport"`         // 传输类型: stdio, sse, http
	Command      string           `gorm:"column:command;size:500" json:"command"`                     // 命令 (用于 stdio)
	Args         string           `gorm:"column:args;type:text" json:"args"`                          // 命令参数 (JSON 数组)
	Env          string           `gorm:"column:env;type:text" json:"env"`                            // 环境变量 (JSON 对象)
	URL          string           `gorm:"column:url;size:500" json:"url"`                             // URL (用于 sse/http)
	Enabled      bool             `gorm:"column:enabled;default:true" json:"enabled"`                 // 是否启用
	Status       MCPServerStatus  `gorm:"column:status;size:20;default:'disconnected'" json:"status"` // 连接状态
	Version      string           `gorm:"column:version;size:50;default:'1.0.0'" json:"version"`      // 版本
	Author       string           `gorm:"column:author;size:255" json:"author"`                       // 作者
	Icon         string           `gorm:"column:icon;size:255" json:"icon"`                           // 图标
	Capabilities string           `gorm:"column:capabilities;type:text" json:"capabilities"`          // 能力 (JSON)
	LastError    string           `gorm:"column:last_error;type:text" json:"last_error"`              // 最后错误信息
	ToolsCount   int64            `gorm:"column:tools_count;default:0" json:"tools_count"`            // 工具数量
}

// TableName 返回表名
func (m *MCPServer) TableName() string {
	return "mcp_servers"
}

// MCPServerTool MCP 服务器提供的工具
type MCPServerTool struct {
	BaseModel
	ServerID    string `gorm:"column:server_id;size:36;not null;index" json:"server_id"` // 服务器 ID
	Name        string `gorm:"column:name;size:255;not null" json:"name"`                // 工具名称
	Description string `gorm:"column:description;type:text" json:"description"`          // 工具描述
	InputSchema string `gorm:"column:input_schema;type:text" json:"input_schema"`        // 输入模式 (JSON)
}

// TableName 返回表名
func (m *MCPServerTool) TableName() string {
	return "mcp_server_tools"
}

package models

import "encoding/json"

// ToolType 工具类型
type ToolType string

const (
	ToolTypeBuiltin  ToolType = "builtin"  // 内置工具
	ToolTypeCustom   ToolType = "custom"   // 自定义工具
	ToolTypeExternal ToolType = "external" // 外部工具（API 调用）
	ToolTypePlugin   ToolType = "plugin"   // 插件工具
	ToolTypeScript   ToolType = "script"   // 脚本工具（JavaScript）
)

// ToolStatus 工具状态
type ToolStatus string

const (
	ToolStatusEnabled  ToolStatus = "enabled"  // 启用
	ToolStatusDisabled ToolStatus = "disabled" // 禁用
	ToolStatusError    ToolStatus = "error"    // 错误
)

// ToolParameter 工具参数定义
type ToolParameter struct {
	Name     string      `json:"name"`              // 参数名
	Type     string      `json:"type"`              // 参数类型：string, number, boolean, array, object
	Desc     string      `json:"desc"`              // 参数描述
	Required bool        `json:"required"`          // 是否必需
	Default  interface{} `json:"default,omitempty"` // 默认值
	Enum     []string    `json:"enum,omitempty"`    // 枚举值
}

// ToolDefinition 工具定义
type ToolDefinition struct {
	Name       string          `json:"name" binding:"required"` // 工具名称
	Type       ToolType        `json:"type"`                    // 工具类型
	Desc       string          `json:"desc"`                    // 工具描述
	Parameters []ToolParameter `json:"parameters"`              // 参数定义
	Returns    string          `json:"returns"`                 // 返回值描述
	Example    string          `json:"example,omitempty"`       // 使用示例
}

type Tool struct {
	BaseModel
	Name          string     `gorm:"column:name;size:255;not null" json:"name"`
	Desc          string     `gorm:"column:desc;type:text" json:"desc"`
	Returns       string     `gorm:"column:returns;type:text" json:"returns"`           // 返回值描述
	Example       string     `gorm:"column:example;type:text" json:"example,omitempty"` // 使用示例
	Type          ToolType   `gorm:"column:type;size:20" json:"type"`
	Status        ToolStatus `gorm:"column:status;size:20;default:'enabled'" json:"status"`
	ScriptContent string     `gorm:"column:script_content;type:text" json:"script_content,omitempty"` // 脚本内容（仅对脚本工具）
	CallCount     int64      `gorm:"column:call_count;default:0" json:"call_count"`                   // 调用次数
	ErrorCount    int64      `gorm:"column:error_count;default:0" json:"error_count"`                 // 错误次数
	Config        string     `gorm:"column:config;type:text" json:"config,omitempty"`                 // 工具配置（JSON）
	Parameters    string     `gorm:"column:parameters;type:text" json:"parameters,omitempty"`         // 参数定义（JSON）
	Version       string     `gorm:"column:version;default:1.0.0" json:"version"`                     // 版本
	Author        string     `gorm:"column:author" json:"author"`                                     // 作者
}

func (table *Tool) TableName() string {
	return "tools"
}

// ToDefinition 转换为工具定义
func (table *Tool) ToDefinition() *ToolDefinition {
	return &ToolDefinition{
		Name:       table.Name,
		Type:       table.Type,
		Desc:       table.Desc,
		Parameters: table.GetParameters(),
		Returns:    table.Returns,
		Example:    table.Example,
	}
}

// GetParameters 获取参数定义
func (table *Tool) GetParameters() []ToolParameter {
	if table.Parameters == "" {
		return nil
	}
	var params []ToolParameter
	json.Unmarshal([]byte(table.Parameters), &params)
	return params
}

// CommandConfig 命令执行工具配置
type CommandConfig struct {
	AllowedCommands []string `json:"allowed_commands"` // 允许执行的命令列表
	Timeout         int      `json:"timeout"`          // 超时时间（秒）
	WorkingDir      string   `json:"working_dir"`      // 工作目录
	Shell           string   `json:"shell"`            // shell 类型: powershell, cmd, bash
}

// GetCommandConfig 获取命令配置
func (table *Tool) GetCommandConfig() *CommandConfig {
	if table.Config == "" {
		return nil
	}
	var config CommandConfig
	if err := json.Unmarshal([]byte(table.Config), &config); err != nil {
		return nil
	}
	return &config
}

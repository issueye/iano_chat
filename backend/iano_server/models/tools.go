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
	Name       string     `gorm:"size:255;not null" json:"name"`
	Desc       string     `gorm:"type:text" json:"desc"`
	Returns    string     `gorm:"type:text" json:"returns"`           // 返回值描述
	Example    string     `gorm:"type:text" json:"example,omitempty"` // 使用示例
	Type       ToolType   `gorm:"size:20" json:"type"`
	Status     ToolStatus `gorm:"size:20;default:'enabled'" json:"status"`
	CallCount  int64      `json:"call_count" gorm:"default:0"`  // 调用次数
	ErrorCount int64      `json:"error_count" gorm:"default:0"` // 错误次数
	Config     string     `json:"config" gorm:"type:text"`      // 工具配置（JSON）
	Parameters string     `json:"parameters" gorm:"type:text"`  // 参数定义（JSON）
	Version    string     `json:"version" gorm:"default:1.0.0"` // 版本
	Author     string     `json:"author"`                       // 作者
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

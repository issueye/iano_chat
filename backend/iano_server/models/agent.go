package models

// AgentType Agent 类型
type AgentType string

const (
	AgentTypeMain   AgentType = "main"
	AgentTypeSub    AgentType = "sub"
	AgentTypeCustom AgentType = "custom"
)

// Agent Agent 模型
type Agent struct {
	BaseModel
	Name         string    `gorm:"column:name" json:"name"`
	Description  string    `gorm:"column:description" json:"description"`
	Type         AgentType `gorm:"size:20;default:'main'" json:"type"`
	IsSubAgent   bool      `gorm:"default:false" json:"is_sub_agent"`
	ProviderID   string    `gorm:"column:provider_id" json:"provider_id"`
	Model        string    `gorm:"column:model" json:"model"`
	Instructions string    `gorm:"column:instructions;type:text" json:"instructions"`
	Tools        string    `gorm:"column:tools;type:text" json:"tools"`
	MCPServerIDs StrArray  `gorm:"column:mcp_server_ids;type:text" json:"mcp_server_ids"`
}

func (Agent) TableName() string {
	return "agents"
}

// IsMainAgent 是否为主 Agent
func (a *Agent) IsMainAgent() bool {
	return a.Type == AgentTypeMain && !a.IsSubAgent
}

// IsSubAgent 判断是否为 SubAgent
func (a *Agent) IsSubAgentType() bool {
	return a.Type == AgentTypeSub || a.IsSubAgent
}

// IsCustomAgent 是否为自定义 Agent
func (a *Agent) IsCustomAgent() bool {
	return a.Type == AgentTypeCustom
}

// GetInstructions 获取指令
func (a *Agent) GetInstructions() string {
	return a.Instructions
}

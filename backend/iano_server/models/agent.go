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
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Type         AgentType `gorm:"size:20;default:'main'" json:"type"`
	IsSubAgent   bool      `gorm:"default:false" json:"is_sub_agent"`
	ProviderID   string    `json:"provider_id"`
	Model        string    `json:"model"`
	Instructions string    `gorm:"type:text" json:"instructions"`
	Tools        string    `gorm:"type:text" json:"tools"`
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

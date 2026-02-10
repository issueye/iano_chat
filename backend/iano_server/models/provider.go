package models

import "time"

// 供应商
type Provider struct {
	BaseModel
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	BaseUrl     string    `json:"base_url"`
	ApiKey      string    `json:"api_key"`
	Model       string    `json:"model"`
	Temperature float32   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

func (table *Provider) TableName() string {
	return "providers"
}

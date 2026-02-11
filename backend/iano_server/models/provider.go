package models

type Provider struct {
	BaseModel
	Name        string  `json:"name"`
	BaseUrl     string  `json:"base_url"`
	ApiKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	IsDefault   bool    `gorm:"default:false" json:"is_default"`
}

func (table *Provider) TableName() string {
	return "providers"
}

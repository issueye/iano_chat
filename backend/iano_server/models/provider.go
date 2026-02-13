package models

type Provider struct {
	BaseModel
	Name        string  `gorm:"column:name;size:255;not null" json:"name"`
	BaseUrl     string  `gorm:"column:base_url;size:255;not null" json:"base_url"`
	ApiKey      string  `gorm:"column:api_key;size:255;not null" json:"api_key"`
	Model       string  `gorm:"column:model;size:255;not null" json:"model"`
	Temperature float32 `gorm:"column:temperature;default:0.7" json:"temperature"`
	MaxTokens   int     `gorm:"column:max_tokens;default:2048" json:"max_tokens"`
	IsDefault   bool    `gorm:"column:is_default;default:false" json:"is_default"`
}

func (table *Provider) TableName() string {
	return "providers"
}

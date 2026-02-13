package models

// APIKey API 密钥模型
type APIKey struct {
	BaseModel
	Desc string `gorm:"column:desc;size:255;not null" json:"desc"` // 描述
	Key  string `gorm:"column:key;size:255;not null" json:"key"`   // API 密钥
}

func (APIKey) TableName() string {
	return "api_keys"
}

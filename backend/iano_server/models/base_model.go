package models

import (
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;" json:"created_at"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;" json:"updated_at"` // 更新时间
}

// NewID 生成新的 ID
func (b *BaseModel) NewID() {
	b.ID = uuid.New().String()
}

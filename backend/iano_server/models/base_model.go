package models

import (
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at" json:"updated_at"`
}

// NewID 生成新的 ID
func (b *BaseModel) NewID() {
	b.ID = uuid.New().String()
}

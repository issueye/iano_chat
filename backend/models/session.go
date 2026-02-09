package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// SessionStatus 会话状态
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusPaused    SessionStatus = "paused"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusArchived  SessionStatus = "archived"
)

// SessionConfig 会话配置
type SessionConfig struct {
	ModelID         int64    `json:"model_id"`
	Temperature     float32  `json:"temperature"`
	MaxTokens       int      `json:"max_tokens"`
	SystemPrompt    string   `json:"system_prompt"`
	EnableTools     bool     `json:"enable_tools"`
	SelectedTools   []string `json:"selected_tools"`
	EnableSummary   bool     `json:"enable_summary"`
	KeepRounds      int      `json:"keep_rounds"`
	EnableRateLimit bool     `json:"enable_rate_limit"`
	RateLimitRPM    int      `json:"rate_limit_rpm"`
}

// DefaultSessionConfig 返回默认配置
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		Temperature:     0.7,
		MaxTokens:       2000,
		EnableTools:     true,
		SelectedTools:   []string{},
		EnableSummary:   true,
		KeepRounds:      4,
		EnableRateLimit: true,
		RateLimitRPM:    20,
	}
}

// Session 会话模型
type Session struct {
	BaseModel
	UserID       int64          `gorm:"index;not null" json:"user_id"`
	Title        string         `gorm:"size:255;not null;default:'新会话'" json:"title"`
	Status       SessionStatus  `gorm:"size:20;default:'active'" json:"status"`
	ConfigJSON   string         `gorm:"type:text" json:"-"`
	MessageCount int            `gorm:"default:0" json:"message_count"`
	TotalTokens  int            `gorm:"default:0" json:"total_tokens"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	LastActiveAt time.Time      `json:"last_active_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 返回表名
func (Session) TableName() string {
	return "sessions"
}

// GetConfig 获取配置对象
func (s *Session) GetConfig() (*SessionConfig, error) {
	if s.ConfigJSON == "" {
		return DefaultSessionConfig(), nil
	}
	var config SessionConfig
	if err := json.Unmarshal([]byte(s.ConfigJSON), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetConfig 设置配置对象
func (s *Session) SetConfig(config *SessionConfig) error {
	if config == nil {
		config = DefaultSessionConfig()
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	s.ConfigJSON = string(data)
	return nil
}

// BeforeCreate 创建前钩子
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.Status == "" {
		s.Status = SessionStatusActive
	}
	if s.Title == "" {
		s.Title = "新会话"
	}
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	if s.LastActiveAt.IsZero() {
		s.LastActiveAt = now
	}
	// 设置默认配置
	if s.ConfigJSON == "" {
		config := DefaultSessionConfig()
		data, _ := json.Marshal(config)
		s.ConfigJSON = string(data)
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (s *Session) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}

// SessionSummary 会话摘要（用于列表展示）
type SessionSummary struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	MessageCount int       `json:"message_count"`
	LastActiveAt time.Time `json:"last_active_at"`
	Preview      string    `json:"preview"`
}

// ToSummary 转换为摘要
func (s *Session) ToSummary(preview string) *SessionSummary {
	return &SessionSummary{
		Title:        s.Title,
		Status:       string(s.Status),
		MessageCount: s.MessageCount,
		LastActiveAt: s.LastActiveAt,
		Preview:      preview,
	}
}

// IsActive 是否处于活跃状态
func (s *Session) IsActive() bool {
	return s.Status == SessionStatusActive
}

// CanChat 是否可以进行对话
func (s *Session) CanChat() bool {
	return s.Status == SessionStatusActive || s.Status == SessionStatusPaused
}

// UpdateLastActive 更新最后活跃时间
func (s *Session) UpdateLastActive() {
	s.LastActiveAt = time.Now()
}

// IncrementMessageCount 增加消息计数
func (s *Session) IncrementMessageCount() {
	s.MessageCount++
}

// AddTokens 增加 Token 计数
func (s *Session) AddTokens(tokens int) {
	s.TotalTokens += tokens
}

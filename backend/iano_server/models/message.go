package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeSystem    MessageType = "system"
	MessageTypeTool      MessageType = "tool"
)

// MessageStatus 消息状态
type MessageStatus string

const (
	MessageStatusSending   MessageStatus = "sending"
	MessageStatusCompleted MessageStatus = "completed"
	MessageStatusFailed    MessageStatus = "failed"
	MessageStatusStreaming MessageStatus = "streaming"
)

// FeedbackRating 反馈评分
type FeedbackRating string

const (
	FeedbackRatingLike    FeedbackRating = "like"
	FeedbackRatingDislike FeedbackRating = "dislike"
)

// ToolCall 工具调用
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Attachment 附件
type Attachment struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// MessageContent 消息内容
type MessageContent struct {
	Text             string       `json:"text"`
	ToolCalls        []ToolCall   `json:"tool_calls,omitempty"`
	ReasoningContent string       `json:"reasoning_content,omitempty"`
	Attachments      []Attachment `json:"attachments,omitempty"`
}

// Message 消息模型
type Message struct {
	BaseModel
	SessionID       string          `gorm:"not null" json:"session_id"`
	Type            MessageType     `gorm:"size:20;not null" json:"type"`
	Content         string          `gorm:"type:text;not null" json:"content"`
	Status          MessageStatus   `gorm:"size:20;default:'completed'" json:"status"`
	InputTokens     int             `gorm:"default:0" json:"input_tokens"`
	OutputTokens    int             `gorm:"default:0" json:"output_tokens"`
	ParentID        *string         `gorm:"size:36;index" json:"parent_id,omitempty"`
	FeedbackRating  *FeedbackRating `gorm:"size:10" json:"feedback_rating,omitempty"`
	FeedbackComment string          `gorm:"type:text" json:"feedback_comment,omitempty"`
	FeedbackAt      *time.Time      `json:"feedback_at,omitempty"`
}

// TableName 返回表名
func (Message) TableName() string {
	return "messages"
}

// BeforeCreate 创建前钩子
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	// 调用 BaseModel 的 NewID 生成 UUID
	if m.ID == "" {
		m.BaseModel.NewID()
	}
	if m.Status == "" {
		m.Status = MessageStatusCompleted
	}
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Message) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

// GetContent 获取消息内容对象
func (m *Message) GetContent() (*MessageContent, error) {
	if m.Content == "" {
		return &MessageContent{}, nil
	}
	var content MessageContent
	if err := json.Unmarshal([]byte(m.Content), &content); err != nil {
		return nil, err
	}
	return &content, nil
}

// SetContent 设置消息内容对象
func (m *Message) SetContent(content *MessageContent) error {
	if content == nil {
		content = &MessageContent{}
	}
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}
	m.Content = string(data)
	return nil
}

// SetTextContent 设置纯文本内容
func (m *Message) SetTextContent(text string) error {
	return m.SetContent(&MessageContent{Text: text})
}

// GetTextContent 获取纯文本内容
func (m *Message) GetTextContent() string {
	content, err := m.GetContent()
	if err != nil {
		return ""
	}
	return content.Text
}

// IsUserMessage 是否为用户消息
func (m *Message) IsUserMessage() bool {
	return m.Type == MessageTypeUser
}

// IsAssistantMessage 是否为 AI 消息
func (m *Message) IsAssistantMessage() bool {
	return m.Type == MessageTypeAssistant
}

// IsSystemMessage 是否为系统消息
func (m *Message) IsSystemMessage() bool {
	return m.Type == MessageTypeSystem
}

// IsToolMessage 是否为工具消息
func (m *Message) IsToolMessage() bool {
	return m.Type == MessageTypeTool
}

// IsStreaming 是否正在流式输出
func (m *Message) IsStreaming() bool {
	return m.Status == MessageStatusStreaming
}

// IsCompleted 是否已完成
func (m *Message) IsCompleted() bool {
	return m.Status == MessageStatusCompleted
}

// IsFailed 是否失败
func (m *Message) IsFailed() bool {
	return m.Status == MessageStatusFailed
}

// GetTotalTokens 获取总 Token 数
func (m *Message) GetTotalTokens() int {
	return m.InputTokens + m.OutputTokens
}

// AddFeedback 添加反馈
func (m *Message) AddFeedback(rating FeedbackRating, comment string) {
	now := time.Now()
	m.FeedbackRating = &rating
	m.FeedbackComment = comment
	m.FeedbackAt = &now
}

// HasFeedback 是否有反馈
func (m *Message) HasFeedback() bool {
	return m.FeedbackRating != nil
}

// MessageHistory 消息历史（用于 Agent 输入）
type MessageHistory struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToHistory 转换为历史记录格式
func (m *Message) ToHistory() *MessageHistory {
	role := ""
	switch m.Type {
	case MessageTypeUser:
		role = "user"
	case MessageTypeAssistant:
		role = "assistant"
	case MessageTypeSystem:
		role = "system"
	case MessageTypeTool:
		role = "tool"
	}
	return &MessageHistory{
		Role:    role,
		Content: m.GetTextContent(),
	}
}

// CreateUserMessage 创建用户消息
func CreateUserMessage(sessionID string, text string) *Message {
	msg := &Message{
		SessionID: sessionID,
		Type:      MessageTypeUser,
		Status:    MessageStatusCompleted,
	}
	msg.NewID()
	msg.SetTextContent(text)
	return msg
}

// CreateAssistantMessage 创建助手消息
func CreateAssistantMessage(sessionID string, status MessageStatus) *Message {
	msg := &Message{
		SessionID: sessionID,
		Type:      MessageTypeAssistant,
		Status:    status,
	}
	msg.NewID()
	return msg
}

// CreateSystemMessage 创建系统消息
func CreateSystemMessage(sessionID string, text string) *Message {
	msg := &Message{
		SessionID: sessionID,
		Type:      MessageTypeSystem,
		Status:    MessageStatusCompleted,
	}
	msg.NewID()
	msg.SetTextContent(text)
	return msg
}

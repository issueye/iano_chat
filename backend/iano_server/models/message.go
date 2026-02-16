package models

import (
	"encoding/json"
	"time"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeSystem    MessageType = "system"
	MessageTypeTool      MessageType = "tool"
)

func (t MessageType) ToString() string {
	return string(t)
}

// MessageStatus 消息状态
type MessageStatus string

const (
	MessageStatusSending   MessageStatus = "sending"
	MessageStatusCompleted MessageStatus = "completed"
	MessageStatusFailed    MessageStatus = "failed"
	MessageStatusStreaming MessageStatus = "streaming"
)

func (s MessageStatus) ToString() string {
	return string(s)
}

// FeedbackRating 反馈评分
type FeedbackRating string

const (
	FeedbackRatingLike    FeedbackRating = "like"
	FeedbackRatingDislike FeedbackRating = "dislike"
)

type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type ContentBlock struct {
	Type     string    `json:"type"`                // "text" 或 "tool_call"
	Text     string    `json:"text,omitempty"`      // 当 type 为 "text" 时
	ToolCall *ToolCall `json:"tool_call,omitempty"` // 当 type 为 "tool_call" 时
}

type Attachment struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

type MessageContent struct {
	Blocks           []ContentBlock `json:"blocks,omitempty"`
	Text             string         `json:"text,omitempty"`
	ToolCalls        []ToolCall     `json:"tool_calls,omitempty"`
	ReasoningContent string         `json:"reasoning_content,omitempty"`
	ThinkContent     string         `json:"think_content,omitempty"`
	IsThink          bool           `json:"is_think,omitempty"`
	Attachments      []Attachment   `json:"attachments,omitempty"`
}

// Message 消息模型
type Message struct {
	BaseModel
	SessionID       string          `gorm:"column:session_id;not null" json:"session_id"`
	Type            MessageType     `gorm:"column:type;size:20;not null" json:"type"`
	Content         string          `gorm:"column:content;type:text;not null" json:"content"`
	Status          MessageStatus   `gorm:"column:status;size:20;default:'completed'" json:"status"`
	InputTokens     int             `gorm:"column:input_tokens;default:0" json:"input_tokens"`
	OutputTokens    int             `gorm:"column:output_tokens;default:0" json:"output_tokens"`
	ParentID        *string         `gorm:"column:parent_id;size:36;index" json:"parent_id,omitempty"`
	FeedbackRating  *FeedbackRating `gorm:"column:feedback_rating;size:10" json:"feedback_rating,omitempty"`
	FeedbackComment string          `gorm:"column:feedback_comment;type:text" json:"feedback_comment,omitempty"`
	FeedbackAt      *JSONTime       `gorm:"column:feedback_at" json:"feedback_at,omitempty"`
}

// TableName 返回表名
func (Message) TableName() string {
	return "messages"
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

// AddToolCall 添加工具调用
func (m *Message) AddToolCall(toolCall ToolCall) error {
	content, err := m.GetContent()
	if err != nil {
		return err
	}
	content.ToolCalls = append(content.ToolCalls, toolCall)
	return m.SetContent(content)
}

// GetText 获取文本内容
func (m *Message) GetText() string {
	content, err := m.GetContent()
	if err != nil {
		return m.Content
	}
	return content.Text
}

// SetText 设置文本内容
func (m *Message) SetText(text string) error {
	content, err := m.GetContent()
	if err != nil {
		return err
	}
	content.Text = text
	return m.SetContent(content)
}

// AddAttachment 添加附件
func (m *Message) AddAttachment(attachment Attachment) error {
	content, err := m.GetContent()
	if err != nil {
		return err
	}
	content.Attachments = append(content.Attachments, attachment)
	return m.SetContent(content)
}

// AddFeedback 添加反馈
func (m *Message) AddFeedback(rating FeedbackRating, comment string) {
	now := PtrJSONTime(time.Now())
	m.FeedbackRating = &rating
	m.FeedbackComment = comment
	m.FeedbackAt = now
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

// CreateUserMessage 创建用户消息
func CreateUserMessage(sessionID string, text string) *Message {
	msg := &Message{
		SessionID: sessionID,
		Type:      MessageTypeUser,
		Status:    MessageStatusCompleted,
	}
	msg.NewID()
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

type MessageEvent string

const (
	MessageEventCreated   MessageEvent = "message_created"   // 消息创建事件
	MessageEventContent   MessageEvent = "message_content"   // 助手消息内容事件
	MessageEventCompleted MessageEvent = "message_completed" // 消息完成事件
	MessageEventError     MessageEvent = "error"             // 错误事件
	MessageEventDone      MessageEvent = "done"              // 会话完成事件
)

func (e MessageEvent) ToString() string {
	return string(e)
}

type ErrCompleted struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

func CreateErrCompleted(id string, status MessageStatus, err string) *ErrCompleted {
	return &ErrCompleted{
		ID:     id,
		Status: status.ToString(),
		Error:  err,
	}
}

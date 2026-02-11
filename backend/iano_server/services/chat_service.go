package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	iano "iano_agent"
	"iano_agent/store"
	"iano_server/models"

	"github.com/cloudwego/eino/components/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ChatService struct {
	db          *gorm.DB
	coordinator *iano.Coordinator
	redis       *redis.Client
	mu          sync.RWMutex
}

type ChatServiceConfig struct {
	MaxIdleTime     int
	CleanupInterval int
	RedisEnabled    bool
}

func DefaultChatServiceConfig() *ChatServiceConfig {
	return &ChatServiceConfig{
		MaxIdleTime:     30,
		CleanupInterval: 5,
		RedisEnabled:    false,
	}
}

func NewChatService(db *gorm.DB, chatModel model.ToolCallingChatModel, redisClient *redis.Client, cfg *ChatServiceConfig) *ChatService {
	if cfg == nil {
		cfg = DefaultChatServiceConfig()
	}

	var st store.ConversationStore
	if redisClient != nil && cfg.RedisEnabled {
		st = store.NewRedisStore(redisClient, nil)
	} else {
		st = store.NewMemoryStore()
	}

	poolConfig := &iano.AgentPoolConfig{
		MaxIdleTime:     DurationFromMinutes(cfg.MaxIdleTime),
		CleanupInterval: DurationFromMinutes(cfg.CleanupInterval),
	}

	coordinator := iano.NewCoordinator(chatModel, st, poolConfig)

	return &ChatService{
		db:          db,
		coordinator: coordinator,
		redis:       redisClient,
	}
}

func DurationFromMinutes(minutes int) time.Duration {
	return time.Duration(minutes) * time.Minute
}

type ChatRequest struct {
	SessionID string
	AgentID   string
	Message   string
	Provider  *models.Provider
	Agent     *models.Agent
}

type ChatResponse struct {
	Content    string
	TokenUsage *iano.TokenUsage
	Duration   time.Duration
}

func (s *ChatService) Chat(ctx context.Context, req *ChatRequest, callback iano.MessageCallback) (*ChatResponse, error) {
	result, err := s.coordinator.Chat(ctx, req.SessionID, req.AgentID, req.Message, callback)
	if err != nil {
		slog.Error("Chat failed", "error", err, "sessionID", req.SessionID)
		return nil, err
	}

	return &ChatResponse{
		Content:    result.Response,
		TokenUsage: &iano.TokenUsage{TotalTokens: result.TokenUsed},
		Duration:   result.Duration,
	}, nil
}

func (s *ChatService) buildAgentOptions(agent *models.Agent) []iano.Option {
	opts := []iano.Option{}

	if agent != nil {
		if agent.Instructions != "" {
			opts = append(opts, iano.WithSystemPrompt(agent.Instructions))
		}
		if agent.Tools != "" {
			opts = append(opts, iano.WithAllowedTools(parseTools(agent.Tools)))
		}
	}

	return opts
}

func parseTools(toolsStr string) []string {
	if toolsStr == "" {
		return nil
	}
	return []string{toolsStr}
}

func (s *ChatService) ClearSession(ctx context.Context, sessionID string) error {
	return s.coordinator.ClearSession(ctx, sessionID)
}

func (s *ChatService) GetPoolStats() map[string]interface{} {
	return s.coordinator.GetPoolStats()
}

func (s *ChatService) GetConversationInfo(sessionID string, agentID string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"sessionId": sessionID,
		"agentId":   agentID,
	}, nil
}

func (s *ChatService) Close() {
	s.coordinator.Close()
}

func (s *ChatService) SaveMessage(message *models.Message) error {
	return s.db.Create(message).Error
}

func (s *ChatService) GetMessagesBySessionID(sessionID string) ([]models.Message, error) {
	var messages []models.Message
	if err := s.db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

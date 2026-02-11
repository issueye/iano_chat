package agent

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudwego/eino/schema"
)

type ConversationStore interface {
	Save(ctx context.Context, sessionID string, layer *ConversationLayer) error
	Load(ctx context.Context, sessionID string) (*ConversationLayer, error)
	Delete(ctx context.Context, sessionID string) error
	Exists(ctx context.Context, sessionID string) (bool, error)
}

type ConversationData struct {
	RecentRounds     []RoundData `json:"recent_rounds"`
	SummaryContent   string      `json:"summary_content"`
	SummarizedRounds int         `json:"summarized_rounds"`
	SavedAt          time.Time   `json:"saved_at"`
}

type RoundData struct {
	UserContent      string    `json:"user_content"`
	AssistantContent string    `json:"assistant_content"`
	Timestamp        time.Time `json:"timestamp"`
	TokenCount       int       `json:"token_count"`
}

func LayerToData(layer *ConversationLayer) *ConversationData {
	rounds := make([]RoundData, 0, len(layer.RecentRounds))
	for _, r := range layer.RecentRounds {
		userContent := ""
		assistantContent := ""
		if r.UserMessage != nil {
			userContent = r.UserMessage.Content
		}
		if r.AssistantMessage != nil {
			assistantContent = r.AssistantMessage.Content
		}
		rounds = append(rounds, RoundData{
			UserContent:      userContent,
			AssistantContent: assistantContent,
			Timestamp:        r.Timestamp,
			TokenCount:       r.TokenCount,
		})
	}

	return &ConversationData{
		RecentRounds:     rounds,
		SummaryContent:   layer.SummaryContent,
		SummarizedRounds: layer.SummarizedRounds,
		SavedAt:          time.Now(),
	}
}

func DataToLayer(data *ConversationData) *ConversationLayer {
	rounds := make([]*ConversationRound, 0, len(data.RecentRounds))
	for _, r := range data.RecentRounds {
		rounds = append(rounds, &ConversationRound{
			UserMessage:      newUserMessage(r.UserContent),
			AssistantMessage: newAssistantMessage(r.AssistantContent),
			Timestamp:        r.Timestamp,
			TokenCount:       r.TokenCount,
		})
	}

	return &ConversationLayer{
		RecentRounds:     rounds,
		SummaryContent:   data.SummaryContent,
		SummarizedRounds: data.SummarizedRounds,
	}
}

func newUserMessage(content string) *schema.Message {
	if content == "" {
		return nil
	}
	return schema.UserMessage(content)
}

func newAssistantMessage(content string) *schema.Message {
	if content == "" {
		return nil
	}
	return schema.AssistantMessage(content, nil)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (cd *ConversationData) ToJSON() ([]byte, error) {
	return json.Marshal(cd)
}

func ConversationDataFromJSON(data []byte) (*ConversationData, error) {
	var cd ConversationData
	if err := json.Unmarshal(data, &cd); err != nil {
		return nil, err
	}
	return &cd, nil
}

type MemoryStore struct {
	data map[string]*ConversationLayer
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]*ConversationLayer),
	}
}

func (s *MemoryStore) Save(ctx context.Context, sessionID string, layer *ConversationLayer) error {
	s.data[sessionID] = layer
	return nil
}

func (s *MemoryStore) Load(ctx context.Context, sessionID string) (*ConversationLayer, error) {
	layer, ok := s.data[sessionID]
	if !ok {
		return nil, nil
	}
	return layer, nil
}

func (s *MemoryStore) Delete(ctx context.Context, sessionID string) error {
	delete(s.data, sessionID)
	return nil
}

func (s *MemoryStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	_, ok := s.data[sessionID]
	return ok, nil
}

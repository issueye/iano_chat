package agent

import (
	"context"
	"iano_server/models"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

type OpenAiChatModel struct {
	chatModel model.ToolCallingChatModel
}

func NewOpenAiChatModel(ctx context.Context, p *models.Provider) (*OpenAiChatModel, error) {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     p.BaseUrl,
		APIKey:      p.ApiKey,
		Model:       p.Model,
		Temperature: &p.Temperature,
		MaxTokens:   &p.MaxTokens,
	})
	if err != nil {
		return nil, err
	}

	rtn := &OpenAiChatModel{
		chatModel: chatModel,
	}

	return rtn, nil
}

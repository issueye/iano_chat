package model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// init 注册 OpenAI 工厂
func init() {
	GlobalRegistry.Register(ProviderOpenAI, &OpenAIFactory{})
}

// OpenAIFactory OpenAI 模型工厂
type OpenAIFactory struct{}

// Create 创建 OpenAI 模型实例
func (f *OpenAIFactory) Create(config *Config) (model.ToolCallingChatModel, error) {
	ctx := context.Background()

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     config.BaseURL,
		APIKey:      config.APIKey,
		Model:       config.Model,
		Temperature: &config.Temperature,
		MaxTokens:   &config.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 OpenAI 模型失败: %w", err)
	}

	return chatModel, nil
}

// Support 是否支持该提供商
func (f *OpenAIFactory) Support(providerType ProviderType) bool {
	return providerType == ProviderOpenAI
}

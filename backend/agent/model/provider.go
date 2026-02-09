package model

import (
	"fmt"

	"github.com/cloudwego/eino/components/model"
)

// ProviderType 模型提供商类型
type ProviderType string

const (
	ProviderOpenAI   ProviderType = "openai"
	ProviderClaude   ProviderType = "claude"
	ProviderGemini   ProviderType = "gemini"
	ProviderOllama   ProviderType = "ollama"
	ProviderAzure    ProviderType = "azure"
	ProviderDeepSeek ProviderType = "deepseek"
)

// Config 模型配置
type Config struct {
	// 提供商类型
	Type ProviderType
	// API 基础 URL
	BaseURL string
	// API Key
	APIKey string
	// 模型名称
	Model string
	// 温度参数
	Temperature float32
	// 最大 Token 数
	MaxTokens int
	// 其他自定义参数
	Extra map[string]interface{}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("模型提供商类型不能为空")
	}
	if c.Model == "" {
		return fmt.Errorf("模型名称不能为空")
	}
	if c.APIKey == "" && c.Type != ProviderOllama {
		return fmt.Errorf("API Key 不能为空")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("BaseURL 不能为空")
	}
	return nil
}

// Factory 模型工厂接口
type Factory interface {
	// Create 创建模型实例
	Create(config *Config) (model.ToolCallingChatModel, error)
	// Support 是否支持该提供商
	Support(providerType ProviderType) bool
}

// Registry 模型工厂注册表
type Registry struct {
	factories map[ProviderType]Factory
}

// NewRegistry 创建模型工厂注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[ProviderType]Factory),
	}
}

// Register 注册工厂
func (r *Registry) Register(providerType ProviderType, factory Factory) {
	r.factories[providerType] = factory
}

// Get 获取工厂
func (r *Registry) Get(providerType ProviderType) (Factory, bool) {
	f, ok := r.factories[providerType]
	return f, ok
}

// GlobalRegistry 全局模型工厂注册表
var GlobalRegistry = NewRegistry()

// CreateModel 创建模型实例
func CreateModel(config *Config) (model.ToolCallingChatModel, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	factory, ok := GlobalRegistry.Get(config.Type)
	if !ok {
		return nil, fmt.Errorf("不支持的模型提供商: %s", config.Type)
	}

	return factory.Create(config)
}

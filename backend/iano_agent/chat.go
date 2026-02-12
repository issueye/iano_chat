package iano_agent

import (
	"context"
	"errors"
	"fmt"
	"iano_agent/callback"
	"io"
	"log/slog"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

func (a *Agent) Chat(ctx context.Context, userInput string, cb MessageCallback) (string, error) {
	userMsg := schema.UserMessage(userInput)
	messages := []*schema.Message{userMsg}

	return a.Loop(ctx, messages, cb)
}

func (a *Agent) ChatSimple(ctx context.Context, userInput string, cb MessageCallback) (string, error) {
	return a.Chat(ctx, userInput, cb)
}

func (a *Agent) invokeTool(ctx context.Context, name string, arguments string) (string, error) {
	// 查找工具
	tool, isFind := a.toolRegistry.Get(name)
	if !isFind {
		return "", fmt.Errorf("工具 %s 不存在", name)
	}

	// 调用工具
	result, err := tool.InvokableRun(ctx, arguments)
	if err != nil {
		return "", fmt.Errorf("工具调用失败: %w", err)
	}
	return result, nil
}

func (a *Agent) ChatWithHistory(ctx context.Context, messages []*schema.Message, cb MessageCallback) (string, error) {
	return a.Loop(ctx, messages, cb)
}

func (a *Agent) Loop(ctx context.Context, messages []*schema.Message, cb MessageCallback) (string, error) {
	loopMessage := make([]*schema.Message, 0)
	for _, msg := range messages {
		loopMessage = append(loopMessage, &schema.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	fullResponse := ""

	// 构建流式对话选项
	opts := a.MakeStreamOpts()
	msgReader, err := a.ra.Stream(ctx, loopMessage, opts...)
	if err != nil {
		return "", fmt.Errorf("流式对话失败: %w", err)
	}

	for {
		msg, err := msgReader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info("流式对话结束")
				break
			}
			slog.Error("读取消息失败", slog.String("error", err.Error()))
			return "", fmt.Errorf("流式对话接收消息失败: %w", err)
		}

		// 处理消息内容
		if msg.Content != "" {
			fullResponse += msg.Content

			// 添加到对话历史
			loopMessage = append(loopMessage, &schema.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})

			// 调用回调函数
			if cb != nil {
				cb(msg.Content, len(msg.ToolCalls) > 0, nil)
			}
		}

		// 处理工具
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				// 调用工具
				toolResult, err := a.invokeTool(ctx, tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					slog.Error("工具调用失败", "id", tc.ID, "name", tc.Function.Name, "arguments", tc.Function.Arguments, "error", err.Error())
					return "", fmt.Errorf("工具调用失败: %w", err)
				}
				slog.Info("工具调用成功", "id", tc.ID, "name", tc.Function.Name, "arguments", tc.Function.Arguments, "result", toolResult)

				// 将工具调用结果添加到对话历史
				loopMessage = append(loopMessage, &schema.Message{
					Role:    schema.Tool,
					Content: fmt.Sprintf("工具调用结果: %s", toolResult),
				})
				fullResponse += fmt.Sprintf("\n工具调用结果: %s", toolResult)
			}
		}
	}

	return fullResponse, nil
}

func (a *Agent) MakeStreamOpts() []agent.AgentOption {
	return []agent.AgentOption{
		agent.WithComposeOptions(compose.WithCallbacks(&callback.LogCallbackHandler{})),
	}
}

func (a *Agent) AddTool(name string, t tool.InvokableTool) error {
	return a.AddToolToRegistry(name, t)
}

func (a *Agent) AddToolToRegistry(name string, t tool.InvokableTool) error {
	if name == "" {
		return fmt.Errorf("工具名称不能为空")
	}
	if t == nil {
		return fmt.Errorf("工具实例不能为空")
	}

	if err := a.toolRegistry.Register(name, t); err != nil {
		return fmt.Errorf("注册工具失败: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	toolsConfig, err := a.makeToolsConfig()
	if err != nil {
		return fmt.Errorf("创建工具配置失败: %w", err)
	}

	ctx := context.Background()
	ra, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: a.chatModel,
		ToolsConfig:      toolsConfig,
		MaxStep:          30,
	})
	if err != nil {
		return fmt.Errorf("创建代理失败: %w", err)
	}

	a.ra = ra
	return nil
}

func (a *Agent) RemoveTool(name string) error {
	if err := a.toolRegistry.Unregister(name); err != nil {
		return fmt.Errorf("注销工具失败: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	toolsConfig, err := a.makeToolsConfig()
	if err != nil {
		return fmt.Errorf("创建工具配置失败: %w", err)
	}

	ctx := context.Background()
	ra, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: a.chatModel,
		ToolsConfig:      toolsConfig,
		MaxStep:          30,
	})
	if err != nil {
		return fmt.Errorf("创建代理失败: %w", err)
	}

	a.ra = ra
	return nil
}

func (a *Agent) ListTools() []string {
	return a.toolRegistry.Names()
}

func (a *Agent) SetCallback(callback MessageCallback) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.Callback = callback
}

func (a *Agent) GetConfig() *Config {
	a.mu.RLock()
	defer a.mu.RUnlock()

	cfg := *a.config
	return &cfg
}

func (a *Agent) GetTokenUsage() *TokenUsage {
	a.mu.RLock()
	defer a.mu.RUnlock()

	usage := *a.tokenUsage
	return &usage
}

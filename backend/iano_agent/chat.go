package iano_agent

import (
	"context"
	"errors"
	"fmt"
	"iano_agent/callback"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

const (
	ThinkStartTag        = "<think>"            // 思考开始标签
	ThinkEndTag          = "</think>"           // 思考结束标签
	ThinkUnicodeStartTag = "\u003cthink\u003e"  // 思考开始标签的 Unicode 编码
	ThinkUnicodeEndTag   = "\u003c/think\u003e" // 思考结束标签的 Unicode 编码
)

func (a *Agent) Chat(ctx context.Context, userInput string) (string, error) {
	userMsg := schema.UserMessage(userInput)
	messages := []*schema.Message{userMsg}

	return a.Loop(ctx, messages)
}

func (a *Agent) ChatSimple(ctx context.Context, userInput string) (string, error) {
	return a.Chat(ctx, userInput)
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

func (a *Agent) ChatWithHistory(ctx context.Context, messages []*schema.Message) (string, error) {
	return a.Loop(ctx, messages)
}

func (a *Agent) WaitForResponse() {
	// 等待回调
	for !a.IsDone {
		time.Sleep(time.Second)
	}
}

func (a *Agent) Loop(ctx context.Context, messages []*schema.Message) (string, error) {
	loopMessage := make([]*schema.Message, 0)
	for _, msg := range messages {
		loopMessage = append(loopMessage, &schema.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	fullResponse := ""
	maxIterations := 30
	iteration := 0

	for iteration < maxIterations {
		iteration++

		opts := a.MakeStreamOpts()
		msgReader, err := a.ra.Stream(ctx, loopMessage, opts...)
		if err != nil {
			return "", fmt.Errorf("流式对话失败: %w", err)
		}

		hasToolCalls := false

		for {
			msg, err := msgReader.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					slog.Info("流式对话结束", "iteration", iteration)
					break
				}
				slog.Error("读取消息失败", slog.String("error", err.Error()))
				return "", fmt.Errorf("流式对话接收消息失败: %w", err)
			}

			if msg.Content != "" || msg.ReasoningContent != "" {
				// 处理思考或者推理
				a.AiThinkStart(msg)

				fullResponse += msg.Content + msg.ReasoningContent

				loopMessage = append(loopMessage, &schema.Message{
					Role:             msg.Role,
					Content:          msg.Content,
					ReasoningContent: msg.ReasoningContent,
				})

				//消息回调
				a.InvokeMsgCB(msg)

				// 处理思考结束
				a.AiThinkEnd(msg)
			}

			if len(msg.ToolCalls) > 0 {
				hasToolCalls = true
				for _, tc := range msg.ToolCalls {
					toolResult, err := a.invokeTool(ctx, tc.Function.Name, tc.Function.Arguments)
					if err != nil {
						slog.Error("工具调用失败", "id", tc.ID, "name", tc.Function.Name, "arguments", tc.Function.Arguments, "error", err.Error())
						toolResult = fmt.Sprintf("工具调用错误: %s", err.Error())
					}
					slog.Info("工具调用完成", "id", tc.ID, "name", tc.Function.Name, "arguments", tc.Function.Arguments, "result", toolResult)

					loopMessage = append(loopMessage, &schema.Message{
						Role:    schema.Tool,
						Content: fmt.Sprintf("工具调用结果: %s", toolResult),
					})
					fullResponse += fmt.Sprintf("\n工具调用结果: %s", toolResult)

					// 工具回调
					a.InvokeToolCB(tc, toolResult)
				}
			}
		}

		if !hasToolCalls {
			break
		}

		slog.Info("工具调用完成，继续对话循环", "iteration", iteration)
	}

	a.IsDone = true

	return fullResponse, nil
}

func (a *Agent) InvokeToolCB(tc schema.ToolCall, callToolError string) {
	if a.CBs == nil {
		return
	}

	info := ToolCallInfo{
		ID:        tc.ID,
		Name:      tc.Function.Name,
		Arguments: tc.Function.Arguments,
	}

	msgStruct := &Message{
		Role:             "tool",
		Content:          "",
		ReasoningContent: "",
		ThinkContent:     "",
		ToolCall:         &info,
		IsThink:          false,
		IsReasoning:      false,
		IsToolCall:       true,
		CallToolError:    callToolError,
	}

	for _, cb := range a.CBs {
		cb(msgStruct)
	}
}

type SplitMessage struct {
	Content string
	IsThink bool
}

func (a *Agent) InvokeMsgCB(msg *schema.Message) {
	if a.CBs == nil {
		return
	}

	if msg.Content == "" && msg.ReasoningContent == "" {
		return
	}

	// thinkContent := ""
	// if a.IsThink {
	// 	thinkContent = msg.Content
	// }

	msgs := []*SplitMessage{}

	// 如果是 think 的话，就需要将内容拆分为思考内容和正常内容
	if a.IsThink {
		// 应该直接回答这个问题。\n\u003c/think\u003e\n\n我是 **MiniMax-M2.5**，一个由 MiniMax 公司开发的 AI 助手。
		if strings.Contains(msg.Content, ThinkEndTag) {
			arr := strings.Split(msg.Content, ThinkEndTag)
			for _, item := range arr {
				msgs = append(msgs, &SplitMessage{
					Content: item,
					IsThink: true,
				})
			}
		}
	}

	msgs = append(msgs, &SplitMessage{
		Content: msg.Content,
		IsThink: a.IsThink,
	})

	for _, c := range msgs {
		msgStruct := &Message{
			Role:             string(msg.Role),
			Content:          msg.Content,
			ReasoningContent: msg.ReasoningContent,
			ThinkContent:     c.Content,
			ToolCall:         nil,
			IsThink:          c.IsThink,
			IsReasoning:      a.IsReasoning,
			IsToolCall:       false,
		}

		for _, cb := range a.CBs {
			cb(msgStruct)
		}
	}
}

// IsStartThink 判断是否是开始思考
func (a *Agent) AiThinkStart(msg *schema.Message) {
	if strings.Contains(msg.Content, ThinkStartTag) {
		a.IsThink = true
	}

	if msg.ReasoningContent != "" {
		a.IsReasoning = true
	}
}

// IsStartThink 判断是否是开始思考
func (a *Agent) AiThinkEnd(msg *schema.Message) {
	if strings.Contains(msg.Content, ThinkEndTag) {
		a.IsThink = false
	}

	if msg.ReasoningContent == "" {
		a.IsReasoning = false
	}
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

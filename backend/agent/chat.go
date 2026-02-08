package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type Agent struct {
	BaseUrl string
	Key     string
	Model   string

	// Agent
	ra        *react.Agent
	chatModel model.BaseChatModel
}

func NewAgent(chatModel model.ToolCallingChatModel) (*Agent, error) {
	ctx := context.Background()

	agent := &Agent{}
	// 创建 Tools
	tools, err := agent.MakeTools()
	if err != nil {
		return nil, err
	}

	// 创建 agent
	ra, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      tools,

		// 消息修改器
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			return input
		},
	})

	agent.ra = ra
	return agent, nil
}

func (a *Agent) MakeTools() (compose.ToolsNodeConfig, error) {
	return compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{},
	}, nil
}

func (a *Agent) Chat(ctx context.Context, input []*schema.Message) error {
	msgReader, err := a.ra.Stream(ctx, input)
	if err != nil {
		return err
	}

	// 读取消息
	for {
		// 读取消息
		msg, err := msgReader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// 读取完成
				break
			}
			// error
			slog.Error("读取消息失败", slog.String("error", err.Error()))
			return err
		}

		fmt.Print(msg.Content)
	}

	return nil
}

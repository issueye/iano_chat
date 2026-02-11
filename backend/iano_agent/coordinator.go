package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
)

type Coordinator struct {
	pool        *AgentPool
	store       ConversationStore
	mainAgentID string
	mu          sync.RWMutex
}

func NewCoordinator(chatModel model.ToolCallingChatModel, store ConversationStore, poolConfig *AgentPoolConfig) *Coordinator {
	pool := NewAgentPool(chatModel, poolConfig)
	return &Coordinator{
		pool:  pool,
		store: store,
	}
}

type TaskResult struct {
	AgentID   string
	Response  string
	TokenUsed int64
	Duration  time.Duration
	Error     error
}

func (c *Coordinator) Chat(ctx context.Context, sessionID string, agentID string, userInput string, opts ...Option) (*TaskResult, error) {
	start := time.Now()

	agent, err := c.pool.Get(sessionID, agentID, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取 Agent 失败: %w", err)
	}

	if c.store != nil {
		layer, err := c.store.Load(ctx, sessionID)
		if err != nil {
			slog.Warn("加载对话历史失败", "error", err)
		} else if layer != nil {
			agent.RestoreConversation(layer)
		}
	}

	response, err := agent.Chat(ctx, userInput)
	if err != nil {
		return &TaskResult{
			AgentID:  agentID,
			Error:    err,
			Duration: time.Since(start),
		}, err
	}

	if c.store != nil {
		if err := c.store.Save(ctx, sessionID, agent.GetConversationLayer()); err != nil {
			slog.Warn("保存对话历史失败", "error", err)
		}
	}

	usage := agent.GetTokenUsage()

	return &TaskResult{
		AgentID:   agentID,
		Response:  response,
		TokenUsed: usage.TotalTokens,
		Duration:  time.Since(start),
	}, nil
}

func (c *Coordinator) ChatWithCallback(ctx context.Context, sessionID string, agentID string, userInput string, callback MessageCallback, opts ...Option) (*TaskResult, error) {
	opts = append(opts, WithCallback(callback))
	return c.Chat(ctx, sessionID, agentID, userInput, opts...)
}

func (c *Coordinator) ClearSession(ctx context.Context, sessionID string) error {
	agent, err := c.pool.Get(sessionID, "")
	if err == nil {
		agent.ClearHistory()
	}

	if c.store != nil {
		if err := c.store.Delete(ctx, sessionID); err != nil {
			return fmt.Errorf("删除会话数据失败: %w", err)
		}
	}

	c.pool.Remove(sessionID, "")
	return nil
}

func (c *Coordinator) GetAgent(sessionID string, agentID string, opts ...Option) (*Agent, error) {
	return c.pool.Get(sessionID, agentID, opts...)
}

func (c *Coordinator) GetPoolStats() map[string]interface{} {
	return c.pool.Stats()
}

func (c *Coordinator) Close() {
	c.pool.Close()
}

type MultiAgentTask struct {
	AgentID string
	Input   string
}

func (c *Coordinator) ParallelChat(ctx context.Context, sessionID string, tasks []MultiAgentTask, opts ...Option) []*TaskResult {
	results := make([]*TaskResult, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t MultiAgentTask) {
			defer wg.Done()
			result, err := c.Chat(ctx, sessionID+":"+t.AgentID, t.AgentID, t.Input, opts...)
			if err != nil {
				results[idx] = &TaskResult{
					AgentID: t.AgentID,
					Error:   err,
				}
				return
			}
			results[idx] = result
		}(i, task)
	}

	wg.Wait()
	return results
}

func (c *Coordinator) Broadcast(ctx context.Context, sessionID string, agentIDs []string, input string, opts ...Option) []*TaskResult {
	tasks := make([]MultiAgentTask, len(agentIDs))
	for i, agentID := range agentIDs {
		tasks[i] = MultiAgentTask{
			AgentID: agentID,
			Input:   input,
		}
	}
	return c.ParallelChat(ctx, sessionID, tasks, opts...)
}

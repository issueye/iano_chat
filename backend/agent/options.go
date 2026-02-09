package agent

// Option 配置选项函数类型
type Option func(*Config)

// WithTools 设置工具列表
func WithTools(tools []Tool) Option {
	return func(c *Config) {
		c.Tools = tools
	}
}

// WithCallback 设置消息回调
func WithCallback(callback MessageCallback) Option {
	return func(c *Config) {
		c.Callback = callback
	}
}

// WithSummaryConfig 设置摘要配置
func WithSummaryConfig(cfg SummaryConfig) Option {
	return func(c *Config) {
		c.Summary = cfg
	}
}

// WithMaxRounds 设置最大对话轮数
func WithMaxRounds(maxRounds int) Option {
	return func(c *Config) {
		if maxRounds > 0 {
			c.MaxRounds = maxRounds
		}
	}
}

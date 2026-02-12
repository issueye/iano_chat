package iano_agent

type Option func(*Config)

func WithTools(tools []Tool) Option {
	return func(c *Config) {
		c.Tools = tools
	}
}

func WithCallback(callback MessageCallback) Option {
	return func(c *Config) {
		c.Callback = callback
	}
}

func WithMaxRounds(maxRounds int) Option {
	return func(c *Config) {
		if maxRounds > 0 {
			c.MaxRounds = maxRounds
		}
	}
}

func WithAllowedTools(allowedTools []string) Option {
	return func(c *Config) {
		c.AllowedTools = allowedTools
	}
}

func WithSystemPrompt(prompt string) Option {
	return func(c *Config) {
		c.SystemPrompt = prompt
	}
}

// WithWorkDir 设置 Agent 的工作目录，限制文件操作工具的操作范围
func WithWorkDir(workDir string) Option {
	return func(c *Config) {
		c.WorkDir = workDir
	}
}

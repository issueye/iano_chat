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

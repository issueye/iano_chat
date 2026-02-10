package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics 指标收集器
type Metrics struct {
	// 对话相关指标
	ChatTotal    *prometheus.CounterVec
	ChatDuration *prometheus.HistogramVec
	ChatTokens   *prometheus.CounterVec
	ChatRounds   *prometheus.GaugeVec

	// 工具相关指标
	ToolCallsTotal   *prometheus.CounterVec
	ToolCallDuration *prometheus.HistogramVec
	ToolCallErrors   *prometheus.CounterVec

	// 摘要相关指标
	SummaryTotal    prometheus.Counter
	SummaryTokens   *prometheus.CounterVec
	SummaryDuration *prometheus.HistogramVec

	// 限流相关指标
	RateLimitHits   prometheus.Counter
	RateLimitMisses prometheus.Counter

	// 错误相关指标
	ErrorsTotal *prometheus.CounterVec
}

// NewMetrics 创建指标收集器
func NewMetrics(namespace string) *Metrics {
	if namespace == "" {
		namespace = "iano_chat"
	}

	return &Metrics{
		// 对话总数
		ChatTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "chat_total",
			Help:      "对话总次数",
		}, []string{"status"}), // status: success, error

		// 对话耗时
		ChatDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "chat_duration_seconds",
			Help:      "对话耗时（秒）",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		}, []string{"status"}),

		// Token 使用量
		ChatTokens: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "chat_tokens_total",
			Help:      "Token 使用总量",
		}, []string{"type"}), // type: prompt, completion, total

		// 对话轮数
		ChatRounds: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "chat_rounds",
			Help:      "当前对话轮数",
		}, []string{"conversation_id"}),

		// 工具调用总数
		ToolCallsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "tool_calls_total",
			Help:      "工具调用总次数",
		}, []string{"tool_name", "status"}),

		// 工具调用耗时
		ToolCallDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "tool_call_duration_seconds",
			Help:      "工具调用耗时（秒）",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		}, []string{"tool_name"}),

		// 工具调用错误数
		ToolCallErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "tool_call_errors_total",
			Help:      "工具调用错误次数",
		}, []string{"tool_name", "error_type"}),

		// 摘要总数
		SummaryTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "summary_total",
			Help:      "摘要总次数",
		}),

		// 摘要 Token 数
		SummaryTokens: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "summary_tokens_total",
			Help:      "摘要 Token 数量",
		}, []string{"type"}), // type: old, summary, saved

		// 摘要耗时
		SummaryDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "summary_duration_seconds",
			Help:      "摘要耗时（秒）",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
		}, []string{}),

		// 限流命中次数
		RateLimitHits: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "rate_limit_hits_total",
			Help:      "限流命中次数",
		}),

		// 限流未命中次数
		RateLimitMisses: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "rate_limit_misses_total",
			Help:      "限流未命中次数",
		}),

		// 错误总数
		ErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "agent",
			Name:      "errors_total",
			Help:      "错误总次数",
		}, []string{"error_code"}),
	}
}

// GlobalMetrics 全局指标收集器
var GlobalMetrics = NewMetrics("")

// RecordChat 记录对话指标
func (m *Metrics) RecordChat(duration time.Duration, status string) {
	m.ChatTotal.WithLabelValues(status).Inc()
	m.ChatDuration.WithLabelValues(status).Observe(duration.Seconds())
}

// RecordChatTokens 记录 Token 使用量
func (m *Metrics) RecordChatTokens(promptTokens, completionTokens int) {
	m.ChatTokens.WithLabelValues("prompt").Add(float64(promptTokens))
	m.ChatTokens.WithLabelValues("completion").Add(float64(completionTokens))
	m.ChatTokens.WithLabelValues("total").Add(float64(promptTokens + completionTokens))
}

// SetChatRounds 设置对话轮数
func (m *Metrics) SetChatRounds(conversationID string, rounds int) {
	m.ChatRounds.WithLabelValues(conversationID).Set(float64(rounds))
}

// RecordToolCall 记录工具调用指标
func (m *Metrics) RecordToolCall(toolName string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
		m.ToolCallErrors.WithLabelValues(toolName, "execution").Inc()
	}

	m.ToolCallsTotal.WithLabelValues(toolName, status).Inc()
	m.ToolCallDuration.WithLabelValues(toolName).Observe(duration.Seconds())
}

// RecordSummary 记录摘要指标
func (m *Metrics) RecordSummary(duration time.Duration, oldTokens, summaryTokens, savedTokens int) {
	m.SummaryTotal.Inc()
	m.SummaryDuration.WithLabelValues().Observe(duration.Seconds())
	m.SummaryTokens.WithLabelValues("old").Add(float64(oldTokens))
	m.SummaryTokens.WithLabelValues("summary").Add(float64(summaryTokens))
	m.SummaryTokens.WithLabelValues("saved").Add(float64(savedTokens))
}

// RecordRateLimit 记录限流指标
func (m *Metrics) RecordRateLimit(hit bool) {
	if hit {
		m.RateLimitHits.Inc()
	} else {
		m.RateLimitMisses.Inc()
	}
}

// RecordError 记录错误指标
func (m *Metrics) RecordError(errorCode string) {
	m.ErrorsTotal.WithLabelValues(errorCode).Inc()
}

// Timer 计时器辅助类型
type Timer struct {
	start time.Time
}

// NewTimer 创建新的计时器
func NewTimer() *Timer {
	return &Timer{start: time.Now()}
}

// Elapsed 获取经过的时间
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.start)
}

// ContextKey 上下文键类型
type ContextKey string

const (
	// MetricsContextKey 指标上下文键
	MetricsContextKey ContextKey = "metrics"
)

// WithMetrics 将指标收集器添加到上下文
func WithMetrics(ctx context.Context, metrics *Metrics) context.Context {
	return context.WithValue(ctx, MetricsContextKey, metrics)
}

// GetMetrics 从上下文获取指标收集器
func GetMetrics(ctx context.Context) *Metrics {
	if metrics, ok := ctx.Value(MetricsContextKey).(*Metrics); ok {
		return metrics
	}
	return GlobalMetrics
}

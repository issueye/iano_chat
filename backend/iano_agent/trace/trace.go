package trace

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer 链路追踪器
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer 创建新的追踪器
func NewTracer(name string) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(name),
	}
}

// GlobalTracer 全局追踪器
var GlobalTracer = NewTracer("iano-chat-agent")

// SpanContext Span 上下文
type SpanContext struct {
	context.Context
	span trace.Span
}

// StartSpan 开始一个新的 Span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) *SpanContext {
	ctx, span := t.tracer.Start(ctx, name, opts...)
	return &SpanContext{
		Context: ctx,
		span:    span,
	}
}

// End 结束 Span
func (sc *SpanContext) End() {
	if sc.span != nil {
		sc.span.End()
	}
}

// SetError 设置错误信息
func (sc *SpanContext) SetError(err error) {
	if sc.span != nil && err != nil {
		sc.span.RecordError(err)
		sc.span.SetStatus(codes.Error, err.Error())
	}
}

// SetAttribute 设置属性
func (sc *SpanContext) SetAttribute(key string, value interface{}) {
	if sc.span == nil {
		return
	}

	switch v := value.(type) {
	case string:
		sc.span.SetAttributes(attribute.String(key, v))
	case int:
		sc.span.SetAttributes(attribute.Int(key, v))
	case int64:
		sc.span.SetAttributes(attribute.Int64(key, v))
	case float64:
		sc.span.SetAttributes(attribute.Float64(key, v))
	case bool:
		sc.span.SetAttributes(attribute.Bool(key, v))
	default:
		sc.span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}

// SetAttributes 批量设置属性
func (sc *SpanContext) SetAttributes(attrs map[string]interface{}) {
	for k, v := range attrs {
		sc.SetAttribute(k, v)
	}
}

// AddEvent 添加事件
func (sc *SpanContext) AddEvent(name string, attrs ...attribute.KeyValue) {
	if sc.span != nil {
		sc.span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SpanKind Span 类型
type SpanKind int

const (
	SpanKindInternal SpanKind = iota
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

// toTraceSpanKind 转换为 OpenTelemetry SpanKind
func toTraceSpanKind(kind SpanKind) trace.SpanKind {
	switch kind {
	case SpanKindServer:
		return trace.SpanKindServer
	case SpanKindClient:
		return trace.SpanKindClient
	case SpanKindProducer:
		return trace.SpanKindProducer
	case SpanKindConsumer:
		return trace.SpanKindConsumer
	default:
		return trace.SpanKindInternal
	}
}

// StartSpanWithKind 指定类型开始 Span
func (t *Tracer) StartSpanWithKind(ctx context.Context, name string, kind SpanKind) *SpanContext {
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(toTraceSpanKind(kind)),
	}
	return t.StartSpan(ctx, name, opts...)
}

// TraceFunc 包装函数执行，自动创建 Span
func (t *Tracer) TraceFunc(ctx context.Context, name string, fn func(context.Context) error) error {
	spanCtx := t.StartSpan(ctx, name)
	defer spanCtx.End()

	err := fn(spanCtx.Context)
	if err != nil {
		spanCtx.SetError(err)
	}
	return err
}

// TraceFuncWithResult 包装带返回值的函数执行
func TraceFuncWithResult[T any](ctx context.Context, tracer *Tracer, name string, fn func(context.Context) (T, error)) (T, error) {
	spanCtx := tracer.StartSpan(ctx, name)
	defer spanCtx.End()

	result, err := fn(spanCtx.Context)
	if err != nil {
		spanCtx.SetError(err)
	}
	return result, err
}

// ChatSpanAttributes 对话 Span 属性
type ChatSpanAttributes struct {
	Model       string
	InputLength int
	HasTools    bool
	ToolCount   int
}

// Apply 应用属性到 Span
func (a *ChatSpanAttributes) Apply(sc *SpanContext) {
	sc.SetAttribute("chat.model", a.Model)
	sc.SetAttribute("chat.input_length", a.InputLength)
	sc.SetAttribute("chat.has_tools", a.HasTools)
	sc.SetAttribute("chat.tool_count", a.ToolCount)
}

// ToolSpanAttributes 工具调用 Span 属性
type ToolSpanAttributes struct {
	ToolName   string
	Input      string
	Output     string
	Duration   time.Duration
	HasError   bool
}

// Apply 应用属性到 Span
func (a *ToolSpanAttributes) Apply(sc *SpanContext) {
	sc.SetAttribute("tool.name", a.ToolName)
	sc.SetAttribute("tool.input_length", len(a.Input))
	sc.SetAttribute("tool.output_length", len(a.Output))
	sc.SetAttribute("tool.duration_ms", a.Duration.Milliseconds())
	sc.SetAttribute("tool.has_error", a.HasError)
}

// SummarySpanAttributes 摘要 Span 属性
type SummarySpanAttributes struct {
	RoundsToSummarize int
	OldTokens         int
	SummaryTokens     int
	SavedTokens       int
	Duration          time.Duration
}

// Apply 应用属性到 Span
func (a *SummarySpanAttributes) Apply(sc *SpanContext) {
	sc.SetAttribute("summary.rounds", a.RoundsToSummarize)
	sc.SetAttribute("summary.old_tokens", a.OldTokens)
	sc.SetAttribute("summary.summary_tokens", a.SummaryTokens)
	sc.SetAttribute("summary.saved_tokens", a.SavedTokens)
	sc.SetAttribute("summary.duration_ms", a.Duration.Milliseconds())
}

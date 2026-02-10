package trace

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

// ProviderConfig 追踪提供程序配置
type ProviderConfig struct {
	// 服务名称
	ServiceName string
	// 服务版本
	ServiceVersion string
	// 环境
	Environment string
	// 是否启用
	Enabled bool
	// 采样率 (0.0 - 1.0)
	SamplingRate float64
	// 导出器类型: "stdout", "jaeger", "zipkin"
	ExporterType string
	// 导出器端点 (用于 jaeger/zipkin)
	ExporterEndpoint string
}

// DefaultProviderConfig 默认配置
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		ServiceName:    "iano-chat-agent",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Enabled:        true,
		SamplingRate:   1.0,
		ExporterType:   "stdout",
	}
}

// Provider 追踪提供程序
type Provider struct {
	provider *sdktrace.TracerProvider
	config   *ProviderConfig
}

// NewProvider 创建追踪提供程序
func NewProvider(config *ProviderConfig) (*Provider, error) {
	if !config.Enabled {
		return &Provider{config: config}, nil
	}

	// 创建资源
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}

	// 创建导出器
	exporter, err := createExporter(config)
	if err != nil {
		return nil, fmt.Errorf("创建导出器失败: %w", err)
	}

	// 创建提供程序
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SamplingRate)),
	)

	// 设置为全局提供程序
	otel.SetTracerProvider(provider)

	return &Provider{
		provider: provider,
		config:   config,
	}, nil
}

// createExporter 创建导出器
func createExporter(config *ProviderConfig) (sdktrace.SpanExporter, error) {
	switch config.ExporterType {
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "jaeger":
		// 需要导入 "go.opentelemetry.io/otel/exporters/jaeger"
		// return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.ExporterEndpoint)))
		return nil, fmt.Errorf("jaeger 导出器需要额外依赖")
	case "zipkin":
		// 需要导入 "go.opentelemetry.io/otel/exporters/zipkin"
		// return zipkin.New(config.ExporterEndpoint)
		return nil, fmt.Errorf("zipkin 导出器需要额外依赖")
	default:
		return stdouttrace.New()
	}
}

// Shutdown 关闭提供程序
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.provider != nil {
		return p.provider.Shutdown(ctx)
	}
	return nil
}

// IsEnabled 是否启用
func (p *Provider) IsEnabled() bool {
	return p.config.Enabled
}

// InitGlobalTracer 初始化全局追踪器
func InitGlobalTracer(config *ProviderConfig) (*Provider, error) {
	provider, err := NewProvider(config)
	if err != nil {
		return nil, err
	}

	// 更新全局追踪器
	GlobalTracer = NewTracer(config.ServiceName)

	return provider, nil
}

package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerManager manages multiple tracers for different resources
type TracerManager struct {
	// Maps resource name to tracer provider
	providers map[string]*sdktrace.TracerProvider
	// Default tracer provider
	defaultProvider *sdktrace.TracerProvider
	// Exporter for all tracer providers
	exporter sdktrace.SpanExporter
	// Span processor for all tracer providers
	processor sdktrace.SpanProcessor
}

// Global instance of the tracer manager
var tracerManager *TracerManager

// InitTracerManager initializes the global tracer manager
func InitTracerManager(defaultProvider *sdktrace.TracerProvider, exporter sdktrace.SpanExporter, processor sdktrace.SpanProcessor) {
	if tracerManager != nil {
		if err := tracerManager.Shutdown(context.Background()); err != nil {
			fmt.Printf("Error shutting down tracer manager: %v\n", err)
		}
	}
	tracerManager = &TracerManager{
		providers:       make(map[string]*sdktrace.TracerProvider),
		defaultProvider: defaultProvider,
		exporter:        exporter,
		processor:       processor,
	}
}

// GetTracerManager returns the global tracer manager
func GetTracerManager() *TracerManager {
	return tracerManager
}

// CreateTracerForResource creates a new tracer provider for a resource
func (tm *TracerManager) CreateTracerForResource(resourceName string, res *Resource) (trace.Tracer, error) {
	if provider, exists := tm.providers[resourceName]; exists {
		return provider.Tracer("otelgen"), nil
	}

	resAttrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(res.Name),
	}

	for k, v := range res.Attributes {
		resAttrs = append(resAttrs, attribute.String(k, v))
	}

	r := sdkresource.NewWithAttributes(
		semconv.SchemaURL,
		resAttrs...,
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(tm.exporter),
		sdktrace.WithResource(r),
	)
	if tm.processor != nil {
		tp.RegisterSpanProcessor(tm.processor)
	}

	tm.providers[resourceName] = tp

	return tp.Tracer("otelgen"), nil
}

// GetTracerForResource returns a tracer for the given resource
func (tm *TracerManager) GetTracerForResource(resourceName string) trace.Tracer {
	if provider, exists := tm.providers[resourceName]; exists {
		return provider.Tracer("otelgen")
	}

	return nil
}

// GetDefaultTracer returns the default tracer
func (tm *TracerManager) GetDefaultTracer() trace.Tracer {
	return tm.defaultProvider.Tracer("otelgen")
}

// GetExporter returns the exporter used by the tracer manager
func (tm *TracerManager) GetExporter() sdktrace.SpanExporter {
	return tm.exporter
}

// GetSpanProcessor returns the span processor used by the tracer manager
func (tm *TracerManager) GetSpanProcessor() sdktrace.SpanProcessor {
	return tm.processor
}

// Shutdown closes all tracer providers
func (tm *TracerManager) Shutdown(ctx context.Context) error {
	var lastErr error
	for _, provider := range tm.providers {
		if err := provider.Shutdown(ctx); err != nil {
			lastErr = err
		}
	}

	if err := tm.defaultProvider.Shutdown(ctx); err != nil {
		lastErr = err
	}

	return lastErr
}

// CreateDefaultTracerProvider creates a default tracer provider
func CreateDefaultTracerProvider(exporter sdktrace.SpanExporter, processor sdktrace.SpanProcessor) (*sdktrace.TracerProvider, error) {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(resource.Default()),
	)
	if processor != nil {
		tp.RegisterSpanProcessor(processor)
	}

	return tp, nil
}

package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
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
}

// Global instance of the tracer manager
var tracerManager *TracerManager

// InitTracerManager initializes the global tracer manager
func InitTracerManager(defaultProvider *sdktrace.TracerProvider) {
	if tracerManager != nil {
		if err := tracerManager.Shutdown(context.Background()); err != nil {
			fmt.Printf("Error shutting down tracer manager: %v\n", err)
		}
	}
	tracerManager = &TracerManager{
		providers:       make(map[string]*sdktrace.TracerProvider),
		defaultProvider: defaultProvider,
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

	ctx := context.Background()
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
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
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(r),
	)

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
func CreateDefaultTracerProvider() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(resource.Default()),
	)

	return tp, nil
}

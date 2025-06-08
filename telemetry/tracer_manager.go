package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
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
	// Function to create a new exporter
	exporterFn func() (sdktrace.SpanExporter, error)
	// Function to create a new span processor
	processorFn func() (sdktrace.SpanProcessor, error)
	// Current span processor. This is only used for testing
	processor sdktrace.SpanProcessor
}

// Global instance of the tracer manager
var tracerManager *TracerManager

// InitTracerManager initializes the global tracer manager
func InitTracerManager(exporterFn func() (sdktrace.SpanExporter, error), processorFn func() (sdktrace.SpanProcessor, error)) error {
	if tracerManager != nil {
		if err := tracerManager.Shutdown(context.Background()); err != nil {
			fmt.Printf("Error shutting down tracer manager: %v\n", err)
		}
	}
	exporter, err := exporterFn()
	if err != nil {
		return err
	}
	defaultProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(sdkresource.Default()),
	)
	var processor sdktrace.SpanProcessor
	if processorFn != nil {
		processor, err = processorFn()
		if err != nil {
			return err
		}
		defaultProvider.RegisterSpanProcessor(processor)
	}
	tracerManager = &TracerManager{
		providers:       make(map[string]*sdktrace.TracerProvider),
		defaultProvider: defaultProvider,
		exporterFn:      exporterFn,
		processorFn:     processorFn,
		processor:       processor,
	}
	return nil
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

	exporter, err := tm.exporterFn()
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(r),
	)
	var processor sdktrace.SpanProcessor
	if tm.processorFn != nil {
		processor, err = tm.processorFn()
		if err != nil {
			return nil, err
		}
		tp.RegisterSpanProcessor(processor)
	}

	tm.processor = processor
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

// GetExporterFn returns the function to create exporter used by the tracer manager
func (tm *TracerManager) GetExporterFn() func() (sdktrace.SpanExporter, error) {
	return tm.exporterFn
}

// GetSpanProcessorFn returns the function to create span processor used by the tracer manager
func (tm *TracerManager) GetSpanProcessorFn() func() (sdktrace.SpanProcessor, error) {
	return tm.processorFn
}

// GetSpanProcessor returns the current span processor. This is only used for testing
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

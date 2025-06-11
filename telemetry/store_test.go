package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSpanAddChild(t *testing.T) {
	span := &Span{Name: "parent"}
	child := &Span{Name: "child"}

	span.AddChild(child)

	assert.Equal(t, 1, len(span.Children), "Expected 1 child span")
	assert.Equal(t, child, span.Children[0], "Expected child span to be added correctly")
}

func TestStore(t *testing.T) {
	t.Run("Trace", func(t *testing.T) {
		InitStore()

		traces := GetTraces()
		assert.NotNil(t, traces, "Expected traces to be initialized")
		assert.Empty(t, traces, "Expected no traces initially")
		assert.False(t, IsTraceExists("test_trace"), "Expected trace to not exist initially")

		traceName := "test_trace"
		trace := CreateTrace(traceName)

		traces = GetTraces()
		assert.Len(t, traces, 1, "Expected one trace after creation")
		assert.Equal(t, trace, traces[traceName], "Expected created trace to be retrieved correctly")
		assert.True(t, IsTraceExists(traceName), "Expected trace to exist after creation")
	})

	t.Run("Span", func(t *testing.T) {
		InitStore()

		spans := GetSpans()
		assert.NotNil(t, spans, "Expected spans to be initialized")
		assert.Empty(t, spans, "Expected no spans initially")
		assert.False(t, IsSpanExists("test_span"), "Expected span to not exist initially")

		traceName := "test_trace"

		t.Run("AddToTrace_OK", func(t *testing.T) {
			InitStore()

			trace := CreateTrace(traceName)

			spanName := "test_span"
			span, err := AddSpanToTrace(traceName, spanName, map[string]string{"key": "value"})
			assert.NoError(t, err, "Expected no error when adding span to trace")

			spans = GetSpans()
			assert.Len(t, spans, 1, "Expected one span after creation")
			assert.Equal(t, span, spans[spanName], "Expected created span to be retrieved correctly")
			assert.Equal(t, trace.RootSpan, span, "Expected span to be added to trace root span")
			assert.True(t, IsSpanExists(spanName), "Expected span to exist after creation")
		})

		t.Run("AddToTrace_Error", func(t *testing.T) {
			InitStore()

			_, err := AddSpanToTrace("non_existent_trace", "span_in_non_existent_trace", map[string]string{})
			assert.Error(t, err, "Expected error when adding span to non-existent trace")
		})

		t.Run("AddToSpan_OK", func(t *testing.T) {
			InitStore()

			CreateTrace(traceName)

			parentSpanName := "parent_span"

			parentSpan, err := AddSpanToTrace(traceName, parentSpanName, map[string]string{"key": "value"})
			assert.NoError(t, err, "Expected no error when adding parent span to trace")

			childSpanName := "child_span"
			childSpan, err := AddSpanToSpan(parentSpanName, childSpanName, map[string]string{"key": "value"})
			assert.NoError(t, err, "Expected no error when adding child span to parent span")

			spans = GetSpans()
			assert.Len(t, spans, 2, "Expected two spans after adding child span")
			assert.Equal(t, childSpan, spans[childSpanName], "Expected created child span to be retrieved correctly")
			assert.Contains(t, parentSpan.Children, childSpan, "Expected child span to be added to parent span's children")
		})

		t.Run("AddToSpan_Error", func(t *testing.T) {
			_, err := AddSpanToSpan("non_existent_span", "child_of_non_existent_span", map[string]string{})
			assert.Error(t, err, "Expected error when adding child span to non-existent parent span")
		})
	})

	t.Run("Resource", func(t *testing.T) {
		InitStore()

		resources := GetResources()
		assert.NotNil(t, resources, "Expected resources to be initialized")
		assert.Empty(t, resources, "Expected no resources initially")
		assert.False(t, IsResourceExists("test_resource"), "Expected resource to not exist initially")

		resourceName := "test_resource"
		resource := CreateResource(resourceName, map[string]string{"key": "value"})

		resources = GetResources()
		assert.Len(t, resources, 1, "Expected one resource after creation")
		assert.Equal(t, resource, resources[resourceName], "Expected created resource to be retrieved correctly")
		assert.True(t, IsResourceExists(resourceName), "Expected resource to exist after creation")

		traceName := "test_trace"

		t.Run("SetResourceToSpan_OK", func(t *testing.T) {
			InitStore()

			CreateTrace(traceName)
			spanName := "span_with_resource"
			span, err := AddSpanToTrace(traceName, spanName, map[string]string{"key": "value"})
			assert.NoError(t, err, "Expected no error when adding span to trace")

			resourceName := "resource_for_span"
			resource := CreateResource(resourceName, map[string]string{"key": "value"})

			setResource, err := SetResourceToSpan(spanName, resourceName)
			assert.NoError(t, err, "Expected no error when setting resource to span")
			assert.Equal(t, resource, setResource, "Expected set resource to match created resource")
			assert.Equal(t, resource.Name, span.Resource.Name, "Expected span's resource to be set correctly")
		})

		t.Run("SetResourceToSpan_Error", func(t *testing.T) {
			InitStore()

			_, err := SetResourceToSpan("non_existent_span", "non_existent_resource")
			assert.Error(t, err, "Expected error when setting resource to non-existent span")
		})
	})
}

func TestSendAllTraces(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()

	exporterFn := func() (trace.SpanExporter, error) {
		return tracetest.NewNoopExporter(), nil
	}
	processorFn := func() (trace.SpanProcessor, error) {
		return recorder, nil
	}

	InitTracerManager(exporterFn, processorFn)
	t.Cleanup(func() {
		if err := GetTracerManager().Shutdown(context.Background()); err != nil {
			t.Fatalf("Failed to shutdown tracer manager: %v", err)
		}
	})

	InitStore()

	// Trace
	traceName := "test_trace"
	CreateTrace(traceName)

	// Resource
	resourceName := "test_service"
	CreateResource(resourceName, map[string]string{
		"service.version": "1.0.0",
		"environment":     "test",
	})

	// Spans
	rootSpanName := "root_span"
	rootSpan, err := AddSpanToTrace(traceName, rootSpanName, map[string]string{
		"operation": "main",
		"status":    "success",
	})
	assert.NoError(t, err)

	_, err = SetResourceToSpan(rootSpanName, resourceName)
	assert.NoError(t, err)

	childSpan1Name := "child_span_1"
	_, err = AddSpanToSpan(rootSpanName, childSpan1Name, map[string]string{
		"operation": "process_data",
		"status":    "success",
	})
	assert.NoError(t, err)

	_, err = SetResourceToSpan(childSpan1Name, resourceName)
	assert.NoError(t, err)

	grandChildSpanName := "grandchild_span"
	_, err = AddSpanToSpan(childSpan1Name, grandChildSpanName, map[string]string{
		"operation": "validate_input",
		"status":    "success",
	})
	assert.NoError(t, err)

	_, err = SetResourceToSpan(grandChildSpanName, resourceName)
	assert.NoError(t, err)

	childSpan2Name := "child_span_2"
	childSpan2, err := AddSpanToSpan(rootSpanName, childSpan2Name, map[string]string{
		"operation": "store_result",
		"status":    "success",
	})
	assert.NoError(t, err)

	// Link
	rootSpan.AddLink(childSpan2, map[string]string{"key": "value"})

	_, err = SetResourceToSpan(childSpan2Name, resourceName)
	assert.NoError(t, err)

	SendAllTraces()

	spans := recorder.Ended()
	assert.Equal(t, 4, len(spans), "Expected 4 spans to be exported")

	spanNames := make(map[string]bool)
	gotSpans := make(map[string]trace.ReadOnlySpan)
	for _, span := range spans {
		gotSpanName := span.Name()

		spanNames[gotSpanName] = true

		attrs := span.Attributes()
		if gotSpanName == rootSpanName {
			assert.Equal(t, "main", getAttributeValue(attrs, "operation"))
			assert.Equal(t, "success", getAttributeValue(attrs, "status"))
			assert.False(t, span.Parent().HasSpanID())
		} else if gotSpanName == childSpan1Name {
			assert.Equal(t, "process_data", getAttributeValue(attrs, "operation"))
		} else if gotSpanName == grandChildSpanName {
			assert.Equal(t, "validate_input", getAttributeValue(attrs, "operation"))
		} else if gotSpanName == "child_span_2" {
			assert.Equal(t, "store_result", getAttributeValue(attrs, "operation"))
		}
		gotSpans[gotSpanName] = span
	}

	for _, span := range spans {
		gotSpanName := span.Name()
		if gotSpanName == rootSpanName {
			assert.Equal(t, gotSpans[childSpan2Name].SpanContext().SpanID(), gotSpans[rootSpanName].Links()[0].SpanContext.SpanID())
			assert.Len(t, span.Links(), 1, "Root span should have one link to child_span_2")
			assert.Equal(t, "value", span.Links()[0].Attributes[0].Value.AsString(), "Link attribute should match")
		} else if gotSpanName == childSpan1Name {
			assert.Equal(t, gotSpans[rootSpanName].SpanContext().SpanID(), span.Parent().SpanID())
		} else if gotSpanName == grandChildSpanName {
			assert.Equal(t, gotSpans[childSpan1Name].SpanContext().SpanID(), span.Parent().SpanID())
		} else if gotSpanName == "child_span_2" {
			assert.Equal(t, gotSpans[rootSpanName].SpanContext().SpanID(), span.Parent().SpanID())
		}
	}

	assert.True(t, spanNames[rootSpanName], "Root span should be exported")
	assert.True(t, spanNames[childSpan1Name], "Child span 1 should be exported")
	assert.True(t, spanNames[grandChildSpanName], "Grandchild span should be exported")
	assert.True(t, spanNames[childSpan2Name], "Child span 2 should be exported")

	assert.Empty(t, GetTraces(), "Store should be reset after sending traces")
	assert.Empty(t, GetSpans(), "Store should be reset after sending spans")
	assert.Empty(t, GetResources(), "Store should be reset after sending resources")
}

func getAttributeValue(attributes []attribute.KeyValue, key string) string {
	for _, attr := range attributes {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}

package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Span struct {
	Name       string
	Attributes map[string]string
	Children   []*Span
}

func (s *Span) AddChild(child *Span) {
	s.Children = append(s.Children, child)
}

type Trace struct {
	Name     string
	RootSpan *Span
}

type Store struct {
	traces map[string]*Trace
	spans  map[string]*Span
}

var store *Store

func InitStore() {
	store = &Store{
		traces: make(map[string]*Trace),
		spans:  make(map[string]*Span),
	}
}

func GetTraces() map[string]*Trace {
	return store.traces
}

func GetSpans() map[string]*Span {
	return store.spans
}

func CreateTrace(name string) *Trace {
	trace := &Trace{
		Name: name,
	}
	store.traces[name] = trace
	return trace
}

func AddSpanToTrace(traceName, spanName string, attributes map[string]string) error {
	trace, ok := store.traces[traceName]
	if !ok {
		return fmt.Errorf("trace %s not found", traceName)
	}
	span := Span{
		Name:       spanName,
		Attributes: attributes,
	}
	trace.RootSpan = &span
	store.spans[spanName] = &span
	return nil
}

func AddSpanToSpan(parentSpanName, spanName string, attributes map[string]string) error {
	parentSpan, ok := store.spans[parentSpanName]
	if !ok {
		return fmt.Errorf("parent span %s not found", parentSpanName)
	}
	span := Span{
		Name:       spanName,
		Attributes: attributes,
	}
	parentSpan.AddChild(&span)
	return nil
}

func SendAllTraces() {
	tracer := otel.Tracer("otelgen")

	for _, traceData := range store.traces {
		if traceData.RootSpan != nil {
			spanCount := 0
			processSpan(nil, tracer, traceData.RootSpan, &spanCount, 1.0, nil)
			fmt.Printf("Trace '%s' sent with %d spans.\n", traceData.Name, spanCount)
		} else {
			fmt.Printf("Trace '%s' has no spans.\n", traceData.Name)
		}
	}
}

// processSpan handles span creation with duration distribution:
// - Root span takes 1 second (current time - 1s to current time)
// - Each child span takes 90% of parent's duration
// - Child spans are centered within their parent's timeframe
func processSpan(parentCtx context.Context, tracer trace.Tracer, s *Span, spanCount *int, parentDuration float64, parentStartTime *time.Time) {
	attrs := []attribute.KeyValue{}
	for k, v := range s.Attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	var (
		spanCtx   context.Context
		span      trace.Span
		startTime time.Time
	)

	if parentCtx == nil {
		now := time.Now()
		startTime = now.Add(-1 * time.Second)

		spanCtx, span = tracer.Start(context.Background(), s.Name, trace.WithAttributes(attrs...), trace.WithTimestamp(startTime))
	} else {
		childDuration := parentDuration * 0.9

		timePadding := time.Duration((parentDuration - childDuration) / 2 * float64(time.Second))
		startTime = parentStartTime.Add(timePadding)

		spanCtx, span = tracer.Start(parentCtx, s.Name, trace.WithAttributes(attrs...), trace.WithTimestamp(startTime))
	}

	*spanCount++

	for _, childSpan := range s.Children {
		childDuration := parentDuration * 0.9
		processSpan(spanCtx, tracer, childSpan, spanCount, childDuration, &startTime)
	}

	span.End()
}

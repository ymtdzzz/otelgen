package telemetry

import (
	"context"
	"fmt"
	"maps"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Resource struct {
	Name       string
	Attributes map[string]string
}

type Span struct {
	Name       string
	Attributes map[string]string
	Children   []*Span
	Resource   *Resource
}

func (s *Span) AddChild(child *Span) {
	s.Children = append(s.Children, child)
}

type Trace struct {
	Name     string
	RootSpan *Span
}

type Store struct {
	traces    map[string]*Trace
	spans     map[string]*Span
	resources map[string]*Resource
}

var store *Store

func InitStore() {
	store = &Store{
		traces:    make(map[string]*Trace),
		spans:     make(map[string]*Span),
		resources: make(map[string]*Resource),
	}
}

func GetTraces() map[string]*Trace {
	return store.traces
}

func GetSpans() map[string]*Span {
	return store.spans
}

func GetResources() map[string]*Resource {
	return store.resources
}

func IsTraceExists(name string) bool {
	_, exists := store.traces[name]
	return exists
}

func IsSpanExists(name string) bool {
	_, exists := store.spans[name]
	return exists
}

func IsResourceExists(name string) bool {
	_, exists := store.resources[name]
	return exists
}

func CreateTrace(name string) *Trace {
	trace := &Trace{
		Name: name,
	}
	store.traces[name] = trace
	return trace
}

func UpdateSpan(name, newName, resource string, attributes map[string]string) (*Span, error) {
	span, ok := store.spans[name]
	if !ok {
		return nil, fmt.Errorf("span %s not found", name)
	}
	if newName != "" {
		if _, exists := store.spans[newName]; exists {
			return nil, fmt.Errorf("span with name %s already exists", newName)
		}
		delete(store.spans, name)
		span.Name = newName
		store.spans[newName] = span
	}
	if resource != "" {
		res, ok := store.resources[resource]
		if !ok {
			return nil, fmt.Errorf("resource %s not found", resource)
		}
		span.Resource = res
	}
	if attributes != nil {
		span.Attributes = make(map[string]string)
		maps.Copy(span.Attributes, attributes)
	}
	return span, nil
}

func CreateResource(name string, attributes map[string]string) *Resource {
	resource := &Resource{
		Name:       name,
		Attributes: attributes,
	}
	store.resources[name] = resource
	return resource
}

func UpdateResource(name, newName string, attributes map[string]string) (*Resource, error) {
	resource, ok := store.resources[name]
	if !ok {
		return nil, fmt.Errorf("resource %s not found", name)
	}
	if newName != "" {
		if _, exists := store.resources[newName]; exists {
			return nil, fmt.Errorf("resource with name %s already exists", newName)
		}
		delete(store.resources, name)
		resource.Name = newName
		store.resources[newName] = resource
	}
	if attributes != nil {
		resource.Attributes = make(map[string]string)
		maps.Copy(resource.Attributes, attributes)
	}
	return resource, nil
}

func AddSpanToTrace(traceName, spanName string, attributes map[string]string) (*Span, error) {
	trace, ok := store.traces[traceName]
	if !ok {
		return nil, fmt.Errorf("trace %s not found", traceName)
	}
	span := Span{
		Name:       spanName,
		Attributes: attributes,
	}
	if trace.RootSpan != nil {
		return nil, fmt.Errorf("trace %s already has a root span", traceName)
	}
	trace.RootSpan = &span
	store.spans[spanName] = &span
	return &span, nil
}

func AddSpanToSpan(parentSpanName, spanName string, attributes map[string]string) (*Span, error) {
	parentSpan, ok := store.spans[parentSpanName]
	if !ok {
		return nil, fmt.Errorf("parent span %s not found", parentSpanName)
	}
	span := Span{
		Name:       spanName,
		Attributes: attributes,
	}
	parentSpan.AddChild(&span)
	store.spans[spanName] = &span
	return &span, nil
}

func SetResourceToSpan(spanName, resourceName string) (*Resource, error) {
	span, ok := store.spans[spanName]
	if !ok {
		return nil, fmt.Errorf("span %s not found", spanName)
	}
	resource, ok := store.resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource %s not found", resourceName)
	}
	span.Resource = resource
	return resource, nil

}

func SendAllTraces() {
	for _, traceData := range store.traces {
		if traceData.RootSpan != nil {
			spanCount := 0
			processSpan(nil, traceData.RootSpan, &spanCount, 1.0, nil)
			fmt.Printf("Trace '%s' sent with %d spans.\n", traceData.Name, spanCount)
		} else {
			fmt.Printf("Trace '%s' has no spans.\n", traceData.Name)
		}
	}

	InitStore()

	exporter := GetTracerManager().GetExporter()
	processor := GetTracerManager().GetSpanProcessor()

	InitTracerManager(exporter, processor)
}

// processSpan handles span creation with duration distribution:
// - Root span takes 1 second (current time - 1s to current time)
// - Each child span takes 90% of parent's duration
// - Child spans are centered within their parent's timeframe
func processSpan(parentCtx context.Context, s *Span, spanCount *int, parentDuration float64, parentStartTime *time.Time) {
	var tracer trace.Tracer
	if s.Resource != nil {
		tm := GetTracerManager()
		resourceName := s.Resource.Name

		if t := tm.GetTracerForResource(resourceName); t != nil {
			tracer = t
		} else {
			if t, err := tm.CreateTracerForResource(resourceName, s.Resource); err != nil {
				fmt.Printf("Warning: Failed to create tracer for resource '%s': %v\n", resourceName, err)
				tracer = tm.GetDefaultTracer()
			} else {
				tracer = t
			}
		}
	} else {
		// Use default tracer when no resource is attached to span
		tracer = GetTracerManager().GetDefaultTracer()
	}
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
		processSpan(spanCtx, childSpan, spanCount, childDuration, &startTime)
	}

	span.End()
}

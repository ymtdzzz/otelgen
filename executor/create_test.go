package executor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otelgen/telemetry"
)

func TestHandleCreateSpan_Trace_OK(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateResource("my-resource", map[string]string{})

	cmd, err := ParseCommand("create span my-span in trace my-trace attributes key=value,http.method=GET resource my-resource")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.Create, "Create command should not be nil")

	handleCreateCommand(cmd.Create)

	trace, exists := telemetry.GetTraces()["my-trace"]
	assert.True(t, exists, "Trace should exist after creation")
	assert.Equal(t, "my-trace", trace.Name, "Trace name should match")

	span, exists := telemetry.GetSpans()["my-span"]
	assert.True(t, exists, "Span should exist after creation")
	assert.Equal(t, "my-span", span.Name, "Span name should match")
	assert.Equal(t, map[string]string{"key": "value", "http.method": "GET"}, span.Attributes, "Span attributes should match")
	assert.Equal(t, "my-resource", span.Resource.Name, "Span resource should match")
}

func TestHandleCreateSpan_Trace_RootSpanAlreadyExists(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateTrace("my-trace")
	telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{})

	cmd, err := ParseCommand("create span err-span in trace my-trace")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.Create, "Create command should not be nil")

	output := captureOutput(func() {
		handleCreateCommand(cmd.Create)
	})

	assert.Equal(t, "Error creating span: trace my-trace already has a root span\n", output)
}

func TestHandleCreateSpan_ParentSpan_OK(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateTrace("my-trace")
	telemetry.AddSpanToTrace("my-trace", "parent-span", map[string]string{})

	cmd, err := ParseCommand("create span child-span with parent parent-span attributes key=value,http.method=GET")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.Create, "Create command should not be nil")

	handleCreateCommand(cmd.Create)

	span, exists := telemetry.GetSpans()["child-span"]
	assert.True(t, exists, "Child span should exist after creation")
	assert.Equal(t, "child-span", span.Name, "Child span name should match")
	assert.Equal(t, map[string]string{"key": "value", "http.method": "GET"}, span.Attributes, "Child span attributes should match")

	parentSpan, exists := telemetry.GetSpans()["parent-span"]
	assert.True(t, exists, "Parent span should exist")
	assert.Equal(t, span, parentSpan.Children[0], "Child span should be added to parent span's children")
}

func TestHandleCreateResource(t *testing.T) {
	telemetry.InitStore()

	cmd, err := ParseCommand("create resource my-resource attributes key=value,http.method=GET")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.Create, "Create command should not be nil")

	handleCreateCommand(cmd.Create)
	resource, exists := telemetry.GetResources()["my-resource"]
	assert.True(t, exists, "Resource should exist after creation")
	assert.Equal(t, "my-resource", resource.Name, "Resource name should match")
	assert.Equal(t, map[string]string{"key": "value", "http.method": "GET"}, resource.Attributes, "Resource attributes should match")
}

func TestHandleCreateCommand_ValidateError(t *testing.T) {
	telemetry.InitStore()

	cmd, err := ParseCommand("create span")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.Create, "Create command should not be nil")

	output := captureOutput(func() {
		handleCreateCommand(cmd.Create)
	})

	assert.True(t, strings.Contains(output, "Error validating create command"))
	assert.True(t, strings.Contains(output, "type and name must be specified"))
}

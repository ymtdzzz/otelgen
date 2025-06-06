package executor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otelgen/telemetry"
)

func TestHandleSetSpan_OK(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateResource("my-resource", map[string]string{})
	telemetry.CreateTrace("my-trace")
	telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{})

	cmd, err := ParseCommand("set span my-span name new-span-name resource my-resource")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.Set, "Set command should not be nil")

	handleSetCommand(cmd.Set)

	_, exists := telemetry.GetSpans()["my-span"]
	assert.False(t, exists)

	span, exists := telemetry.GetSpans()["new-span-name"]
	assert.True(t, exists)
	assert.Equal(t, "new-span-name", span.Name, "Span name should match")
	assert.Equal(t, "my-resource", span.Resource.Name, "Span resource should match")
}

func TestHandleSetSpan_NonExistingResource(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateResource("my-resource", map[string]string{})
	telemetry.CreateTrace("my-trace")
	telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{})

	cmd, err := ParseCommand("set span my-span resource non-exsting-resource")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.Set, "Set command should not be nil")

	output := captureOutput(func() {
		handleSetCommand(cmd.Set)
	})

	assert.Equal(t, "Error validating set command: resource 'non-exsting-resource' does not exist\n", output)
}

func TestHandleSetResource_OK(t *testing.T) {
	telemetry.InitStore()
	telemetry.CreateResource("my-resource", map[string]string{})

	cmd, err := ParseCommand("set resource my-resource name new-resource-name")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.Set, "Set command should not be nil")

	handleSetCommand(cmd.Set)

	_, exists := telemetry.GetResources()["my-resource"]
	assert.False(t, exists)

	resource, exists := telemetry.GetResources()["new-resource-name"]
	assert.True(t, exists)
	assert.Equal(t, "new-resource-name", resource.Name)
}

func TestHandleSetCommand_ValidateError(t *testing.T) {
	telemetry.InitStore()

	cmd, err := ParseCommand("set span")
	assert.Nil(t, err, "ParseCommand should not return an error")
	assert.NotNil(t, cmd.Set, "Set command should not be nil")

	output := captureOutput(func() {
		handleSetCommand(cmd.Set)
	})

	assert.True(t, strings.Contains(output, "Error validating set command"))
	assert.True(t, strings.Contains(output, "type and name must be specified"))
}

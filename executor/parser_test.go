package executor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otelgen/telemetry"
)

func TestCreateCommandValidate(t *testing.T) {
	tests := []struct {
		input string
		want  error
	}{
		{
			input: "create span span1 in trace my-trace",
			want:  nil,
		},
		{
			input: "create resource resource1",
			want:  nil,
		},
		{
			input: "create span",
			want:  fmt.Errorf("type and name must be specified for create command"),
		},
		{
			input: "create resource",
			want:  fmt.Errorf("type and name must be specified for create command"),
		},
		{
			input: "create event",
			want:  fmt.Errorf("type and name must be specified for create command"),
		},
		{
			input: "create span span1",
			want:  fmt.Errorf("span must be created in a trace or with a parent span"),
		},
		{
			input: "create span span1 in trace trace1 with parent span1",
			want:  fmt.Errorf("span cannot have both a trace and a parent span"),
		},
		{
			input: "create span span1 with parent non_existing_span",
			want:  fmt.Errorf("parent span 'non_existing_span' does not exist"),
		},
		{
			input: "create span span1 in trace my-trace resource non_existing_resource",
			want:  fmt.Errorf("resource 'non_existing_resource' does not exist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			telemetry.InitStore()
			telemetry.CreateTrace("my-trace")
			telemetry.CreateResource("my-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("my-span", "my-resource")

			gotCmd, err := ParseCommand(tt.input)
			assert.Nil(t, err, "ParseCommand should not return an error for input: %s", tt.input)
			assert.NotNil(t, gotCmd.Create, "Create command should not be nil for input: %s", tt.input)
			gotErr := gotCmd.Create.Validate()
			assert.Equal(t, tt.want, gotErr, "Validate should return %v for input: %s", tt.want, tt.input)
		})
	}
}

func TestSetCommandValidate(t *testing.T) {
	tests := []struct {
		input string
		want  error
	}{
		{
			input: "set span my-span name new-my-span resource my-resource",
			want:  nil,
		},
		{
			input: "set span my-span resource my-resource name new-my-span",
			want:  nil,
		},
		{
			input: "set span my-span resource my-resource name new-my-span resource another-resource",
			want:  fmt.Errorf("duplicated operation: resource"),
		},
		{
			input: "set span my-span name new-my-span resource my-resource name new-my-span-2",
			want:  fmt.Errorf("duplicated operation: name"),
		},
		{
			input: "set span my-span attributes key=val name new-my-span attributes another_key=another_val",
			want:  fmt.Errorf("duplicated operation: attributes"),
		},
		{
			input: "set span non-existing-span name new-my-span",
			want:  fmt.Errorf("span 'non-existing-span' does not exist"),
		},
		{
			input: "set span my-span name new-my-span resource non-existing-resource",
			want:  fmt.Errorf("resource 'non-existing-resource' does not exist"),
		},
		{
			input: "set resource non-existing-resource name new-resource-name",
			want:  fmt.Errorf("resource 'non-existing-resource' does not exist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			telemetry.InitStore()
			telemetry.CreateTrace("my-trace")
			telemetry.CreateResource("my-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("my-span", "my-resource")

			gotCmd, err := ParseCommand(tt.input)
			assert.Nil(t, err, "ParseCommand should not return an error for input: %s", tt.input)
			assert.NotNil(t, gotCmd.Set, "Set command should not be nil for input: %s", tt.input)
			gotErr := gotCmd.Set.Validate()
			assert.Equal(t, tt.want, gotErr, "Validate should return %v for input: %s", tt.want, tt.input)
		})
	}
}

func TestAddLinkCommandValidate(t *testing.T) {
	tests := []struct {
		input string
		want  error
	}{
		{
			input: "add link my-span another-span",
			want:  nil,
		},
		{
			input: "add link my-span another-span attributes key=value",
			want:  nil,
		},
		{
			input: "add link",
			want:  fmt.Errorf("both 'from' and 'to' must be specified for add link command"),
		},
		{
			input: "add link wrong-span another-span",
			want:  fmt.Errorf("span 'wrong-span' does not exist"),
		},
		{
			input: "add link my-span wrong-span",
			want:  fmt.Errorf("span 'wrong-span' does not exist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			telemetry.InitStore()
			telemetry.CreateTrace("my-trace")
			telemetry.CreateResource("my-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("my-span", "my-resource")

			telemetry.CreateTrace("another-trace")
			telemetry.AddSpanToTrace("another-trace", "another-span", map[string]string{"key": "value"})

			gotCmd, err := ParseCommand(tt.input)
			assert.Nil(t, err, "ParseCommand should not return an error for input: %s", tt.input)
			assert.NotNil(t, gotCmd.AddLink, "AddLink command should not be nil for input: %s", tt.input)
			gotErr := gotCmd.AddLink.Validate()
			assert.Equal(t, tt.want, gotErr, "Validate should return %v for input: %s", tt.want, tt.input)
		})
	}
}

func TestAddEventCommandValidate(t *testing.T) {
	tests := []struct {
		input string
		want  error
	}{
		{
			input: "add event my-span my-event",
			want:  nil,
		},
		{
			input: "add event",
			want:  fmt.Errorf("event name must be specified for add event command"),
		},
		{
			input: "add event wrong-span my-event",
			want:  fmt.Errorf("span 'wrong-span' does not exist"),
		},
		{
			input: "add event my-span wrong-event",
			want:  fmt.Errorf("event 'wrong-event' does not exist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			telemetry.InitStore()
			telemetry.CreateTrace("my-trace")
			telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{"key": "value"})
			telemetry.CreateEvent("my-event", map[string]string{"key": "value"})

			gotCmd, err := ParseCommand(tt.input)
			assert.Nil(t, err, "ParseCommand should not return an error for input: %s", tt.input)
			assert.NotNil(t, gotCmd.AddEvent, "AddEvent command should not be nil for input: %s", tt.input)
			gotErr := gotCmd.AddEvent.Validate()
			assert.Equal(t, tt.want, gotErr, "Validate should return %v for input: %s", tt.want, tt.input)
		})
	}
}

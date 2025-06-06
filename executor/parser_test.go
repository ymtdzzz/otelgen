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
			want:  fmt.Errorf("duplicate operation: resource"),
		},
		{
			input: "set span my-span name new-my-span resource my-resource name new-my-span-2",
			want:  fmt.Errorf("duplicate operation: name"),
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

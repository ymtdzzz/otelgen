package completer

import (
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otelgen/telemetry"
)

func TestCompleteCommand(t *testing.T) {
	tests := []struct {
		input string
		want  []prompt.Suggest
	}{
		{
			input: "",
			want:  commandSuggestions[""],
		},
		{
			input: "c",
			want: []prompt.Suggest{
				{Text: "create", Description: "Create a new signal"},
			},
		},
		{
			input: "p",
			want:  []prompt.Suggest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			buf := prompt.NewBuffer()
			buf.InsertText(tt.input, false, true)
			doc := buf.Document()
			got := Completer(*doc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompleteCreate(t *testing.T) {
	tests := []struct {
		input string
		want  []prompt.Suggest
	}{
		{
			input: "create ",
			want:  commandSuggestions["create_type"],
		},
		{
			input: "create r",
			want: []prompt.Suggest{
				{Text: "resource", Description: "Create a new resource"},
			},
		},
		{
			input: "create s",
			want: []prompt.Suggest{
				{Text: "span", Description: "Create a new span"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			buf := prompt.NewBuffer()
			buf.InsertText(tt.input, false, true)
			doc := buf.Document()
			got := Completer(*doc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompleteCreateSpan(t *testing.T) {
	tests := []struct {
		input string
		want  []prompt.Suggest
	}{
		{
			input: "create span ",
			want:  []prompt.Suggest{},
		},
		{
			input: "create span span1",
			want:  []prompt.Suggest{},
		},
		{
			input: "create span span1 ",
			want:  commandSuggestions["create_in_or_with"],
		},
		{
			input: "create span span1 i",
			want: []prompt.Suggest{
				{Text: "in", Description: "Create a span in a trace"},
			},
		},
		{
			input: "create span span1 in",
			want: []prompt.Suggest{
				{Text: "in", Description: "Create a span in a trace"},
			},
		},
		{
			input: "create span span1 in ",
			want:  commandSuggestions["create_in_trace"],
		},
		{
			input: "create span span1 in trace ",
			want: []prompt.Suggest{
				{Text: "me-trace"},
				{Text: "my-trace"},
			},
		},
		{
			input: "create span span1 in trace me",
			want: []prompt.Suggest{
				{Text: "me-trace"},
			},
		},
		{
			input: "create span span1 with ",
			want:  commandSuggestions["create_with_parent"],
		},
		{
			input: "create span span1 with parent ",
			want: []prompt.Suggest{
				{Text: "me-span"},
				{Text: "my-span"},
			},
		},
		{
			input: "create span span1 with parent me",
			want: []prompt.Suggest{
				{Text: "me-span"},
			},
		},
		{
			input: "create span span1 in trace my-trace ",
			want: []prompt.Suggest{
				{Text: "resource", Description: "Set a resource for the span"},
				{Text: "attributes", Description: "Add attributes to the span"},
			},
		},
		{
			input: "create span span1 in trace my-trace a",
			want: []prompt.Suggest{
				{Text: "attributes", Description: "Add attributes to the span"},
			},
		},
		{
			input: "create span span1 in trace my-trace attributes ",
			want:  []prompt.Suggest{},
		},
		{
			input: "create span span1 in trace my-trace resource ",
			want: []prompt.Suggest{
				{Text: "me-resource"},
				{Text: "my-resource"},
			},
		},
		{
			input: "create span span1 in trace my-trace resource me",
			want: []prompt.Suggest{
				{Text: "me-resource"},
			},
		},
		{
			input: "create span span1 in trace my-trace resource me-resource ",
			want: []prompt.Suggest{
				{Text: "attributes", Description: "Add attributes to the span"},
			},
		},
		{
			input: "create span span1 in trace my-trace attributes key=val ",
			want: []prompt.Suggest{
				{Text: "resource", Description: "Set a resource for the span"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			telemetry.InitStore()
			telemetry.CreateTrace("my-trace")
			telemetry.CreateResource("my-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("my-span", "my-resource")
			telemetry.CreateTrace("me-trace")
			telemetry.CreateResource("me-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("me-trace", "me-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("me-span", "me-resource")

			buf := prompt.NewBuffer()
			buf.InsertText(tt.input, false, true)
			doc := buf.Document()
			got := Completer(*doc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompleteCreateResource(t *testing.T) {
	tests := []struct {
		input string
		want  []prompt.Suggest
	}{
		{
			input: "create resource ",
			want:  []prompt.Suggest{},
		},
		{
			input: "create resource resource1",
			want:  []prompt.Suggest{},
		},
		{
			input: "create resource resource1 ",
			want: []prompt.Suggest{
				{Text: "attributes", Description: "Add attributes to the resource"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			buf := prompt.NewBuffer()
			buf.InsertText(tt.input, false, true)
			doc := buf.Document()
			got := Completer(*doc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompleteSet(t *testing.T) {
	tests := []struct {
		input string
		want  []prompt.Suggest
	}{
		{
			input: "set ",
			want:  commandSuggestions["set_type"],
		},
		{
			input: "set s",
			want: []prompt.Suggest{
				{Text: "span", Description: "Update a span"},
			},
		},
		{
			input: "set r",
			want: []prompt.Suggest{
				{Text: "resource", Description: "Update a resource"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			buf := prompt.NewBuffer()
			buf.InsertText(tt.input, false, true)
			doc := buf.Document()
			got := Completer(*doc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompleteSetSpan(t *testing.T) {
	tests := []struct {
		input string
		want  []prompt.Suggest
	}{
		{
			input: "set span ",
			want: []prompt.Suggest{
				{Text: "me-span"},
				{Text: "my-span"},
			},
		},
		{
			input: "set span my",
			want: []prompt.Suggest{
				{Text: "my-span"},
			},
		},
		{
			input: "set span my-span ",
			want:  commandSuggestions["set_span_operations"],
		},
		{
			input: "set span my-span n",
			want: []prompt.Suggest{
				{Text: "name", Description: "Set a new name for the span"},
			},
		},
		{
			input: "set span my-span name ",
			want:  []prompt.Suggest{},
		},
		{
			input: "set span my-span name new-span-name ",
			want: []prompt.Suggest{
				{Text: "resource", Description: "Set a resource for the span"},
			},
		},
		{
			input: "set span my-span name new-span-name resource ",
			want: []prompt.Suggest{
				{Text: "me-resource"},
				{Text: "my-resource"},
			},
		},
		{
			input: "set span my-span name new-span-name resource me-resource ",
			want:  []prompt.Suggest{},
		},
		{
			input: "set span my-span resource my-resource ",
			want: []prompt.Suggest{
				{Text: "name", Description: "Set a new name for the span"},
			},
		},
		{
			input: "set span my-span resource my-resource name ",
			want:  []prompt.Suggest{},
		},
		{
			input: "set span my-span resource my-resource name new-span-name ",
			want:  []prompt.Suggest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			telemetry.InitStore()
			telemetry.CreateTrace("my-trace")
			telemetry.CreateResource("my-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("my-span", "my-resource")
			telemetry.CreateTrace("me-trace")
			telemetry.CreateResource("me-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("me-trace", "me-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("me-span", "me-resource")

			buf := prompt.NewBuffer()
			buf.InsertText(tt.input, false, true)
			doc := buf.Document()
			got := Completer(*doc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompleteSetResource(t *testing.T) {
	tests := []struct {
		input string
		want  []prompt.Suggest
	}{
		{
			input: "set resource ",
			want: []prompt.Suggest{
				{Text: "me-resource"},
				{Text: "my-resource"},
			},
		},
		{
			input: "set resource my",
			want: []prompt.Suggest{
				{Text: "my-resource"},
			},
		},
		{
			input: "set resource my-resource ",
			want:  commandSuggestions["set_resource_operations"],
		},
		{
			input: "set resource my-resource n",
			want: []prompt.Suggest{
				{Text: "name", Description: "Set a new name for the resource"},
			},
		},
		{
			input: "set resource my-resource name ",
			want:  []prompt.Suggest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			telemetry.InitStore()
			telemetry.CreateTrace("my-trace")
			telemetry.CreateResource("my-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("my-span", "my-resource")
			telemetry.CreateTrace("me-trace")
			telemetry.CreateResource("me-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("me-trace", "me-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("me-span", "me-resource")

			buf := prompt.NewBuffer()
			buf.InsertText(tt.input, false, true)
			doc := buf.Document()
			got := Completer(*doc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompleteAdd(t *testing.T) {
	tests := []struct {
		input string
		want  []prompt.Suggest
	}{
		{
			input: "add ",
			want:  commandSuggestions["add_type"],
		},
		{
			input: "add l",
			want: []prompt.Suggest{
				{Text: "link", Description: "Add a link to the span"},
			},
		},
		{
			input: "add link ",
			want: []prompt.Suggest{
				{Text: "me-span"},
				{Text: "my-span"},
			},
		},
		{
			input: "add link me",
			want: []prompt.Suggest{
				{Text: "me-span"},
			},
		},
		{
			input: "add link me-span ",
			want: []prompt.Suggest{
				{Text: "me-span"},
				{Text: "my-span"},
			},
		},
		{
			input: "add link me-span my",
			want: []prompt.Suggest{
				{Text: "my-span"},
			},
		},
		{
			input: "add link me-span my-span ",
			want: []prompt.Suggest{
				{Text: "attributes", Description: "Add attributes to the link"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			telemetry.InitStore()
			telemetry.CreateTrace("my-trace")
			telemetry.CreateResource("my-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("my-trace", "my-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("my-span", "my-resource")
			telemetry.CreateTrace("me-trace")
			telemetry.CreateResource("me-resource", map[string]string{"key": "value"})
			telemetry.AddSpanToTrace("me-trace", "me-span", map[string]string{"key": "value"})
			telemetry.SetResourceToSpan("me-span", "me-resource")

			buf := prompt.NewBuffer()
			buf.InsertText(tt.input, false, true)
			doc := buf.Document()
			got := Completer(*doc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompleteList(t *testing.T) {
	tests := []struct {
		input string
		want  []prompt.Suggest
	}{
		{
			input: "list",
			want: []prompt.Suggest{
				{Text: "list", Description: "List available traces and spans"},
			},
		},
		{
			input: "list ",
			want:  commandSuggestions["list"],
		},
		{
			input: "list t",
			want: []prompt.Suggest{
				{Text: "traces", Description: "List all available traces"},
			},
		},
		{
			input: "list p",
			want:  []prompt.Suggest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			buf := prompt.NewBuffer()
			buf.InsertText(tt.input, false, true)
			doc := buf.Document()
			got := Completer(*doc)
			assert.Equal(t, tt.want, got)
		})
	}
}

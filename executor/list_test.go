package executor

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otelgen/telemetry"
)

func captureOutput(f func()) string {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

func TestListTraces(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		setupFunc func()
		want      string
	}{
		{
			name:  "list without target",
			input: "list",
			setupFunc: func() {
				telemetry.InitStore()
			},
			want: "No target specified for list command.\n",
		},
		{
			name:  "list traces with no traces",
			input: "list traces",
			setupFunc: func() {
				telemetry.InitStore()
			},
			want: "No traces available.\n",
		},
		{
			name:  "list traces with empty trace",
			input: "list traces",
			setupFunc: func() {
				telemetry.InitStore()
				telemetry.CreateTrace("empty-trace")
			},
			want: `Available traces: 1
----------------------------------------
Trace: empty-trace
  No spans in this trace
----------------------------------------
`,
		},
		{
			name:  "list traces with spans",
			input: "list traces",
			setupFunc: func() {
				telemetry.InitStore()

				telemetry.CreateTrace("test-trace")
				telemetry.AddSpanToTrace("test-trace", "root-span", map[string]string{
					"service.name": "test-service",
					"operation":    "test",
				})

				childSpan, _ := telemetry.AddSpanToSpan("root-span", "child-span", map[string]string{
					"http.method": "GET",
					"http.url":    "https://example.com",
				})

				resource := telemetry.CreateResource("test-resource", map[string]string{
					"service.name": "resource-service",
					"environment":  "test",
				})
				childSpan.Resource = resource
			},
			want: `Available traces: 1
----------------------------------------
Trace: test-trace
  - Span: root-span
    Attributes:
      operation: test
      service.name: test-service
    - Span: child-span
      Attributes:
        http.method: GET
        http.url: https://example.com
      Resource:
        Name: test-resource
        environment: test
        service.name: resource-service
----------------------------------------
`,
		},
		{
			name:  "list traces with multiple traces",
			input: "list traces",
			setupFunc: func() {
				telemetry.InitStore()

				telemetry.CreateTrace("trace1")
				telemetry.AddSpanToTrace("trace1", "span1", map[string]string{
					"trace": "1",
				})

				telemetry.CreateTrace("trace2")
				telemetry.AddSpanToTrace("trace2", "span2", map[string]string{
					"trace": "2",
				})
			},
			want: `Available traces: 2
----------------------------------------
Trace: trace1
  - Span: span1
    Attributes:
      trace: 1
----------------------------------------
Trace: trace2
  - Span: span2
    Attributes:
      trace: 2
----------------------------------------
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			output := captureOutput(func() {
				Executor(tt.input)
			})

			assert.Equal(t, tt.want, output)
		})
	}
}

func TestListResources(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		setupFunc func()
		want      string
	}{
		{
			name:  "list resources with no resources",
			input: "list resources",
			setupFunc: func() {
				telemetry.InitStore()
			},
			want: "No resources available.\n",
		},
		{
			name:  "list resources with one resource without attributes",
			input: "list resources",
			setupFunc: func() {
				telemetry.InitStore()
				telemetry.CreateResource("empty-resource", map[string]string{})
			},
			want: `Available resources: 1
----------------------------------------
Resource: empty-resource
  No attributes
----------------------------------------
`,
		},
		{
			name:  "list resources with one resource with attributes",
			input: "list resources",
			setupFunc: func() {
				telemetry.InitStore()
				telemetry.CreateResource("test-resource", map[string]string{
					"service.name": "test-service",
					"environment":  "test",
					"version":      "1.0.0",
				})
			},
			want: `Available resources: 1
----------------------------------------
Resource: test-resource
  Attributes:
    environment: test
    service.name: test-service
    version: 1.0.0
----------------------------------------
`,
		},
		{
			name:  "list resources with multiple resources",
			input: "list resources",
			setupFunc: func() {
				telemetry.InitStore()

				telemetry.CreateResource("resource1", map[string]string{
					"service.name": "service1",
					"environment":  "prod",
				})

				telemetry.CreateResource("resource2", map[string]string{
					"service.name": "service2",
					"environment":  "staging",
				})

				telemetry.CreateResource("resource3", map[string]string{})
			},
			want: `Available resources: 3
----------------------------------------
Resource: resource1
  Attributes:
    environment: prod
    service.name: service1
----------------------------------------
Resource: resource2
  Attributes:
    environment: staging
    service.name: service2
----------------------------------------
Resource: resource3
  No attributes
----------------------------------------
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			output := captureOutput(func() {
				Executor(tt.input)
			})

			assert.Equal(t, tt.want, output)
		})
	}
}

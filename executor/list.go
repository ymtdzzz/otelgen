package executor

import (
	"fmt"
	"strings"

	"github.com/ymtdzzz/otelgen/telemetry"
)

func handleListCommand(args []string) {
	if len(args) != 1 {
		println("Usage: list traces")
		return
	}

	switch args[0] {
	case "traces":
		listTraces()
	default:
		println("Unknown list command:", args[0])
	}
}

func listTraces() {
	traces := telemetry.GetTraces()
	if len(traces) == 0 {
		fmt.Println("No traces available.")
		return
	}

	fmt.Printf("Available traces: %d\n", len(traces))
	fmt.Println("----------------------------------------")

	for name, trace := range traces {
		fmt.Printf("Trace: %s\n", name)
		if trace.RootSpan == nil {
			fmt.Println("  No spans in this trace")
		} else {
			printSpan(trace.RootSpan, 1)
		}
		fmt.Println("----------------------------------------")
	}
}

func printSpan(span *telemetry.Span, depth int) {
	indent := strings.Repeat("  ", depth)

	fmt.Printf("%s- Span: %s\n", indent, span.Name)

	if len(span.Attributes) > 0 {
		fmt.Printf("%s  Attributes:\n", indent)
		for key, value := range span.Attributes {
			fmt.Printf("%s    %s: %s\n", indent, key, value)
		}
	}

	for _, childSpan := range span.Children {
		printSpan(childSpan, depth+1)
	}
}

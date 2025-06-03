package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ymtdzzz/otelgen/telemetry"
)

func handleListCommand(cmd *ListCommand) {
	if cmd.Target == nil {
		fmt.Println("No target specified for list command.")
		return
	}

	switch *cmd.Target {
	case "traces":
		listTraces()
	case "resources":
		listResources()
	default:
		fmt.Printf("Unknown target type for list command: %s\n", *cmd.Target)
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
		keys := make([]string, 0, len(span.Attributes))
		for key := range span.Attributes {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			fmt.Printf("%s    %s: %s\n", indent, key, span.Attributes[key])
		}
	}

	if span.Resource != nil {
		fmt.Printf("%s  Resource:\n", indent)
		fmt.Printf("%s    Name: %s\n", indent, span.Resource.Name)
		keys := make([]string, 0, len(span.Resource.Attributes))
		for key := range span.Resource.Attributes {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			fmt.Printf("%s    %s: %s\n", indent, key, span.Resource.Attributes[key])
		}
	}

	for _, childSpan := range span.Children {
		printSpan(childSpan, depth+1)
	}
}

func listResources() {
	resources := telemetry.GetResources()
	if len(resources) == 0 {
		fmt.Println("No resources available.")
		return
	}

	fmt.Printf("Available resources: %d\n", len(resources))
	fmt.Println("----------------------------------------")

	// Sort resource names for consistent output
	names := make([]string, 0, len(resources))
	for name := range resources {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		resource := resources[name]
		fmt.Printf("Resource: %s\n", name)

		if len(resource.Attributes) == 0 {
			fmt.Println("  No attributes")
		} else {
			fmt.Println("  Attributes:")
			// Sort attribute keys for consistent output
			keys := make([]string, 0, len(resource.Attributes))
			for key := range resource.Attributes {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				fmt.Printf("    %s: %s\n", key, resource.Attributes[key])
			}
		}
		fmt.Println("----------------------------------------")
	}
}

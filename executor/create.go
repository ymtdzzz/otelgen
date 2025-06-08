package executor

import (
	"fmt"

	"github.com/ymtdzzz/otelgen/telemetry"
)

func handleCreateCommand(cmd *CreateCommand) {
	if err := cmd.Validate(); err != nil {
		fmt.Printf("Error validating create command: %v\n", err)
		return
	}

	switch *cmd.Type {
	case "span":
		if err := handleCreateSpan(cmd); err != nil {
			fmt.Printf("Error creating span: %v\n", err)
		}
	case "resource":
		if err := handleCreateResource(cmd); err != nil {
			fmt.Printf("Error creating resource: %v\n", err)
		}
	default:
		fmt.Printf("Unknown target type for create command: %s\n", *cmd.Type)
	}
}

func handleCreateSpan(cmd *CreateCommand) error {
	var (
		resourceName string
		attributes   map[string]string
	)

	for _, arg := range cmd.Args {
		if arg.Resource != nil {
			resourceName = *arg.Resource
		}
		if len(arg.Attrs) > 0 {
			attributes = convertKeyValuesToMap(arg.Attrs)
		}
	}

	if cmd.Trace != nil {
		trace, exists := telemetry.GetTraces()[*cmd.Trace]
		if !exists {
			trace = telemetry.CreateTrace(*cmd.Trace)
			fmt.Printf("Created trace: %s\n", trace.Name)
		}
		span, err := telemetry.AddSpanToTrace(*cmd.Trace, *cmd.Name, attributes)
		if err != nil {
			return err
		}
		fmt.Printf("Created span: %s in trace: %s\n", span.Name, trace.Name)
	} else if cmd.ParentSpan != nil {
		span, err := telemetry.AddSpanToSpan(*cmd.ParentSpan, *cmd.Name, attributes)
		if err != nil {
			return err
		}
		fmt.Printf("Created span: %s with parent span: %s\n", span.Name, *cmd.ParentSpan)
	}
	if resourceName != "" {
		resource, err := telemetry.SetResourceToSpan(*cmd.Name, resourceName)
		if err != nil {
			return err
		}
		fmt.Printf("Set resource %s to span %s\n", resource.Name, *cmd.Name)
	}
	return nil
}

func handleCreateResource(cmd *CreateCommand) error {
	var (
		attributes map[string]string
	)

	for _, arg := range cmd.Args {
		if len(arg.Attrs) > 0 {
			attributes = convertKeyValuesToMap(arg.Attrs)
		}
	}

	resource := telemetry.CreateResource(*cmd.Name, attributes)
	fmt.Printf("Created resource: %s with attributes: %v\n", resource.Name, attributes)
	return nil
}

package executor

import (
	"fmt"

	"github.com/ymtdzzz/otelgen/telemetry"
)

func handleSetCommand(cmd *SetCommand) {
	if err := cmd.Validate(); err != nil {
		fmt.Printf("Error validating set command: %v\n", err)
		return
	}

	switch *cmd.Type {
	case "span":
		if err := handleSetSpan(cmd); err != nil {
			fmt.Printf("Error setting span: %v\n", err)
		}
	case "resource":
		if err := handleSetResource(cmd); err != nil {
			fmt.Printf("Error setting resource: %v\n", err)
		}
	default:
		fmt.Printf("Unknown target type for set command: %s\n", *cmd.Type)
	}
}

func handleSetSpan(cmd *SetCommand) error {
	var (
		newName      string
		resourceName string
	)

	for _, op := range cmd.Operations {
		if op.Name != nil {
			newName = *op.Name
		}
		if op.Resource != nil {
			resourceName = *op.Resource
		}
	}

	if _, err := telemetry.UpdateSpan(*cmd.Name, newName, resourceName); err != nil {
		return err
	}
	fmt.Printf("Updated span: %s with new name: %s and resource: %s\n", *cmd.Name, newName, resourceName)

	return nil
}

func handleSetResource(cmd *SetCommand) error {
	var newName string

	for _, op := range cmd.Operations {
		if op.Name != nil {
			newName = *op.Name
		}
	}

	if _, err := telemetry.UpdateResource(*cmd.Name, newName); err != nil {
		return err
	}
	fmt.Printf("Updated resource: %s with new name: %s\n", *cmd.Name, newName)

	return nil
}

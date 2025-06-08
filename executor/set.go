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

	for _, arg := range cmd.Args {
		if arg.SetCreateArg != nil {
			if arg.SetCreateArg.Resource != nil {
				resourceName = *arg.SetCreateArg.Resource
			}
		}
		if arg.SetOnlyArg != nil {
			if arg.SetOnlyArg.Name != nil {
				newName = *arg.SetOnlyArg.Name
			}
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

	for _, arg := range cmd.Args {
		if arg.SetOnlyArg != nil {
			if arg.SetOnlyArg.Name != nil {
				newName = *arg.SetOnlyArg.Name
			}
		}
	}

	if _, err := telemetry.UpdateResource(*cmd.Name, newName); err != nil {
		return err
	}
	fmt.Printf("Updated resource: %s with new name: %s\n", *cmd.Name, newName)

	return nil
}

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
	case "event":
		if err := handleSetEvent(cmd); err != nil {
			fmt.Printf("Error setting event: %v\n", err)
		}
	default:
		fmt.Printf("Unknown target type for set command: %s\n", *cmd.Type)
	}
}

func handleSetSpan(cmd *SetCommand) error {
	var (
		newName      string
		resourceName string
		attributes   map[string]string
	)

	for _, arg := range cmd.Args {
		if arg.SetCreateArg != nil {
			if arg.SetCreateArg.Resource != nil {
				resourceName = *arg.SetCreateArg.Resource
			}
			if len(arg.SetCreateArg.Attrs) > 0 {
				attributes = convertKeyValuesToMap(arg.SetCreateArg.Attrs)
			}
		}
		if arg.SetOnlyArg != nil {
			if arg.SetOnlyArg.Name != nil {
				newName = *arg.SetOnlyArg.Name
			}
		}
	}

	if _, err := telemetry.UpdateSpan(*cmd.Name, newName, resourceName, attributes); err != nil {
		return err
	}
	fmt.Printf("Updated span\n")

	return nil
}

func handleSetResource(cmd *SetCommand) error {
	var (
		newName    string
		attributes map[string]string
	)

	for _, arg := range cmd.Args {
		if arg.SetCreateArg != nil {
			if len(arg.SetCreateArg.Attrs) > 0 {
				attributes = convertKeyValuesToMap(arg.SetCreateArg.Attrs)
			}
		}
		if arg.SetOnlyArg != nil {
			if arg.SetOnlyArg.Name != nil {
				newName = *arg.SetOnlyArg.Name
			}
		}
	}

	if _, err := telemetry.UpdateResource(*cmd.Name, newName, attributes); err != nil {
		return err
	}
	fmt.Printf("Updated resource: %s with new name: %s\n", *cmd.Name, newName)

	return nil
}

func handleSetEvent(cmd *SetCommand) error {
	var (
		newName    string
		attributes map[string]string
	)

	for _, arg := range cmd.Args {
		if arg.SetCreateArg != nil {
			if len(arg.SetCreateArg.Attrs) > 0 {
				attributes = convertKeyValuesToMap(arg.SetCreateArg.Attrs)
			}
		}
		if arg.SetOnlyArg != nil {
			if arg.SetOnlyArg.Name != nil {
				newName = *arg.SetOnlyArg.Name
			}
		}
	}

	if _, err := telemetry.UpdateEvent(*cmd.Name, newName, attributes); err != nil {
		return err
	}
	fmt.Printf("Updated event: %s with new name: %s\n", *cmd.Name, newName)

	return nil
}

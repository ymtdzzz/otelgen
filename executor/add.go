package executor

import (
	"fmt"

	"github.com/ymtdzzz/otelgen/telemetry"
)

func handleAddLinkCommand(cmd *AddLinkCommand) {
	if err := cmd.Validate(); err != nil {
		fmt.Printf("Error validating add link command: %v\n", err)
		return
	}

	var (
		attributes map[string]string
	)

	for _, arg := range cmd.Args {
		if arg.Attrs != nil {
			attributes = convertKeyValuesToMap(arg.Attrs)
		}
	}

	_, err := telemetry.AddLinkToSpan(*cmd.From, *cmd.To, attributes)
	if err != nil {
		fmt.Printf("Error adding link: %v\n", err)
		return
	}
	fmt.Printf("Added link from '%s' to '%s'\n", *cmd.From, *cmd.To)
}

func handleAddEventCommand(cmd *AddEventCommand) {
	if err := cmd.Validate(); err != nil {
		fmt.Printf("Error validating add event command: %v\n", err)
		return
	}

	if _, err := telemetry.AddEventToSpan(*cmd.SpanName, *cmd.EventName); err != nil {
		fmt.Printf("Error adding event to span: %v\n", err)
		return
	}
	fmt.Printf("Added event '%s' to span '%s'\n", *cmd.EventName, *cmd.SpanName)
}

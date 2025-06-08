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

	_, err := telemetry.AddLinkToSpan(*cmd.From, *cmd.To)
	if err != nil {
		fmt.Printf("Error adding link: %v\n", err)
		return
	}
	fmt.Printf("Added link from '%s' to '%s'\n", *cmd.From, *cmd.To)
}

package executor

import (
	"fmt"
	"os"
	"strings"
)

func Executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	cmd, err := ParseCommand(input)
	if err != nil {
		fmt.Printf("Error parsing command: %v\n", err)
		return
	}

	switch {
	case cmd.Exit != nil:
		fmt.Println("Bye!")
		os.Exit(0)
	case cmd.Create != nil:
		handleCreateCommand(cmd.Create)
	case cmd.Set != nil:
		handleSetCommand(cmd.Set)
	case cmd.AddLink != nil:
		handleAddLinkCommand(cmd.AddLink)
	case cmd.AddEvent != nil:
		handleAddEventCommand(cmd.AddEvent)
	case cmd.Send != nil:
		handleSendCommand()
	case cmd.List != nil:
		handleListCommand(cmd.List)
	default:
		fmt.Printf("Unknown command: %v\n", cmd)
	}
}

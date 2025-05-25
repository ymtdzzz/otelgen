package executor

import (
	"fmt"
	"os"
	"strings"
)

const (
	CMD_EXIT   = "exit"
	CMD_CREATE = "create"
	CMD_ADD    = "add"
	CMD_SEND   = "send"
	CMD_LIST   = "list"
)

func Executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	args := strings.Split(input, " ")
	switch args[0] {
	case CMD_EXIT:
		println("Bye!")
		os.Exit(0)
	case CMD_CREATE:
		handleCreateCommand(args[1:])
	case CMD_ADD:
		handleAddCommand(args[1:])
	case CMD_SEND:
		handleSendCommand()
	case CMD_LIST:
		handleListCommand(args[1:])
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
	}
}

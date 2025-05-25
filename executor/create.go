package executor

import (
	"fmt"

	"github.com/ymtdzzz/otelgen/telemetry"
)

func handleCreateCommand(args []string) {
	if len(args) != 2 {
		println("Usage: create trace [name]")
		return
	}

	switch args[0] {
	case "trace":
		handleCreateTrace(args[1])
		fmt.Printf("Created trace: %s\n", args[1])
	default:
		println("Unknown create command:", args[0])
	}
}

func handleCreateTrace(name string) {
	telemetry.CreateTrace(name)
}

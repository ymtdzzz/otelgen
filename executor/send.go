package executor

import "github.com/ymtdzzz/otelgen/telemetry"

func handleSendCommand() {
	telemetry.SendAllTraces()
}

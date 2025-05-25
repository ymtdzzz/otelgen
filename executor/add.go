package executor

import (
	"fmt"
	"strings"

	"github.com/ymtdzzz/otelgen/telemetry"
)

const (
	addCommandUsage = "Usage: add span to [trace | span] [target name] [span name] [key=value,...]"
)

func handleAddCommand(args []string) {
	// span to trace target-trace span-name key=value,key=value,...
	if len(args) != 6 {
		println(addCommandUsage)
		return
	}

	switch args[0] {
	case "span":
		handleAddSpanCommand(args[2:])
	default:
		println("Unknown add command:", args[0])
	}
}

func handleAddSpanCommand(args []string) {
	// trace target-trace span-name key=value,key=value,...
	if len(args) != 4 {
		println(addCommandUsage)
		return
	}

	switch args[0] {
	case "trace":
		err := addSpanToTrace(args[1], args[2], args[3])
		if err != nil {
			fmt.Printf("Error adding span to trace: %v\n", err)
		}
	case "span":
		err := addSpanToSpan(args[1], args[2], args[3])
		if err != nil {
			fmt.Printf("Error adding span to span: %v\n", err)
		}
	default:
		fmt.Printf("Unknown target type: %s\n", args[0])
	}
}

func parseAttributes(attrsStr string) (map[string]string, error) {
	attrs := make(map[string]string)
	if attrsStr != "" {
		for attr := range strings.SplitSeq(attrsStr, ",") {
			kv := strings.SplitN(attr, "=", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("invalid attribute format: %s", attr)
			}
			attrs[kv[0]] = kv[1]
		}
	}
	return attrs, nil
}

func addSpanToTrace(target, spanName, attrsStr string) error {
	attrs, err := parseAttributes(attrsStr)
	if err != nil {
		return err
	}

	// TODO: add resource attributes
	err = telemetry.AddSpanToTrace(target, spanName, attrs)
	if err != nil {
		return err
	}

	fmt.Printf("Added span '%s' to trace '%s' with attributes: %v\n", spanName, target, attrs)
	return nil
}

func addSpanToSpan(parentSpan, spanName, attrsStr string) error {
	attrs, err := parseAttributes(attrsStr)
	if err != nil {
		return err
	}

	// TODO: add resource attributes
	err = telemetry.AddSpanToSpan(parentSpan, spanName, attrs)
	if err != nil {
		return err
	}

	fmt.Printf("Added span '%s' to parent span '%s' with attributes: %v\n", spanName, parentSpan, attrs)
	return nil
}

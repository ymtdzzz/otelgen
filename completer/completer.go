package completer

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/ymtdzzz/otelgen/executor"
	"github.com/ymtdzzz/otelgen/telemetry"
)

var CommandSuggestions = map[string][]prompt.Suggest{
	"": {
		{Text: "create", Description: "Create a new signal"},
		{Text: "add", Description: "Add a signal to another signal"},
		{Text: "send", Description: "Send all traces to the collector"},
		{Text: "list", Description: "List available traces and spans"},
		{Text: "exit", Description: "Exit the application"},
	},
	"create": {
		{Text: "trace", Description: "Create a new trace with given name"},
	},
	"add": {
		{Text: "span", Description: "Add a span to a trace or another span"},
	},
	"add span": {
		{Text: "to", Description: "Add a span to a trace or another span"},
	},
	"add span to": {
		{Text: "trace", Description: "Add a span to a trace"},
		{Text: "span", Description: "Add span to another span"},
	},
	"list": {
		{Text: "traces", Description: "List all available traces"},
	},
}

func Completer(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()

	args := strings.Split(strings.TrimSpace(text), " ")

	if len(text) == 0 {
		return CommandSuggestions[""]
	}

	if !strings.Contains(text, " ") {
		return prompt.FilterHasPrefix(CommandSuggestions[""], text, false)
	}

	switch args[0] {
	case executor.CMD_CREATE:
		if len(args) == 1 || (len(args) == 2 && !strings.HasSuffix(text, " ")) {
			txt := args[len(args)-1]
			if strings.HasSuffix(text, " ") {
				txt = ""
			}
			return prompt.FilterHasPrefix(CommandSuggestions["create"], txt, false)
		}

	case executor.CMD_ADD:
		if len(args) == 1 || len(args) == 2 && !strings.HasSuffix(text, " ") {
			txt := args[len(args)-1]
			if strings.HasSuffix(text, " ") {
				txt = ""
			}
			return prompt.FilterHasPrefix(CommandSuggestions["add"], txt, false)
		}

		if len(args) == 2 || (len(args) == 3 && !strings.HasSuffix(text, " ")) {
			txt := args[len(args)-1]
			if strings.HasSuffix(text, " ") {
				txt = ""
			}
			return prompt.FilterHasPrefix(CommandSuggestions["add span"], txt, false)
		}

		if len(args) == 3 || (len(args) == 4 && !strings.HasSuffix(text, " ")) {
			txt := args[len(args)-1]
			if strings.HasSuffix(text, " ") {
				txt = ""
			}
			return prompt.FilterHasPrefix(CommandSuggestions["add span to"], txt, false)
		}

		if len(args) == 4 || (len(args) == 5 && !strings.HasSuffix(text, " ")) {
			txt := args[len(args)-1]
			if strings.HasSuffix(text, " ") {
				txt = ""
			}
			var suggestions []prompt.Suggest
			if args[3] == "trace" {
				for traceName := range telemetry.GetTraces() {
					suggestions = append(suggestions, prompt.Suggest{Text: traceName})
				}
			}
			if args[3] == "span" {
				for spanName := range telemetry.GetSpans() {
					suggestions = append(suggestions, prompt.Suggest{Text: spanName})
				}
			}
			return prompt.FilterHasPrefix(suggestions, txt, false)
		}

	case executor.CMD_LIST:
		if len(args) == 1 || (len(args) == 2 && !strings.HasSuffix(text, " ")) {
			txt := args[len(args)-1]
			if strings.HasSuffix(text, " ") {
				txt = ""
			}
			return prompt.FilterHasPrefix(CommandSuggestions["list"], txt, false)
		}
	}

	return []prompt.Suggest{}
}

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
		{Text: "send", Description: "Send all traces to the collector"},
		{Text: "list", Description: "List available traces and spans"},
		{Text: "exit", Description: "Exit the application"},
	},
	"create_type": {
		{Text: "resource", Description: "Create a new resource"},
		{Text: "span", Description: "Create a new span"},
	},
	"create_in_or_with": {
		{Text: "in", Description: "Create a span in a trace"},
		{Text: "with", Description: "Create a span with a parent span"},
	},
	"create_in_trace": {
		{Text: "trace", Description: "Create a span in a trace"},
	},
	"create_with_parent": {
		{Text: "parent", Description: "Create a span with a parent span"},
	},
	"list": {
		{Text: "traces", Description: "List all available traces"},
	},
}

type completerContext struct {
	inputText    string
	currentWord  string
	parsed       *executor.Command
	partialInput []string
}

func (c *completerContext) completeCreate() []prompt.Suggest {
	if c.parsed.Create.Type == nil {
		return prompt.FilterHasPrefix(CommandSuggestions["create_type"], c.currentWord, false)
	}
	switch *c.parsed.Create.Type {
	case "span":
		return c.completeCreateSpan()
	case "resource":
		return c.completeCreateResource()
	}
	return []prompt.Suggest{}
}

func (c *completerContext) completeCreateSpan() []prompt.Suggest {
	if c.parsed.Create.Trace == nil && c.parsed.Create.ParentSpan == nil && c.parsed.Create.Name != nil {
		if strings.Contains(c.inputText, "in trace ") {
			var suggestions []prompt.Suggest
			for traceName := range telemetry.GetTraces() {
				suggestions = append(suggestions, prompt.Suggest{Text: traceName})
			}
			return prompt.FilterHasPrefix(suggestions, c.currentWord, false)
		} else if strings.Contains(c.inputText, "in ") {
			return prompt.FilterHasPrefix(CommandSuggestions["create_in_trace"], c.currentWord, false)
		}
		if strings.Contains(c.inputText, "with parent ") {
			var suggestions []prompt.Suggest
			for spanName := range telemetry.GetSpans() {
				suggestions = append(suggestions, prompt.Suggest{Text: spanName})
			}
			return prompt.FilterHasPrefix(suggestions, c.currentWord, false)
		} else if strings.Contains(c.inputText, "with ") {
			return prompt.FilterHasPrefix(CommandSuggestions["create_with_parent"], c.currentWord, false)
		}
		return prompt.FilterHasPrefix(CommandSuggestions["create_in_or_with"], c.currentWord, false)
	}
	// create span span-a in trace tra...
	if c.parsed.Create.Trace != nil && c.partialInput[len(c.partialInput)-2] == "trace" && !strings.HasSuffix(c.inputText, " ") {
		var suggestions []prompt.Suggest
		for traceName := range telemetry.GetTraces() {
			suggestions = append(suggestions, prompt.Suggest{Text: traceName})
		}
		return prompt.FilterHasPrefix(suggestions, c.currentWord, false)
	}
	// create span span-b with parent span sp...
	if c.parsed.Create.ParentSpan != nil && c.partialInput[len(c.partialInput)-2] == "parent" && !strings.HasSuffix(c.inputText, " ") {
		var suggestions []prompt.Suggest
		for spanName := range telemetry.GetSpans() {
			suggestions = append(suggestions, prompt.Suggest{Text: spanName})
		}
		return prompt.FilterHasPrefix(suggestions, c.currentWord, false)
	}
	if c.parsed.Create.Trace != nil || c.parsed.Create.ParentSpan != nil {
		suggestions := []prompt.Suggest{}
		if c.parsed.Create.Attrs == nil {
			if strings.Contains(c.inputText, "attributes ") {
				return []prompt.Suggest{}
			} else {
				suggestions = append(suggestions, prompt.Suggest{Text: "attributes", Description: "Add attributes to the span"})
			}
		}
		if c.parsed.Create.Resource == nil {
			if strings.Contains(c.inputText, "resource ") {
				suggestions := []prompt.Suggest{}
				for resourceName := range telemetry.GetResources() {
					suggestions = append(suggestions, prompt.Suggest{Text: resourceName})
				}
				return prompt.FilterHasPrefix(suggestions, c.currentWord, false)
			} else {
				suggestions = append(suggestions, prompt.Suggest{Text: "resource", Description: "Set a resource for the span"})
			}
		}

		return prompt.FilterHasPrefix(suggestions, c.currentWord, false)
	}
	return []prompt.Suggest{}
}

func (c *completerContext) completeCreateResource() []prompt.Suggest {
	if c.parsed.Create.Attrs == nil && c.parsed.Create.Name != nil {
		if strings.Contains(c.inputText, "attributes ") {
			return []prompt.Suggest{}
		} else {
			return prompt.FilterHasPrefix([]prompt.Suggest{
				{Text: "attributes", Description: "Add attributes to the resource"},
			}, c.currentWord, false)
		}
	}
	return []prompt.Suggest{}
}

func (c *completerContext) completeList() []prompt.Suggest {
	if c.parsed.List.Target == nil {
		return prompt.FilterHasPrefix(CommandSuggestions["list"], c.currentWord, false)
	}
	return []prompt.Suggest{}
}

func Completer(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	words := strings.Fields(text)

	cmd, _ := executor.ParseCommand(text)

	cctx := &completerContext{
		inputText:    text,
		currentWord:  d.GetWordBeforeCursor(),
		parsed:       cmd,
		partialInput: words,
	}

	if len(cctx.partialInput) == 0 {
		return CommandSuggestions[""]
	}

	if !strings.Contains(text, " ") {
		return prompt.FilterHasPrefix(CommandSuggestions[""], text, false)
	}

	if cctx.parsed == nil {
		return []prompt.Suggest{}
	}

	switch {
	case cctx.parsed.Create != nil:
		return cctx.completeCreate()
	case cctx.parsed.List != nil:
		return cctx.completeList()
	}

	return []prompt.Suggest{}
}

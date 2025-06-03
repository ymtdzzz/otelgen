package completer

import (
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/ymtdzzz/otelgen/executor"
	"github.com/ymtdzzz/otelgen/telemetry"
)

var commandSuggestions = map[string][]prompt.Suggest{
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
		return prompt.FilterHasPrefix(commandSuggestions["create_type"], c.currentWord, false)
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
			return prompt.FilterHasPrefix(convertTracesToSuggestions(), c.currentWord, false)
		} else if strings.Contains(c.inputText, "in ") {
			return prompt.FilterHasPrefix(commandSuggestions["create_in_trace"], c.currentWord, false)
		}
		if strings.Contains(c.inputText, "with parent ") {
			return prompt.FilterHasPrefix(convertSpansToSuggestions(), c.currentWord, false)
		} else if strings.Contains(c.inputText, "with ") {
			return prompt.FilterHasPrefix(commandSuggestions["create_with_parent"], c.currentWord, false)
		}
		return prompt.FilterHasPrefix(commandSuggestions["create_in_or_with"], c.currentWord, false)
	}
	// create span span-a in trace tra...
	if c.parsed.Create.Trace != nil && c.partialInput[len(c.partialInput)-2] == "trace" && !strings.HasSuffix(c.inputText, " ") {
		return prompt.FilterHasPrefix(convertTracesToSuggestions(), c.currentWord, false)
	}
	// create span span-b with parent span sp...
	if c.parsed.Create.ParentSpan != nil && c.partialInput[len(c.partialInput)-2] == "parent" && !strings.HasSuffix(c.inputText, " ") {
		return prompt.FilterHasPrefix(convertSpansToSuggestions(), c.currentWord, false)
	}
	if c.parsed.Create.Trace != nil || c.parsed.Create.ParentSpan != nil {
		suggestions := []prompt.Suggest{}
		if c.parsed.Create.Resource == nil {
			if strings.Contains(c.inputText, "resource ") {
				return prompt.FilterHasPrefix(convertResourcesToSuggestions(), c.currentWord, false)
			} else {
				suggestions = append(suggestions, prompt.Suggest{Text: "resource", Description: "Set a resource for the span"})
			}
		}
		// create span span-a in trace my-trace resource res...
		if c.parsed.Create.Resource != nil && c.partialInput[len(c.partialInput)-2] == "resource" && !strings.HasSuffix(c.inputText, " ") {
			return prompt.FilterHasPrefix(convertResourcesToSuggestions(), c.currentWord, false)
		}
		if c.parsed.Create.Attrs == nil {
			if strings.Contains(c.inputText, "attributes ") || c.parsed.Create.Resource != nil {
				return []prompt.Suggest{}
			} else {
				suggestions = append(suggestions, prompt.Suggest{Text: "attributes", Description: "Add attributes to the span"})
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
		return prompt.FilterHasPrefix(commandSuggestions["list"], c.currentWord, false)
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
		return commandSuggestions[""]
	}

	if !strings.Contains(text, " ") {
		return prompt.FilterHasPrefix(commandSuggestions[""], text, false)
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

func convertTracesToSuggestions() []prompt.Suggest {
	var suggestions []prompt.Suggest
	for traceName := range telemetry.GetTraces() {
		suggestions = append(suggestions, prompt.Suggest{Text: traceName})
	}
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Text < suggestions[j].Text
	})
	return suggestions
}

func convertSpansToSuggestions() []prompt.Suggest {
	var suggestions []prompt.Suggest
	for spanName := range telemetry.GetSpans() {
		suggestions = append(suggestions, prompt.Suggest{Text: spanName})
	}
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Text < suggestions[j].Text
	})
	return suggestions
}

func convertResourcesToSuggestions() []prompt.Suggest {
	var suggestions []prompt.Suggest
	for resourceName := range telemetry.GetResources() {
		suggestions = append(suggestions, prompt.Suggest{Text: resourceName})
	}
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Text < suggestions[j].Text
	})
	return suggestions
}

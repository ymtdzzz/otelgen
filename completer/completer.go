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
		{Text: "set", Description: "Update an existing signal"},
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
	"set_type": {
		{Text: "resource", Description: "Update a resource"},
		{Text: "span", Description: "Update a span"},
	},
	"set_span_operations": {
		{Text: "name", Description: "Set a new name for the span"},
		{Text: "resource", Description: "Set a resource for the span"},
	},
	"set_resource_operations": {
		{Text: "name", Description: "Set a new name for the resource"},
	},
	"list": {
		{Text: "traces", Description: "List all available traces"},
		{Text: "resources", Description: "List all available resources"},
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
	if c.isInputInProgress("trace") {
		return prompt.FilterHasPrefix(convertTracesToSuggestions(), c.currentWord, false)
	}
	// create span span-b with parent span sp...
	if c.isInputInProgress("parent") {
		return prompt.FilterHasPrefix(convertSpansToSuggestions(), c.currentWord, false)
	}
	if c.parsed.Create.Trace != nil || c.parsed.Create.ParentSpan != nil {
		if c.isInputInProgress("attributes") {
			return []prompt.Suggest{}
		}
		if c.isInputInProgress("resource") {
			return prompt.FilterHasPrefix(convertResourcesToSuggestions(), c.currentWord, false)
		}

		suggestions := []prompt.Suggest{}
		if c.parsed.Create.Resource == nil {
			suggestions = append(suggestions, prompt.Suggest{Text: "resource", Description: "Set a resource for the span"})
		}
		if c.parsed.Create.Attrs == nil && c.parsed.Create.Resource == nil {
			suggestions = append(suggestions, prompt.Suggest{Text: "attributes", Description: "Add attributes to the span"})
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

func (c *completerContext) completeSet() []prompt.Suggest {
	if c.parsed.Set.Type == nil {
		return prompt.FilterHasPrefix(commandSuggestions["set_type"], c.currentWord, false)
	}
	switch *c.parsed.Set.Type {
	case "span":
		return c.completeSetSpan()
	case "resource":
		return c.completeSetResource()
	}
	return []prompt.Suggest{}
}

func (c *completerContext) completeSetSpan() []prompt.Suggest {
	// 'set span ' or 'set span sp...'
	if (c.parsed.Set.Name == nil && strings.HasSuffix(c.inputText, " ")) || c.isInputInProgress("span") {
		return prompt.FilterHasPrefix(convertSpansToSuggestions(), c.currentWord, false)
	}
	if len(c.parsed.Set.Args) == 0 {
		if !c.isInputInProgress("name") && !c.isInputInProgress("resource") {
			return prompt.FilterHasPrefix(commandSuggestions["set_span_operations"], c.currentWord, false)
		}
	}
	if c.isInputInProgress("resource") {
		return prompt.FilterHasPrefix(convertResourcesToSuggestions(), c.currentWord, false)
	}

	suggesstions := []prompt.Suggest{}
	if !c.isInputInProgress("name") && !c.isInputInProgress("resource") {
		if !c.parsed.Set.HasArgName() {
			suggesstions = append(suggesstions, prompt.Suggest{Text: "name", Description: "Set a new name for the span"})
		}
		if !c.parsed.Set.HasArgResource() {
			suggesstions = append(suggesstions, prompt.Suggest{Text: "resource", Description: "Set a resource for the span"})
		}
	}
	return prompt.FilterHasPrefix(suggesstions, c.currentWord, false)
}

func (c *completerContext) completeSetResource() []prompt.Suggest {
	// 'set resource ' or 'set resource sv...'
	if (c.parsed.Set.Name == nil && strings.HasSuffix(c.inputText, " ")) || c.isInputInProgress("resource") {
		return prompt.FilterHasPrefix(convertResourcesToSuggestions(), c.currentWord, false)
	}
	if !c.isInputInProgress("name") {
		return prompt.FilterHasPrefix(commandSuggestions["set_resource_operations"], c.currentWord, false)
	}

	return []prompt.Suggest{}
}

func (c *completerContext) completeList() []prompt.Suggest {
	if c.parsed.List.Type == nil {
		return prompt.FilterHasPrefix(commandSuggestions["list"], c.currentWord, false)
	}
	return []prompt.Suggest{}
}

func (c *completerContext) isInputInProgress(cmd string) bool {
	return (c.partialInput[len(c.partialInput)-2] == cmd && !strings.HasSuffix(c.inputText, " ")) ||
		(c.partialInput[len(c.partialInput)-1] == cmd && strings.HasSuffix(c.inputText, " "))
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
	case cctx.parsed.Set != nil:
		return cctx.completeSet()
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

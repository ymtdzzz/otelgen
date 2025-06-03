package executor

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/ymtdzzz/otelgen/telemetry"
)

type Command struct {
	Create *CreateCommand `@@`
	List   *ListCommand   `| @@`
	Send   *SendCommand   `| @@`
	Exit   *ExitCommand   `| @@`
}

type ExitCommand struct {
	Exit string `@"exit"`
}

type CreateCommand struct {
	Create     string      `"create"`
	Type       *string     `[ @("resource" | "span") ]`
	Name       *string     `[ @Ident ]`
	Trace      *string     `[ "in" "trace" @Ident ]`
	ParentSpan *string     `[ "with" "parent" @Ident ]`
	Attrs      []*KeyValue `[ "attributes" @@ { "," @@ } ]`
	Resource   *string     `[ "resource" @Ident ]`
}

func (c *CreateCommand) Validate() error {
	if c.Type == nil || c.Name == nil {
		return fmt.Errorf("type and name must be specified for create command")
	}

	switch *c.Type {
	case "span":
		if c.Trace == nil && c.ParentSpan == nil {
			return fmt.Errorf("span must be created in a trace or with a parent span")
		}
		if c.Trace != nil && c.ParentSpan != nil {
			return fmt.Errorf("span cannot have both a trace and a parent span")
		}
		if c.ParentSpan != nil {
			if !telemetry.IsSpanExists(*c.ParentSpan) {
				return fmt.Errorf("parent span '%s' does not exist", *c.ParentSpan)
			}
		}
		if c.Resource != nil {
			if !telemetry.IsResourceExists(*c.Resource) {
				return fmt.Errorf("resource '%s' does not exist", *c.Resource)
			}
		}
	case "resource":
		return nil
	default:
		return fmt.Errorf("unsupported type: %s", *c.Type)
	}

	return nil
}

type ListCommand struct {
	List   string  `"list"`
	Target *string `[ @("traces" | "resources") ]`
}

type SendCommand struct {
	Send string `"send"`
}

type KeyValue struct {
	Key   string `@Ident "="`
	Value string `@Ident`
}

func convertKeyValuesToMap(attrs []*KeyValue) map[string]string {
	attrsMap := make(map[string]string)
	for _, kv := range attrs {
		if kv != nil {
			attrsMap[kv.Key] = kv.Value
		}
	}
	return attrsMap
}

var (
	commandLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Comment", `#[^\n]*`},
		{"Whitespace", `\s+`},
		{"String", `"[^"]*"|'[^']*'`},
		{"Number", `[-+]?\d+(\.\d+)?`},
		{"Ident", `[a-zA-Z_][a-zA-Z0-9_\.\-]*`},
		{"Punct", `[,=]`},
	})

	parser = participle.MustBuild[Command](
		participle.Lexer(commandLexer),
		participle.Elide("Comment", "Whitespace"),
	)
)

func ParseCommand(input string) (*Command, error) {
	return parser.ParseString("", input)
}

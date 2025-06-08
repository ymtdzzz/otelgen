package executor

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/ymtdzzz/otelgen/telemetry"
)

type Command struct {
	Create *CreateCommand `@@`
	Set    *SetCommand    `| @@`
	List   *ListCommand   `| @@`
	Send   *SendCommand   `| @@`
	Exit   *ExitCommand   `| @@`
}

type ExitCommand struct {
	Exit string `@"exit"`
}

type CreateSetArg struct {
	Resource *string `("resource" @Ident)`
	// Attrs    []*KeyValue `| ("attributes" @@ { "," @@ } )`
}

func (arg *CreateSetArg) Validate(t string) error {
	var resource string

	if arg.Resource != nil {
		resource = *arg.Resource
	}

	switch t {
	case "span":
		if resource != "" && !telemetry.IsResourceExists(resource) {
			return fmt.Errorf("resource '%s' does not exist", resource)
		}
	case "resource":
		if resource != "" {
			return errors.New("resource cannot be specified when the type is resource")
		}
	}

	return nil
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

type SetOnlyArg struct {
	Name *string `("name" @Ident)`
}

func (arg *SetOnlyArg) Validate() error {
	var name string

	if arg.Name != nil {
		name = *arg.Name
	}

	if name != "" {
		if _, exists := telemetry.GetSpans()[name]; exists {
			return fmt.Errorf("span with name %s already exists", name)
		}
	}

	return nil
}

type SetArg struct {
	SetCreateArg *CreateSetArg `@@`
	SetOnlyArg   *SetOnlyArg   `| @@`
}

func (arg *SetArg) Validate(t string) error {
	if arg.SetCreateArg != nil {
		return arg.SetCreateArg.Validate(t)
	}
	if arg.SetOnlyArg != nil {
		return arg.SetOnlyArg.Validate()
	}
	return nil
}

type SetCommand struct {
	Set  string    `"set"`
	Type *string   `[ @("resource" | "span") ]`
	Name *string   `[ @Ident ]`
	Args []*SetArg `@@*`
}

func (s *SetCommand) Validate() error {
	if s.Type == nil || s.Name == nil {
		return fmt.Errorf("type and name must be specified for set command")
	}

	if len(s.Args) == 0 {
		return fmt.Errorf("operation (name, resource, attributes etc.) must be specified for set command")
	}

	switch *s.Type {
	case "span":
		if _, exists := telemetry.GetSpans()[*s.Name]; !exists {
			return fmt.Errorf("span '%s' does not exist", *s.Name)
		}
	case "resource":
		if _, exists := telemetry.GetResources()[*s.Name]; !exists {
			return fmt.Errorf("resource '%s' does not exist", *s.Name)
		}
	}

	// check duplicates in operations
	seen := make(map[string]bool)
	for _, arg := range s.Args {
		if arg.SetCreateArg != nil {
			if arg.SetCreateArg.Resource != nil {
				opName := "resource"
				if seen[opName] {
					return fmt.Errorf("duplicate operation: %s", opName)
				}
				seen[opName] = true
			}
		}
		if arg.SetOnlyArg != nil {
			if arg.SetOnlyArg.Name != nil {
				opName := "name"
				if seen[opName] {
					return fmt.Errorf("duplicate operation: %s", opName)
				}
				seen[opName] = true
			}
		}
	}

	for _, arg := range s.Args {
		if err := arg.Validate(*s.Type); err != nil {
			return err
		}
	}

	return nil
}

func (s *SetCommand) HasArgName() bool {
	for _, arg := range s.Args {
		if arg.SetOnlyArg != nil && arg.SetOnlyArg.Name != nil {
			return true
		}
	}
	return false
}

func (s *SetCommand) HasArgResource() bool {
	for _, arg := range s.Args {
		if arg.SetCreateArg != nil && arg.SetCreateArg.Resource != nil {
			return true
		}
	}
	return false
}

type SetOperationCommandOld struct {
	Name *string `("name" @Ident)`
	// Attrs    []*KeyValue `| ("attributes" @@ { "," @@ } )`
	Resource *string `| ("resource" @Ident)`
}

type ListCommand struct {
	List string  `"list"`
	Type *string `[ @("traces" | "resources") ]`
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

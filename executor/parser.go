package executor

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/ymtdzzz/otelgen/telemetry"
)

type Command struct {
	Create   *CreateCommand   `parser:"@@"`
	Set      *SetCommand      `parser:"| @@"`
	AddLink  *AddLinkCommand  `parser:"| @@"`
	AddEvent *AddEventCommand `parser:"| @@"`
	List     *ListCommand     `parser:"| @@"`
	Send     *SendCommand     `parser:"| @@"`
	Exit     *ExitCommand     `parser:"| @@"`
}

type ExitCommand struct {
	Exit string `parser:"@'exit'"`
}

type CreateSetArg struct {
	Resource *string     `parser:"('resource' @Ident)"`
	Attrs    []*KeyValue `parser:"| ('attributes' @@ { ',' @@ } )"`
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

func (arg *CreateSetArg) addOps(ops []string) []string {
	if arg.Resource != nil {
		ops = append(ops, "resource")
	}
	if len(arg.Attrs) > 0 {
		ops = append(ops, "attributes")
	}
	return ops
}

type CreateCommand struct {
	Create     string          `parser:"'create'"`
	Type       *string         `parser:"[ @('resource'| 'span' | 'event') ]"`
	Name       *string         `parser:"[ @Ident ]"`
	Trace      *string         `parser:"[ 'in' 'trace' @Ident ]"`
	ParentSpan *string         `parser:"[ 'with' 'parent' @Ident ]"`
	Args       []*CreateSetArg `parser:"@@*"`
}

func (c *CreateCommand) Validate() error {
	if c.Type == nil || c.Name == nil {
		return fmt.Errorf("type and name must be specified for create command")
	}

	var ops []string
	for _, arg := range c.Args {
		ops = arg.addOps(ops)
	}
	if err := checkDuplicateOps(ops); err != nil {
		return err
	}

	if *c.Type == "span" {
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
	}

	for _, arg := range c.Args {
		if err := arg.Validate(*c.Type); err != nil {
			return err
		}
	}

	return nil
}

func (c *CreateCommand) HasArgAttrs() bool {
	for _, arg := range c.Args {
		if len(arg.Attrs) > 0 {
			return true
		}
	}
	return false
}

func (c *CreateCommand) HasArgResource() bool {
	for _, arg := range c.Args {
		if arg.Resource != nil {
			return true
		}
	}
	return false
}

type SetOnlyArg struct {
	Name *string `parser:"('name' @Ident)"`
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
	SetCreateArg *CreateSetArg `parser:"@@"`
	SetOnlyArg   *SetOnlyArg   `parser:"| @@"`
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

func (arg *SetArg) addOps(ops []string) []string {
	if arg.SetCreateArg != nil {
		ops = arg.SetCreateArg.addOps(ops)
	}
	if arg.SetOnlyArg != nil && arg.SetOnlyArg.Name != nil {
		ops = append(ops, "name")
	}
	return ops
}

type SetCommand struct {
	Set  string    `parser:"'set'"`
	Type *string   `parser:"[ @('resource' | 'span') ]"`
	Name *string   `parser:"[ @Ident ]"`
	Args []*SetArg `parser:"@@*"`
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

	var ops []string
	for _, arg := range s.Args {
		ops = arg.addOps(ops)
	}
	if err := checkDuplicateOps(ops); err != nil {
		return err
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

type AddLinkArg struct {
	Attrs []*KeyValue `parser:"('attributes' @@ { ',' @@ } )"`
}

func (arg *AddLinkArg) addOps(ops []string) []string {
	if len(arg.Attrs) > 0 {
		ops = append(ops, "attributes")
	}
	return ops
}

type AddLinkCommand struct {
	Add  string        `parser:"'add'"`
	Link string        `parser:"'link'"`
	From *string       `parser:"[ @Ident ]"`
	To   *string       `parser:"[ @Ident ]"`
	Args []*AddLinkArg `parser:"@@*"`
}

func (c *AddLinkCommand) Validate() error {
	if c.From == nil || c.To == nil {
		return fmt.Errorf("both 'from' and 'to' must be specified for add link command")
	}

	var ops []string
	for _, arg := range c.Args {
		ops = arg.addOps(ops)
	}
	if err := checkDuplicateOps(ops); err != nil {
		return err
	}

	if !telemetry.IsSpanExists(*c.From) {
		return fmt.Errorf("span '%s' does not exist", *c.From)
	}

	if !telemetry.IsSpanExists(*c.To) {
		return fmt.Errorf("span '%s' does not exist", *c.To)
	}

	return nil
}

func (c *AddLinkCommand) HasArgAttrs() bool {
	for _, arg := range c.Args {
		if len(arg.Attrs) > 0 {
			return true
		}
	}
	return false
}

type AddEventCommand struct {
	Add       string  `parser:"'add'"`
	Event     string  `parser:"'event'"`
	SpanName  *string `parser:"[ @Ident ]"`
	EventName *string `parser:"[ @Ident ]"`
}

func (c *AddEventCommand) Validate() error {
	if c.EventName == nil {
		return fmt.Errorf("event name must be specified for add event command")
	}

	if !telemetry.IsSpanExists(*c.SpanName) {
		return fmt.Errorf("span '%s' does not exist", *c.SpanName)
	}

	if !telemetry.IsEventExists(*c.EventName) {
		return fmt.Errorf("event '%s' does not exist", *c.EventName)
	}

	return nil
}

type ListCommand struct {
	List string  `parser:"'list'"`
	Type *string `parser:"[ @('traces' | 'resources') ]"`
}

type SendCommand struct {
	Send string `parser:"'send'"`
}

type KeyValue struct {
	Key   string `parser:"@Ident '='"`
	Value string `parser:"@Ident"`
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
		{Name: "Comment", Pattern: `#[^\n]*`},
		{Name: "Whitespace", Pattern: `\s+`},
		{Name: "String", Pattern: `"[^"]*"|'[^']*'`},
		{Name: "Number", Pattern: `[-+]?\d+(\.\d+)?`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_\.\-]*`},
		{Name: "Punct", Pattern: `[,=]`},
	})

	parser = participle.MustBuild[Command](
		participle.Lexer(commandLexer),
		participle.Elide("Comment", "Whitespace"),
	)
)

func ParseCommand(input string) (*Command, error) {
	return parser.ParseString("", input)
}

func checkDuplicateOps(ops []string) error {
	seen := make(map[string]bool)
	for _, op := range ops {
		if seen[op] {
			return fmt.Errorf("duplicated operation: %s", op)
		}
		seen[op] = true
	}
	return nil
}

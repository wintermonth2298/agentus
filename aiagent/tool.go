package aiagent

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type Tool interface {
	Execute(args json.RawMessage) (string, error)
	Name() string
	Desc() string
	Params() []Param
}

func MustNewTool[T any](
	name string,
	desc string,
	params []Param,
	action func(T) (string, error),
) Tool {
	tool, err := NewTool(name, desc, params, action)
	if err != nil {
		panic(err)
	}

	return tool
}

func NewTool[T any](
	name string,
	desc string,
	params []Param,
	action func(T) (string, error),
) (Tool, error) {
	tool := &genericTool[T]{
		NameStr:    name,
		DescStr:    desc,
		ParamsList: params,
		Action:     action,
	}
	if err := validateTool(tool); err != nil {
		return nil, fmt.Errorf("validate tool: %w", err)
	}

	return tool, nil
}

type Param struct {
	// Name must match the struct field’s "json" tag,
	// or the default JSON key used by encoding/json.
	Name        string
	Type        ParamType
	Description string
	Required    bool
	Enum        []any
	Items       *Param
	Properties  map[string]Param
}

type ParamType uint

const (
	ParamTypeString ParamType = iota
	ParamTypeInteger
	ParamTypeNumber
	ParamTypeBoolean
	ParamTypeObject
	ParamTypeArray
)

type genericTool[T any] struct {
	NameStr    string
	DescStr    string
	ParamsList []Param
	Action     func(args T) (string, error)
}

func validateTool[T any](tool *genericTool[T]) error {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	validNames := extractJSONNames(typ)

	for _, p := range tool.ParamsList {
		if !validNames[p.Name] {
			return fmt.Errorf(
				"Param.Name %q does not match any JSON field (tag or default) in type %T",
				p.Name, *new(T),
			)
		}
	}

	return nil
}

func extractJSONNames(typ reflect.Type) map[string]bool {
	names := make(map[string]bool)
	for i := range typ.NumField() {
		f := typ.Field(i)
		if !isExported(f) {
			continue
		}
		if name, ok := jsonFieldName(f); ok {
			names[name] = true
		}
	}
	return names
}

func isExported(f reflect.StructField) bool {
	return f.PkgPath == ""
}

func jsonFieldName(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return "", false
	}

	parts := strings.Split(tag, ",")
	if parts[0] != "" {
		return parts[0], true
	}
	return defaultJSONName(f.Name), true
}

func defaultJSONName(fieldName string) string {
	runes := []rune(fieldName)
	if len(runes) == 0 {
		return ""
	}
	if len(runes) > 1 && unicode.IsUpper(runes[0]) && unicode.IsUpper(runes[1]) {
		return fieldName
	}
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func (g *genericTool[T]) Name() string    { return g.NameStr }
func (g *genericTool[T]) Desc() string    { return g.DescStr }
func (g *genericTool[T]) Params() []Param { return g.ParamsList }

func (g *genericTool[T]) Execute(raw json.RawMessage) (string, error) {
	var args T
	_ = json.Unmarshal(raw, &args)

	return g.Action(args)
}

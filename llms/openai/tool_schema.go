package openai

import (
	"encoding/json"
	"fmt"

	"github.com/wintermonth2298/agentus/aiagent"
)

type jsonToolSchema map[string]any

func buildSchema(params []aiagent.Param) (json.RawMessage, error) {
	requiredFields := getRequired(params)
	properties := getProperties(params)

	schema := jsonToolSchema{
		"type":                 "object",
		"properties":           properties,
		"required":             requiredFields,
		"additionalProperties": false,
	}

	return marshalWithIndent(schema)
}

func getProperties(params []aiagent.Param) jsonToolSchema {
	props := make(jsonToolSchema, len(params))
	for _, p := range params {
		props[p.Name] = buildParamSchema(p)
	}
	return props
}

func getRequired(params []aiagent.Param) []string {
	required := make([]string, 0, 1)
	for _, p := range params {
		if p.Required {
			required = append(required, p.Name)
		}
	}
	return required
}

func buildParamSchema(p aiagent.Param) jsonToolSchema {
	s := jsonToolSchema{
		"type":        typeOf(p.Type),
		"description": p.Description,
	}

	if len(p.Enum) > 0 {
		s["enum"] = p.Enum
	}

	switch {
	case p.Type == aiagent.ParamTypeArray && p.Items != nil:
		s["items"] = buildParamSchema(*p.Items)
	case p.Type == aiagent.ParamTypeObject && len(p.Properties) > 0:
		nestedProps := getNestedProperties(p.Properties)
		s["properties"] = nestedProps

		if req := getNestedRequired(p.Properties); len(req) > 0 {
			s["required"] = req
		}
	}

	return s
}

func getNestedProperties(m map[string]aiagent.Param) jsonToolSchema {
	nested := make(jsonToolSchema, len(m))
	for name, sub := range m {
		nested[name] = buildParamSchema(sub)
	}
	return nested
}

func getNestedRequired(m map[string]aiagent.Param) []string {
	required := make([]string, 0, 1)
	for name, sub := range m {
		if sub.Required {
			required = append(required, name)
		}
	}
	return required
}

func typeOf(pt aiagent.ParamType) string {
	switch pt {
	case aiagent.ParamTypeString:
		return "string"
	case aiagent.ParamTypeInteger:
		return "integer"
	case aiagent.ParamTypeNumber:
		return "number"
	case aiagent.ParamTypeBoolean:
		return "boolean"
	case aiagent.ParamTypeObject:
		return "object"
	case aiagent.ParamTypeArray:
		return "array"
	default:
		panic(fmt.Sprintf("unknown param type: %v", pt))
	}
}

func marshalWithIndent(v any) (json.RawMessage, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal schema: %w", err)
	}
	return json.RawMessage(data), nil
}

package openai

import (
	"encoding/json"
	"fmt"

	"github.com/wintermonth2298/agentus/aiagent"
)

type jsonToolSchema map[string]any

func buildSchema(params []aiagent.Param) (json.RawMessage, error) {
	schema := jsonToolSchema{
		"type":                 "object",
		"properties":           propsFromSlice(params),
		"additionalProperties": false,
		"required":             requiredFromSlice(params),
	}

	return marshalWithIndent(schema)
}

func propsFromSlice(params []aiagent.Param) jsonToolSchema {
	props := make(jsonToolSchema, len(params))
	for _, p := range params {
		props[p.Name] = toParamSchema(p)
	}

	return props
}

func requiredFromSlice(params []aiagent.Param) []string {
	req := make([]string, 0, len(params))
	for _, p := range params {
		if p.Required {
			req = append(req, p.Name)
		}
	}

	return req
}

func toParamSchema(p aiagent.Param) jsonToolSchema {
	s := jsonToolSchema{
		"type":        typeOf(p.Type),
		"description": p.Description,
	}
	if len(p.Enum) > 0 {
		s["enum"] = p.Enum
	}

	//nolint:exhaustive // no additional fields for scalar types
	switch p.Type {
	case aiagent.ParamTypeArray:
		s["items"] = toParamSchema(*p.Items)
	case aiagent.ParamTypeObject:
		s["properties"] = propsFromMap(p.Properties)
		s["additionalProperties"] = false
		s["required"] = requiredFromMap(p.Properties)
	}

	return s
}

func propsFromMap(m map[string]aiagent.Param) jsonToolSchema {
	if len(m) == 0 {
		return jsonToolSchema{}
	}
	props := make(jsonToolSchema, len(m))
	for name, sub := range m {
		props[name] = toParamSchema(sub)
	}

	return props
}

func requiredFromMap(m map[string]aiagent.Param) []string {
	req := make([]string, 0, len(m))
	for name, sub := range m {
		if sub.Required {
			req = append(req, name)
		}
	}

	return req
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

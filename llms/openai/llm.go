package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/wintermonth2298/agentus/aiagent"
)

type LLM struct {
	client *openai.Client
	model  string
	tools  []openai.Tool
}

func NewLLM(client *openai.Client, model Model) *LLM {
	return &LLM{
		client: client,
		model:  mapModel(model),
	}
}

func (a *LLM) Call(ctx context.Context, history []aiagent.Message) (aiagent.Message, error) {
	req := newChatCompletionRequest(a.model, a.tools, history)

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return aiagent.Message{}, fmt.Errorf("openai api call: %w", err)
	}

	return parseResponse(resp.Choices[0].Message), nil
}

func (a *LLM) RegisterTool(tool aiagent.Tool) {
	a.tools = append(a.tools, mapToolSpecs(tool))
}

func newChatCompletionRequest(
	model string,
	tools []openai.Tool,
	history []aiagent.Message,
) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model:    model,
		Tools:    tools,
		Messages: mapChat(history),
	}
}

func mapChat(history []aiagent.Message) []openai.ChatCompletionMessage {
	chatMapped := make([]openai.ChatCompletionMessage, 0, len(history))

	for _, m := range history {
		msg, err := mapMessage(m)
		if err != nil {
			// TODO: log error
			continue
		}
		chatMapped = append(chatMapped, msg)
	}

	return chatMapped
}

func mapRole(msgType aiagent.MessageType) string {
	switch msgType {
	case aiagent.MessageTypeSystem:
		return openai.ChatMessageRoleSystem
	case aiagent.MessageTypeUser:
		return openai.ChatMessageRoleUser
	case aiagent.MessageTypeAssistant:
		return openai.ChatMessageRoleAssistant
	case aiagent.MessageTypeToolRequest:
		return openai.ChatMessageRoleAssistant
	case aiagent.MessageTypeToolResponse:
		return openai.ChatMessageRoleTool
	}

	return openai.ChatMessageRoleUser
}

func mapToolSpecs(tool aiagent.Tool) openai.Tool {
	schema, err := buildSchema(tool.Params())
	if err != nil {
		panic(fmt.Errorf("build schema: %w", err))
	}

	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        tool.Name(),
			Description: tool.Desc(),
			Strict:      false,
			Parameters:  schema,
		},
	}
}

func mapModel(model Model) string {
	switch model {
	case ModelGPT3Dot5Turbo0125:
		return openai.GPT3Dot5Turbo0125
	case ModelGPT4o:
		return openai.GPT4o
	}

	// TODO: log fallback functionality
	return openai.GPT3Dot5Turbo0125
}

func mapMessage(m aiagent.Message) (openai.ChatCompletionMessage, error) {
	role := mapRole(m.Type())

	switch m.Type() {
	case aiagent.MessageTypeSystem, aiagent.MessageTypeUser, aiagent.MessageTypeAssistant:
		return openai.ChatCompletionMessage{
			Role:    role,
			Content: m.MustText(),
		}, nil
	case aiagent.MessageTypeToolResponse:
		tcResp := m.MustToolCallResponse()
		return openai.ChatCompletionMessage{
			ToolCallID: tcResp.Call.ID,
			Role:       role,
			Content:    tcResp.Result,
		}, nil
	case aiagent.MessageTypeToolRequest:
		tcRequests := m.MustToolCallRequests()
		return openai.ChatCompletionMessage{
			Role:      role,
			ToolCalls: mapToolCalls(tcRequests),
		}, nil
	}

	return openai.ChatCompletionMessage{}, errors.New("unknown message type")
}

func mapToolCalls(tcRequests []aiagent.ToolCallRequest) []openai.ToolCall {
	tcRequestsMapped := make([]openai.ToolCall, 0, len(tcRequests))
	for _, req := range tcRequests {
		tcRequestsMapped = append(tcRequestsMapped, mapToolCall(req))
	}
	return tcRequestsMapped
}

func mapToolCall(req aiagent.ToolCallRequest) openai.ToolCall {
	return openai.ToolCall{
		ID:   req.Call.ID,
		Type: openai.ToolTypeFunction,
		Function: openai.FunctionCall{
			Name:      req.Call.Name,
			Arguments: string(req.Args),
		},
	}
}

func parseResponse(m openai.ChatCompletionMessage) aiagent.Message {
	if len(m.ToolCalls) > 0 {
		tcRequests := make([]aiagent.ToolCallRequest, 0, len(m.ToolCalls))
		for _, tc := range m.ToolCalls {
			tcRequests = append(tcRequests, parseToolCallRequest(tc))
		}
		return aiagent.NewToolCallRequestMessage(tcRequests)
	}

	return aiagent.NewAssistantMessage(m.Content)
}

func parseToolCallRequest(tc openai.ToolCall) aiagent.ToolCallRequest {
	var args json.RawMessage
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		panic("invalid tool arguments: " + err.Error())
	}

	return aiagent.ToolCallRequest{
		Call: aiagent.ToolCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
		},
		Args: args,
	}
}

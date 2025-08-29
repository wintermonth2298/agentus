package aiagent

import (
	"context"
	"errors"
	"fmt"
)

const (
	toolCallLimit = 10
)

type LLM interface {
	Call(ctx context.Context, history []Message) (Message, error)
	RegisterTool(tool Tool)
}

type Agent struct {
	llm          LLM
	toolRegistry map[string]Tool
	systemPrompt string

	debug bool
}

type AgentOption func(*Agent)

func NewAgent(llm LLM, opts ...AgentOption) *Agent {
	agent := &Agent{
		llm:          llm,
		toolRegistry: make(map[string]Tool),
		debug:        false,
	}
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func WithTool(t Tool) AgentOption {
	return func(a *Agent) {
		a.toolRegistry[t.Name()] = t
		a.llm.RegisterTool(t)
	}
}

func WithSystemPrompt(prompt string) AgentOption {
	return func(a *Agent) {
		a.systemPrompt = prompt
	}
}

func WithDebug() AgentOption {
	return func(a *Agent) {
		a.debug = true
	}
}

func WithTools(tools ...Tool) AgentOption {
	return func(a *Agent) {
		for _, t := range tools {
			a.toolRegistry[t.Name()] = t
			a.llm.RegisterTool(t)
		}
	}
}

func (a *Agent) SendMessage(ctx context.Context, userMessage string) (string, error) {
	history := a.initialHistory(userMessage)

	if a.debug {
		defer func() {
			a.printHistory(history)
		}()
	}

	for range toolCallLimit {
		resp, err := a.llm.Call(ctx, history)
		if err != nil {
			return "", fmt.Errorf("call llm: %w", err)
		}
		history = append(history, resp)
		if resp.Type() == MessageTypeAssistant {
			return resp.MustText(), nil
		}

		// llm resp msgtype != MessageTypeAssistant => msgtype == MessageTypeToolCallRequest
		for _, tcReq := range resp.MustToolCallRequests() {
			tcResponse, errExec := a.executeTool(tcReq)
			if errExec != nil {
				return "", errExec
			}
			history = append(history, tcResponse)
		}
	}

	return "", errors.New("max tool call limit exceeded")
}

func (a *Agent) executeTool(req ToolCallRequest) (Message, error) {
	tcExecutable := a.toolRegistry[req.Call.Name]

	var content string
	content, err := tcExecutable.Execute(req.Args)
	if err != nil {
		return Message{}, fmt.Errorf("call tool %s: %w", req.Call.Name, err)
	}

	return NewToolCallResponseMessage(req.Call.ID, req.Call.Name, content), nil
}

func (a *Agent) initialHistory(userMessage string) []Message {
	if a.systemPrompt == "" {
		return []Message{NewUserMessage(userMessage)}
	}
	return []Message{
		NewSystemMessage(a.systemPrompt),
		NewUserMessage(userMessage),
	}
}

func (a *Agent) printHistory(history []Message) {
	for _, msg := range history {
		fmt.Println(msg.String())
	}
}

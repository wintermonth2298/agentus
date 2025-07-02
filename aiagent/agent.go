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
}

type AgentOption func(*Agent)

func NewAgent(llm LLM, opts ...AgentOption) *Agent {
	agent := &Agent{
		llm:          llm,
		toolRegistry: make(map[string]Tool),
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

	defer func() {
		a.printHistory(history)
	}()

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
			tcResponse := a.executeTool(tcReq)
			history = append(history, tcResponse)
		}
	}

	return "", errors.New("max tool call limit exceeded")
}

func (a *Agent) executeTool(tcReq ToolCallRequest) Message {
	tcExecutable := a.toolRegistry[tcReq.Call.Name]

	var content string
	content, err := tcExecutable.Execute(tcReq.Args)
	if err != nil {
		content = fmt.Sprintf("Error: tool %s failed: %s", tcReq.Call.Name, err)
	}

	return NewToolCallResponseMessage(tcReq.Call.ID, tcReq.Call.Name, content)
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

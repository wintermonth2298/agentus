package aiagent

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fatih/color"
)

var (
	ErrNoTextContent       = errors.New("message has no text content")
	ErrNotToolCallResponse = errors.New("message is not a tool call response")
	ErrNotToolCallRequests = errors.New("message is not a tool call requests")
)

type Message struct {
	text             *string
	toolCallRequests []ToolCallRequest
	toolCallResponse *ToolCallResponse
	messageType      MessageType
}

func NewUserMessage(text string) Message {
	return Message{
		text:        &text,
		messageType: MessageTypeUser,
	}
}

func NewAssistantMessage(text string) Message {
	return Message{
		text:        &text,
		messageType: MessageTypeAssistant,
	}
}

func NewSystemMessage(text string) Message {
	return Message{
		text:        &text,
		messageType: MessageTypeSystem,
	}
}

func NewToolCallRequestMessage(requests []ToolCallRequest) Message {
	return Message{
		toolCallRequests: requests,
		messageType:      MessageTypeToolRequest,
	}
}

func NewToolCallResponseMessage(toolID string, toolName string, result string) Message {
	return Message{
		toolCallResponse: &ToolCallResponse{
			Call: ToolCall{
				ID:   toolID,
				Name: toolName,
			},
			Result: result,
		},
		messageType: MessageTypeToolResponse,
	}
}

func (m *Message) Text() (string, error) {
	if m.text == nil {
		return "", fmt.Errorf("%w (type=%s)", ErrNoTextContent, m.Type())
	}
	return *m.text, nil
}

func (m *Message) MustText() string {
	text, err := m.Text()
	if err != nil {
		panic(err)
	}
	return text
}

func (m *Message) ToolCallRequests() ([]ToolCallRequest, error) {
	if !m.IsToolCallRequest() {
		return nil, ErrNotToolCallRequests
	}

	return m.toolCallRequests, nil
}

func (m *Message) MustToolCallRequests() []ToolCallRequest {
	req, err := m.ToolCallRequests()
	if err != nil {
		panic(err)
	}

	return req
}

func (m *Message) ToolCallResponse() (ToolCallResponse, error) {
	if !m.IsToolCallResponse() {
		return ToolCallResponse{}, ErrNotToolCallResponse
	}

	return *m.toolCallResponse, nil
}

func (m *Message) MustToolCallResponse() ToolCallResponse {
	resp, err := m.ToolCallResponse()
	if err != nil {
		panic(err)
	}

	return resp
}

func (m *Message) IsToolCallRequest() bool {
	return len(m.toolCallRequests) > 0
}

func (m *Message) IsToolCallResponse() bool {
	return m.toolCallResponse != nil
}

func (m *Message) Type() MessageType {
	return m.messageType
}

func (m *Message) String() string {
	var (
		text      = ""
		textLimit = 50
	)
	var (
		magenta = color.New(color.FgYellow)
		green   = color.New(color.FgGreen)
		blue    = color.New(color.FgBlue)
	)
	switch m.messageType {
	case MessageTypeAssistant:
		text = "\n" + *m.text
	case MessageTypeSystem:
		text = *m.text
		if len(text) > textLimit {
			text = text[:textLimit] + "..."
		}
	case MessageTypeToolRequest:
		for _, req := range m.toolCallRequests {
			text += fmt.Sprintf("\n\t%s(%s)", req.Call.Name, string(req.Args))
		}
		text = green.Sprint(text)
	case MessageTypeToolResponse:
		text += m.toolCallResponse.Result
		text = blue.Sprint(text)
	case MessageTypeUser:
		text = *m.text
		if len(text) > textLimit {
			text = text[:textLimit] + "..."
		}
	}

	coloredMsgType := magenta.Sprint(m.messageType.String())
	return fmt.Sprintf("%s: %s", coloredMsgType, text)
}

type ToolCall struct {
	ID   string
	Name string
}

type ToolCallRequest struct {
	Call ToolCall
	Args json.RawMessage
}

type ToolCallResponse struct {
	Call   ToolCall
	Result string
}

type MessageType uint8

const (
	MessageTypeSystem MessageType = iota
	MessageTypeUser
	MessageTypeAssistant
	MessageTypeToolRequest
	MessageTypeToolResponse
)

func (t MessageType) String() string {
	switch t {
	case MessageTypeSystem:
		return "system"
	case MessageTypeUser:
		return "user"
	case MessageTypeAssistant:
		return "assistant"
	case MessageTypeToolRequest:
		return "tool_request"
	case MessageTypeToolResponse:
		return "tool_response"
	default:
		return fmt.Sprintf("unknown_message_type(%d)", t)
	}
}

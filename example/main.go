package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	openaicli "github.com/sashabaranov/go-openai"
	"github.com/wintermonth2298/agentus/aiagent"
	"github.com/wintermonth2298/agentus/example/tools"
	"github.com/wintermonth2298/agentus/llms/openai"
)

type authTransport struct {
	wrapped http.RoundTripper
	token   string
}

func (a *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+a.token)
	return a.wrapped.RoundTrip(req)
}

func main() {
	proxyAPIKey := os.Getenv("PROXY_API_KEY")

	client := newOpenAIClient(proxyAPIKey)

	agent := aiagent.NewAgent(
		openai.NewLLM(client, openai.ModelGPT4o),
		aiagent.WithTool(tools.NewNumbersAdder()),
		aiagent.WithTool(tools.NewRandomNumberGenerator()),
		aiagent.WithTool(tools.NewTimeReporter()),
		aiagent.WithDebug(),
		aiagent.WithSystemPrompt("If required parameters are missing, ask the user before calling"),
	)

	resp, err := agent.SendMessage(
		context.Background(),
		"Generate two random numbers from 10 to 20, add them together, and also tell me the current time.",
		aiagent.WithSystemPromptAppend("all output times are in brackets"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)
}

func newOpenAIClient(key string) *openaicli.Client {
	// reqiured for proxyAPI
	httpClient := &http.Client{
		Transport: &authTransport{
			wrapped: http.DefaultTransport,
			token:   key,
		},
	}

	return openaicli.NewClientWithConfig(openaicli.ClientConfig{
		BaseURL:    "https://api.proxyapi.ru/openai/v1",
		HTTPClient: httpClient,
	})
}

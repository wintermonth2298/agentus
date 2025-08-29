package tools

import (
	"time"

	"github.com/wintermonth2298/agentus/aiagent"
)

func NewTimeReporter() aiagent.Tool {
	return aiagent.MustNewTool(
		"time_reporter",
		"Returns the current server time in RFC3339 format.",
		aiagent.NoParams(),
		func(aiagent.NoArgs) (string, error) {
			return time.Now().Format(time.RFC3339), nil
		},
	)
}

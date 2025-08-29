package tools

import (
	"context"
	"strconv"

	"github.com/wintermonth2298/agentus/aiagent"
)

type AddArgs struct {
	A int
	B int
}

func NewNumbersAdder() aiagent.Tool {
	return aiagent.MustNewTool(
		"add_numbers",
		"Adds two integers together and returns the result.",
		[]aiagent.Param{
			{Name: "a", Type: aiagent.ParamTypeInteger, Description: "First number to add", Required: true},
			{Name: "b", Type: aiagent.ParamTypeInteger, Description: "Second number to add", Required: true},
		},
		func(_ context.Context, args AddArgs) (string, error) {
			return strconv.Itoa(args.A + args.B), nil
		},
	)
}

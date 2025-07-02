package tools

import (
	"errors"
	"math/rand/v2"
	"strconv"

	"github.com/wintermonth2298/agentus/aiagent"
)

type RandomArgs struct {
	Min int
	Max int
}

func NewRandomNumberGenerator() aiagent.Tool {
	return aiagent.MustNewTool(
		"random_number",
		"Generates a random integer number between min and max (inclusive).",
		[]aiagent.Param{
			{Name: "min", Type: aiagent.ParamTypeInteger, Description: "Minimum value (inclusive)", Required: true},
			{Name: "max", Type: aiagent.ParamTypeInteger, Description: "Maximum value (inclusive)", Required: true},
		},
		func(args RandomArgs) (string, error) {
			if args.Min > args.Max {
				return "", errors.New("min cannot be greater than max")
			}
			result := rand.IntN(args.Max-args.Min+1) + args.Min //nolint:gosec // example only
			return strconv.Itoa(result), nil
		},
	)
}

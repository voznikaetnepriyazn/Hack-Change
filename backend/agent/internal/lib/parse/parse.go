package parser

import (
	"encoding/json"
	"fmt"
)

func ParseArgs[T any](argsJSON string) (T, error) {
	var args T
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		var zero T
		return zero, fmt.Errorf("failed to parse arguments: %w", err)
	}
	return args, nil
}

package toolkit

import (
	"encoding/json"
	"strings"
)

type Tool[TArgs any] interface {
	GetArguments() TArgs
	ParseArgument(rawArgs string) error
}

type ToolArgs[TArgs any] struct {
	args TArgs
}

func (t *ToolArgs[TArgs]) ParseArgument(rawArgs string) error {
	var args TArgs
	err := json.NewDecoder(strings.NewReader(rawArgs)).Decode(&args)
	if err != nil {
		return err
	}

	t.args = args
	return nil
}

func (t *ToolArgs[TArgs]) GetArguments() TArgs {
	return t.args
}

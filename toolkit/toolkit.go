package toolkit

import "github.com/sashabaranov/go-openai"

type Parsable interface {
	ParseArgument(rawArgs string) error
}

type Executable interface {
	Execute() string
}

type Definable interface {
	Definition() openai.FunctionDefinition
}

type Callable interface {
	Executable
	Definable
}

type Toolkit struct {
	registry map[string]Callable
}

func NewToolkit() *Toolkit {
	return &Toolkit{
		registry: make(map[string]Callable),
	}
}

func (t *Toolkit) RegisterTool(tool Callable) {
	name := tool.Definition().Name
	if name == "" {
		panic("Tool name cannot be empty")
	}
	if _, exists := t.registry[name]; !exists {
		t.registry[name] = tool
	}
}

func (t *Toolkit) GetTools() []openai.Tool {
	tools := make([]openai.Tool, 0, len(t.registry))
	for _, tool := range t.registry {
		toolDef := tool.Definition()
		tools = append(tools, openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: &toolDef,
		})
	}
	return tools
}

func (t *Toolkit) GetTool(name string) (Callable, bool) {
	tool, exists := t.registry[name]
	return tool, exists
}

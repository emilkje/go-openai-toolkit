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

func (t *Toolkit) RegisterTool(tool Callable, tools ...Callable) {
	for _, callable := range append(tools, tool) {
		name := callable.Definition().Name
		if name == "" {
			panic("Tool name cannot be empty")
		}
		if _, exists := t.registry[name]; !exists {
			t.registry[name] = callable
		}
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

func (t *Toolkit) DefaultSystemMessage() string {
	toolDescriptions := ""
	for _, tool := range t.registry {
		toolDescriptions += "- " + tool.Definition().Name + ": " + tool.Definition().Description + "\n"
	}
	return "You are an assistant that has access to the following set of tools. Here are the names and descriptions for each tool:\n\n" +
		toolDescriptions + "\n" +
		"Given the user's input, you should be able to call the appropriate tools to provide the user with the information they need."
}

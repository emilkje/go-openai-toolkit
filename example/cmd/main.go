package main

import (
	"fmt"
	"os"

	toolkit_runtime "github.com/emilkje/go-openai-toolkit/runtime"
	"github.com/emilkje/go-openai-toolkit/toolkit"
	"github.com/sashabaranov/go-openai"

	"github.com/emilkje/go-openai-toolkit/example/tools"
)

//go:generate go run ../../cmd/toolkit-tools-gen/main.go -path ../tools
// this would normally be  //go:generate toolkit-tools-gen -path ./tools

func main() {
	// Create a new toolkit to hold the tools
	tk := toolkit.NewToolkit()

	// Register the tools
	tk.RegisterTool(
		tools.NewGeocodeTool(),
		tools.NewWeatherTool(),
	)

	// Create a new runtime
	client := openai.NewClientWithConfig(newConfigFromEnv())
	runtime := toolkit_runtime.NewRuntime(client, tk)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: tk.DefaultSystemMessage(), // gives a brief description of the available tools
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "How's the weather in Oslo? Please answer in a very brief sentence.",
		},
	}

	// Start the runtime to fulfill the user's request
	chatLog, err := runtime.ProcessChat(messages)

	if err != nil {
		fmt.Println("😵error processing chat:", err)
		os.Exit(1)
	}

	for _, message := range chatLog {
		prettyPrintMessage(message)
	}
}

func prettyPrintMessage(message openai.ChatCompletionMessage) {
	if len(message.ToolCalls) > 0 {
		for _, toolCall := range message.ToolCalls {
			fmt.Println(getRoleIcon(message.Role)+" "+message.Role+": called tool: ", toolCall.Function.Name, toolCall.Function.Arguments)
		}
	} else {
		fmt.Println(getRoleIcon(message.Role) + " " + message.Role + ": " + message.Content)
	}
}

func getRoleIcon(role string) string {
	switch role {
	case openai.ChatMessageRoleSystem:
		return "🤖"
	case openai.ChatMessageRoleUser:
		return "👤"
	case openai.ChatMessageRoleAssistant:
		return "🤖"
	case openai.ChatMessageRoleTool:
		return "🛠️"
	default:
		return "❓"
	}
}

func newConfigFromEnv() openai.ClientConfig {
	config := openai.DefaultAzureConfig(
		os.Getenv("AOAI_API_KEY"),
		os.Getenv("AOAI_ENDPOINT"),
	)

	config.APIVersion = os.Getenv("AOAI_API_VERSION")

	// If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
	config.AzureModelMapperFunc = func(model string) string {
		azureModelMapping := map[string]string{
			openai.GPT4TurboPreview: os.Getenv("AOAI_MODEL_DEPLOYMENT"),
		}

		return azureModelMapping[model]
	}

	return config
}

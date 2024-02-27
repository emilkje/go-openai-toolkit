package main

import (
	"fmt"
	"os"

	"github.com/emilkje/go-openai-toolkit/runtime"
	"github.com/emilkje/go-openai-toolkit/toolkit"
	"github.com/sashabaranov/go-openai"

	"github.com/emilkje/go-openai-toolkit/example/tools"
)

//go:generate go run ../../cmd/toolkit-tools-gen/main.go -path ../tools
// this would normally be  //go:generate toolkit-tools-gen -path ./tools

func main() {
	tk := toolkit.NewToolkit()
	tk.RegisterTool(tools.NewWeatherTool())

	client := openai.NewClientWithConfig(newConfigFromEnv())
	rt := runtime.NewRuntime(client, tk)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a helpful assistant",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "What is the weather in Oslo?",
		},
	}

	chatLog, err := rt.ProcessChat(messages)

	if err != nil {
		fmt.Println("error processing chat:", err)
		os.Exit(1)
	}

	for _, message := range chatLog {
		if len(message.ToolCalls) > 0 {
			for _, toolCall := range message.ToolCalls {
				fmt.Println(getRoleIcon(message.Role)+" "+message.Role+": called tool: ", toolCall.Function.Name, toolCall.Function.Arguments)
			}
		} else {
			fmt.Println(getRoleIcon(message.Role) + " " + message.Role + ": " + message.Content)
		}
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

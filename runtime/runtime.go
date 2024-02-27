package runtime

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"

	"github.com/emilkje/go-openai-toolkit/toolkit"
)

type Runtime struct {
	client  *openai.Client
	toolkit *toolkit.Toolkit
}

func NewRuntime(client *openai.Client, toolkit *toolkit.Toolkit) *Runtime {
	return &Runtime{
		client:  client,
		toolkit: toolkit,
	}
}

func (r *Runtime) ProcessChat(messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	var err error
	for {
		var response openai.ChatCompletionResponse
		response, err = r.executeChatCompletion(messages, r.toolkit.GetTools())
		if err != nil {
			return messages, err
		}

		lastMessage := response.Choices[0].Message
		messages = append(messages, lastMessage)

		if response.Choices[0].FinishReason != openai.FinishReasonToolCalls {
			break // Exit the loop if the finish reason is not due to tool calls
		}

		err = r.handleToolCalls(lastMessage.ToolCalls, &messages)
		if err != nil {
			return messages, err
		}
	}
	return messages, nil
}

func (r *Runtime) handleToolCalls(toolCalls []openai.ToolCall, messages *[]openai.ChatCompletionMessage) error {
	for _, toolCall := range toolCalls {
		toolResponse, err := r.executeTool(toolCall.Function.Name, toolCall.Function.Arguments)
		if err != nil {
			*messages = append(*messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleTool,
				Content: fmt.Sprintf("error executing tool %s: %v", toolCall.Function.Name, err),
			})
			// Consider whether to continue or return the error based on your use case
			continue // In this case, we continue to attempt other tool calls
		}

		*messages = append(*messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    toolResponse,
			ToolCallID: toolCall.ID,
		})
	}
	return nil
}

func (r *Runtime) executeTool(toolName string, args string) (string, error) {
	tool, exists := r.toolkit.GetTool(toolName)
	if !exists {
		return "", fmt.Errorf("tool %s not found", toolName)
	}

	parser, ok := tool.(toolkit.Parsable)
	if ok {
		if err := parser.ParseArgument(args); err != nil {
			return "", err
		}
	}
	return tool.Execute(), nil
}

func (r *Runtime) executeChatCompletion(messages []openai.ChatCompletionMessage, tools []openai.Tool) (openai.ChatCompletionResponse, error) {
	return r.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4TurboPreview,
			Messages: messages,
			Tools:    tools,
		},
	)
}

// Code generated with go-openai-toolkit. DO NOT EDIT.

package tools

import (
	"github.com/emilkje/go-openai-toolkit/toolkit"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func (w *WeatherTool) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "weather_tool",
		Description: "WeatherTool reports the current weather for a location",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"location": {
					Type:        jsonschema.String,
					Description: "The location to get the weather for.",
				},
			},
			Required: []string{"location"},
		},
	}
}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{&toolkit.ToolArgs[WeatherToolArgs]{}}
}

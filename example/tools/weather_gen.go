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
				"latitude": {
					Type:        jsonschema.Number,
					Description: "The latitude of the location.",
				},
				"longitude": {
					Type:        jsonschema.Number,
					Description: "The longitude of the location.",
				},
			},
			Required: []string{"latitude", "longitude"},
		},
	}
}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{&toolkit.ToolArgs[WeatherToolArgs]{}}
}

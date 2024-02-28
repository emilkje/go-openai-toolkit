// Code generated with go-openai-toolkit. DO NOT EDIT.

package tools

import (
	"github.com/emilkje/go-openai-toolkit/toolkit"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func (g *GeocodeTool) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "geocode_tool",
		Description: "Geocode tool geocodes an address and returns the latitude and longitude.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"address": {
					Type:        jsonschema.String,
					Description: "The address to geocode into latitude/longitude coordinates.",
				},
			},
			Required: []string{"address"},
		},
	}
}

func NewGeocodeTool() *GeocodeTool {
	return &GeocodeTool{&toolkit.ToolArgs[GeocodeArgs]{}}
}

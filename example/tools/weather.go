package tools

import (
	"fmt"
	"github.com/emilkje/go-openai-toolkit/toolkit"
)

type WeatherToolArgs struct {
	Location string `json:"location" desc:"The location to get the weather for."`
}

// WeatherTool is a tool that greets a person
// +tool:name=weather_tool
// +tool:description=WeatherTool reports the current weather for a location
type WeatherTool struct {
	toolkit.Tool[WeatherToolArgs]
}

func (g *WeatherTool) Execute() string {
	location := g.GetArguments().Location
	return fmt.Sprintf("It's sunny with a temperature of 21C in %s", location)
}

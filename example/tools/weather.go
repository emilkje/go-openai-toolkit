package tools

import (
	"encoding/json"
	"fmt"
	"github.com/emilkje/go-openai-toolkit/toolkit"
	"net/http"
)

type WeatherToolArgs struct {
	Latitude  float64 `json:"latitude" desc:"The latitude of the location."`
	Longitude float64 `json:"longitude" desc:"The longitude of the location."`
}

// WeatherTool is a tool that reports the current weather for a location
// +tool:name=weather_tool
// +tool:description=WeatherTool reports the current weather for a location
type WeatherTool struct {
	toolkit.Tool[WeatherToolArgs]
}

func (g *WeatherTool) Execute() string {

	client := &http.Client{}
	queryParams := fmt.Sprintf("lat=%f&lon=%f",
		g.GetArguments().Latitude,
		g.GetArguments().Longitude)

	// add Accept and User-Agent headers
	req, err := http.NewRequest("GET", "https://api.met.no/weatherapi/locationforecast/2.0/compact?"+queryParams, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "go-openai-toolkit")

	res, err := client.Do(req)

	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	if res.StatusCode != 200 {
		return fmt.Sprintf("Error: %v", res.Status)
	}

	defer res.Body.Close()
	// decode the body into a string
	data := make(map[string]interface{})
	json.NewDecoder(res.Body).Decode(&data)

	// metadata about the values is found inside .properties.meta.units as a key value pair
	// the actual values are found inside .properties.timeseries

	// lets return the metadata and the first timeseries
	metadata := data["properties"].(map[string]interface{})["meta"].(map[string]interface{})["units"].(map[string]interface{})

	// format the metadata as a string in the form of key=value\n
	metadataStr := "#metadata:\n"
	for k, v := range metadata {
		metadataStr += fmt.Sprintf("- %s=%v\n", k, v)
	}

	// get the first timeseries
	timeseries := data["properties"].(map[string]interface{})["timeseries"].([]interface{})[0]

	// format the forecast as a string in the form of key=value\n
	// timestamp is found in .time
	// the actual forecast key value paris is found in.data.instant.details
	forecastStr := "\n#forecast:\n"
	forecastStr += fmt.Sprintf("- timestamp=%s\n", timeseries.(map[string]interface{})["time"])
	forecastdata := timeseries.(map[string]interface{})["data"].(map[string]interface{})["instant"].(map[string]interface{})["details"].(map[string]interface{})
	for k, v := range forecastdata {
		forecastStr += fmt.Sprintf("- %s=%v\n", k, v)
	}

	return metadataStr + forecastStr
}

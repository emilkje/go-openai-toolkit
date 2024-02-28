package tools

import (
	"encoding/json"
	"fmt"
	"github.com/emilkje/go-openai-toolkit/toolkit"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type GeocodeArgs struct {
	Address string `json:"address" desc:"The address to geocode into latitude/longitude coordinates."`
}

// GeocodeTool is a tool that geocodes an address
// +tool:name=geocode_tool
// +tool:description=Geocode tool geocodes an address and returns the latitude and longitude.
type GeocodeTool struct{ toolkit.Tool[GeocodeArgs] }

func (g *GeocodeTool) Execute() string {

	// url encode address argument
	address := url.QueryEscape(g.GetArguments().Address)

	client := &http.Client{}
	apikey := os.Getenv("GOOGLE_API_KEY")

	res, err := client.Get("https://maps.googleapis.com/maps/api/geocode/json?address=" + address + "&key=" + apikey)

	if err != nil {
		return err.Error()
	}

	if res.StatusCode != 200 {
		return fmt.Sprintf("Error: %s", res.Status)
	}

	defer res.Body.Close()
	var result geocodeResponse
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return err.Error()
	}

	lat := result.Results[0].Geometry.Location.Lat
	lon := result.Results[0].Geometry.Location.Lng

	return fmt.Sprintf("Latitude: %s, Longitude: %s",
		strconv.FormatFloat(lat, 'f', -1, 64),
		strconv.FormatFloat(lon, 'f', -1, 64))
}

type geocodeResponse struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			}
		}
	}
}

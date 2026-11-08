package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetWeatherInput struct {
	City string `json:"city" jsonschema:"The city name"`
	// City string `json:"city" jsonschema:"required,description=The city name"`
}

type GetWeatherOutput string

func GetWeather(ctx context.Context, req *mcp.CallToolRequest, input GetWeatherInput) (
	*mcp.CallToolResult,
	GetWeatherOutput,
	error,
) {
	resp, err := http.Get(fmt.Sprintf("https://wttr.in/%s?format=3", input.City))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("non 200 status code")
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return nil, GetWeatherOutput(string(b)), nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "Yun's Server", Version: "v0.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "get_weather", Description: "Get current weather for a city"}, GetWeather)
	err := server.Run(context.Background(), &mcp.StdioTransport{})
	if err != nil {
		panic(err)
	}
}

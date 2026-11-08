package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type LoadSkillInput struct {
	Name string `json:"name" jsonschema:"The skill name"`
}

type LoadSkillOutput string

var skillDirPath string

func LoadSkill(ctx context.Context, req *mcp.CallToolRequest, input LoadSkillInput) (
	*mcp.CallToolResult,
	LoadSkillOutput,
	error,
) {
	fmt.Println("LoadSkill invoked")
	if skillDirPath == "" {
		return nil, LoadSkillOutput(""), errors.New("no skill dir path")
	}

	// read dir and build hashmap from scratch
	// return the entire string
	xs, err := os.ReadDir(skillDirPath)
	if err != nil {
		panic(err)
	}

	var res string
	for _, x := range xs {
		if x.IsDir() {
			continue
		}

		if strings.TrimSuffix(x.Name(), filepath.Ext(x.Name())) == input.Name {
			filePath := filepath.Join(skillDirPath, x.Name())

			b, err := os.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
			res = string(b)
		}
	}
	return nil, LoadSkillOutput(res), nil
}

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
	fmt.Println("GetWeather invoked")

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

	flag.StringVar(&skillDirPath, "skillDirPath", "", "The name of the skill dir path")
	flag.Parse()

	server := mcp.NewServer(&mcp.Implementation{Name: "Yun's Server", Version: "v0.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "get_weather", Description: "Get current weather for a city"}, GetWeather)
	mcp.AddTool(server, &mcp.Tool{Name: "load_skill", Description: "Load a skill's detailed instructions"}, LoadSkill)

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return server
		},
		nil,
	)

	fmt.Println("starting server")
	err := http.ListenAndServe("127.0.0.1:7777", handler)
	if err != nil {
		panic(err)
	}
}

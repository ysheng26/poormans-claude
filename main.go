package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
	"github.com/anthropics/anthropic-sdk-go/v2/toolrunner"
)

// https://github.com/anthropics/anthropic-sdk-go/blob/main/tools.md

const (
	// ModelQwen = "qwen2.5-coder:7b"
	ModelQwen = "qwen3:8b"
)

type GetWeatherInput struct {
	City  string `json:"city" jsonschema:"required,description=The city name"`
	Units string `json:"units,omitempty" jsonschema:"enum=celsius,enum=fahrenheit,description=Temperature units"`
}

func main() {
	fmt.Println("poorman's claude")
	client := anthropic.NewClient(
		option.WithAPIKey("ollama"),
		option.WithBaseURL("http://localhost:11434"),
	)

	weatherTool := newWeatherTool()

	var chatHistory []anthropic.BetaMessageParam

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		userInput := scanner.Text()

		userMessage := anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(userInput))
		chatHistory = append(chatHistory, userMessage)

		runner := client.Beta.Messages.NewToolRunner(
			[]anthropic.BetaTool{weatherTool},
			anthropic.BetaToolRunnerParams{
				BetaMessageNewParams: anthropic.BetaMessageNewParams{
					MaxTokens: 1024,
					Messages:  chatHistory,
					Model:     ModelQwen,
				},
				MaxIterations: 5,
				DisableThread: true,
			},
		)

		llmMessage, err := runner.RunToCompletion(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, "LLM error:", err)
			chatHistory = chatHistory[:len(chatHistory)-1]
			continue
		}

		for _, block := range llmMessage.Content {
			if textBlock, ok := block.AsAny().(anthropic.BetaTextBlock); ok {
				fmt.Printf("< %s\n", textBlock.Text)
			}
		}

		chatHistory = runner.Messages()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

}

func newWeatherTool() anthropic.BetaTool {
	weatherTool, err := toolrunner.NewBetaToolFromJSONSchema(
		"get_weather",
		"Get current weather for a city",
		func(ctx context.Context, input GetWeatherInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			// wttr.in/Beijing based
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

			return anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: string(b),
				},
			}, nil
		},
	)
	if err != nil {
		panic(err)
	}

	return weatherTool
}

func sendToLLM(client anthropic.Client, messages []anthropic.BetaMessageParam) *anthropic.BetaMessage {
	message, err := client.Beta.Messages.New(context.TODO(), anthropic.BetaMessageNewParams{
		MaxTokens: 1024,
		Messages:  messages,
		Model:     ModelQwen,
	})
	if err != nil {
		panic(err)
	}
	return message
}

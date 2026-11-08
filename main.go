package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
	"github.com/anthropics/anthropic-sdk-go/v2/toolrunner"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// https://github.com/anthropics/anthropic-sdk-go/blob/main/tools.md

// https://github.com/modelcontextprotocol/go-sdk

const (
	// ModelQwen = "qwen2.5-coder:7b"
	ModelQwen = "qwen3:8b"
)

func main() {
	fmt.Println("poorman's claude")
	client := anthropic.NewClient(
		option.WithAPIKey("ollama"),
		option.WithBaseURL("http://localhost:11434"),
	)

	tools, cleanup := discoverTools()
	defer cleanup()

	var chatHistory []anthropic.BetaMessageParam

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		userInput := scanner.Text()

		userMessage := anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(userInput))
		chatHistory = append(chatHistory, userMessage)

		runner := client.Beta.Messages.NewToolRunner(
			tools,
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

func discoverTools() ([]anthropic.BetaTool, func()) {
	ctx := context.Background()

	mcpClient := mcp.NewClient(
		&mcp.Implementation{Name: "poorman", Version: "v0.0.0"},
		nil,
	)

	cmd := exec.Command("go", "run", "./mcpserver")
	cmd.Stderr = os.Stderr

	session, err := mcpClient.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		session.Close()
	}

	var tools []anthropic.BetaTool

	for mcptool, err := range session.Tools(ctx, nil) {
		if err != nil {
			panic(err)
		}

		mcptoolSchema, err := json.Marshal(mcptool.InputSchema)
		if err != nil {
			panic(err)
		}

		// toolrunner.NewBetaTool()
		t, err := toolrunner.NewBetaToolFromBytes(
			mcptool.Name,
			mcptool.Description,
			mcptoolSchema,
			func(ctx context.Context, input json.RawMessage) (
				anthropic.BetaToolResultBlockParamContentUnion, error,
			) {
				result, err := session.CallTool(ctx, &mcp.CallToolParams{
					Name:      mcptool.Name,
					Arguments: input,
				})
				if err != nil {
					return anthropic.BetaToolResultBlockParamContentUnion{}, err
				}

				var parts []string
				for _, content := range result.Content {
					if text, ok := content.(*mcp.TextContent); ok {
						parts = append(parts, text.Text)
					}
				}
				text := strings.Join(parts, "\n")

				if result.IsError {
					return anthropic.BetaToolResultBlockParamContentUnion{}, fmt.Errorf("MCP tool failed: %s", text)
				}

				return anthropic.BetaToolResultBlockParamContentUnion{
					OfText: &anthropic.BetaTextBlockParam{Text: text},
				}, nil

			},
		)

		if err != nil {
			panic(err)
		}

		tools = append(tools, t)
	}

	return tools, cleanup
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

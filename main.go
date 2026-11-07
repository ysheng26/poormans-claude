package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
)

const (
	ModelQwen = "qwen2.5-coder:7b"
)

func main() {
	fmt.Println("poorman's claude")
	client := anthropic.NewClient(
		option.WithAPIKey("ollama"),
		option.WithBaseURL("http://localhost:11434"),
	)

	var chatHistory []anthropic.MessageParam

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		userInput := scanner.Text()

		userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
		chatHistory = append(chatHistory, userMessage)

		llmMessage := sendToLLM(client, chatHistory)
		for _, block := range llmMessage.Content {
			if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
				fmt.Printf("< %s\n", textBlock.Text)
			}
		}

		chatHistory = append(chatHistory, llmMessage.ToParam())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func sendToLLM(client anthropic.Client, messages []anthropic.MessageParam) *anthropic.Message {
	message, err := client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		MaxTokens: 1024,
		Messages:  messages,
		Model:     ModelQwen,
	})
	if err != nil {
		panic(err)
	}
	return message
}

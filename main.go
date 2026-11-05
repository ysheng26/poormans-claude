package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

/*
Messages =[]
for {
  Get user input
  Append to Messages
  Send Messages to LLM
  Append response to Messages
  Display response
}
*/

// https://developers.openai.com/api/docs/libraries?language=golang

func main() {
	client := openai.NewClient(
		option.WithBaseURL("http://localhost:11434/v1"),
		option.WithAPIKey("ollama"),
	)

	fmt.Println("poorman's codex")

	var messages []string

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		userInput := scanner.Text()

		// fmt.Printf("> %s\n", userInput)
		messages = append(messages, userInput)

		llmResponse := sendToLLM(client, userInput)
		fmt.Printf("< %s\n", llmResponse)
		messages = append(messages, llmResponse)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func sendToLLM(client openai.Client, message string) string {
	response, err := client.Responses.New(context.TODO(),
		responses.ResponseNewParams{
			Model: "qwen2.5-coder:7b",
			Input: responses.ResponseNewParamsInputUnion{
				OfString: openai.String(message),
			},
		})
	if err != nil {
		panic(err)
	}

	return response.OutputText()
}

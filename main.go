package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	Model         = "qwen2.5-coder:7b"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	fmt.Println("poorman's codex")

	var chatHistory []Message

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		userInput := scanner.Text()

		chatHistory = append(chatHistory, Message{Role: RoleUser, Content: userInput})

		llmMessage := sendToLLM(chatHistory)
		fmt.Printf("< %s\n", llmMessage.Content)

		chatHistory = append(chatHistory, llmMessage)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func sendToLLM(messages []Message) Message {

	type payload struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Stream   bool      `json:"stream"`
	}

	p := payload{
		Model:    Model,
		Messages: messages,
		Stream:   false,
	}

	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	type response struct {
		Message Message `json:"message"`
	}
	var r response
	err = json.Unmarshal(body, &r)
	if err != nil {
		panic(err)
	}
	return r.Message
}

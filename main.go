package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

const (
	Model         = "qwen2.5-coder:7b"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const getTimeToolPrompt = `
I am providing you a tool which could tell you time.
When you need to call the tool you say it to me like the following [YOYO]get_time[/YOYO]. and I would return the result to you with [SUPSUP]time[/SUPSUP]

Get the time from between [SUPSUP][/SUPSUP]

For example in 
[SUPSUP]2026-11-07 11:00:56.810491 -0500 EST m=+2.747323334[/SUPSUP]
the time is 2026-11-07 11:00:56.810491 -0500 EST m=+2.747323334

DO NOT return the SUPSUP tags to user

`

const getTimeToolRegexRaw = `\[YOYO\](.*)\[\/YOYO\]`

/*
tool prompt
parse tool response
execute tool
*/

func main() {
	fmt.Println("poorman's codex")

	var chatHistory []Message
	getTimeToolMessage := Message{Role: RoleTool, Content: getTimeToolPrompt}
	getTimeToolRegex := regexp.MustCompile(getTimeToolRegexRaw)

	chatHistory = append(chatHistory, getTimeToolMessage)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		userInput := scanner.Text()

		chatHistory = append(chatHistory, Message{Role: RoleUser, Content: userInput})

		llmMessage := sendToLLM(chatHistory)

		matches := getTimeToolRegex.FindStringSubmatch(llmMessage.Content)

		for matches != nil {
			if matches[1] == "get_time" {
				timeResult := Message{
					Role:    RoleTool,
					Content: fmt.Sprintf("[SUPSUP]%s[/SUPSUP]", time.Now()),
				}
				chatHistory = append(chatHistory, timeResult)
				llmMessage = sendToLLM(chatHistory)
				matches = getTimeToolRegex.FindStringSubmatch(llmMessage.Content)
			}
		}

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

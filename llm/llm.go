package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/blixt/first-aid/tool"
)

type LLM struct {
	model    string
	messages []Message
	toolbox  *tool.Toolbox

	totalPromptTokens, totalCompletionTokens int

	// SystemPrompt should return the system prompt for the LLM. It's a function
	// to allow the system prompt to dynamically change throughout a single
	// conversation.
	SystemPrompt func() Content
}

func New(model string, tools ...tool.Tool) *LLM {
	var toolbox *tool.Toolbox
	if len(tools) > 0 {
		toolbox = tool.Box(tools...)
	}
	return &LLM{
		model:   model,
		toolbox: toolbox,
	}
}

// Chat sends a text message to the LLM and immediately returns a channel over
// which updates will come in. The LLM will use the tools available and keep
// generating more messages until it's done using tools.
func (l *LLM) Chat(message string) <-chan Update {
	return l.ChatUsingContent(Text(message))
}

// ChatUsingContent sends a message (which can contain images) to the LLM and
// immediately returns a channel over which updates will come in. The LLM will
// use the tools available and keep generating more messages until it's done
// using tools.
func (l *LLM) ChatUsingContent(message Content) <-chan Update {
	l.messages = append(l.messages, Message{
		Role:    "user",
		Content: message,
	})

	// Send off the user's message to the LLM, and keep asking the LLM for more
	// responses for as long as it's making tool calls.
	updateChan := make(chan Update)
	go func() {
		defer close(updateChan)
		for {
			shouldContinue, err := l.step(updateChan)
			if err != nil {
				updateChan <- ErrorUpdate{Error: err}
				return
			}
			if !shouldContinue {
				return
			}
		}
	}()

	return updateChan
}

func (l *LLM) AddTool(t tool.Tool) {
	if l.toolbox == nil {
		l.toolbox = tool.Box(t)
	} else {
		l.toolbox.Add(t)
	}
}

func (l *LLM) Usage() (promptTokens, completionTokens int) {
	return l.totalPromptTokens, l.totalCompletionTokens
}

func (l *LLM) step(updateChan chan<- Update) (bool, error) {
	var messages []Message
	if l.SystemPrompt != nil {
		messages = make([]Message, 0, len(l.messages)+1)
		messages = append(messages, Message{
			Role:    "system",
			Content: l.SystemPrompt(),
		})
		messages = append(messages, l.messages...)
	} else {
		messages = l.messages
	}

	payload := map[string]any{
		"model":          l.model,
		"messages":       messages,
		"stream":         true,
		"stream_options": map[string]any{"include_usage": true},
	}

	if l.toolbox != nil {
		payload["tools"] = l.toolbox.Schema()
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("error encoding JSON: %w", err)
	}

	// TODO: Support more than OpenAI.
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		decoder := json.NewDecoder(resp.Body)
		var response map[string]any
		if err := decoder.Decode(&response); err != nil {
			return false, fmt.Errorf("received error status %s and failed to parse the body: %w", resp.Status, err)
		}
		return false, fmt.Errorf("%s: %v", resp.Status, response)
	}

	// This will hold results from tool calls, to be sent back to the LLM.
	var toolMessages []Message

	stream := NewMessageStream(resp.Body)

	// Write the entire message history to the file debug.yaml. The function is
	// deferred so that we get data even if a panic occurs.
	defer func() {
		var toolsSchema []tool.Schema
		if l.toolbox != nil {
			toolsSchema = l.toolbox.Schema()
		}
		debugData := map[string]any{
			// Prefixed with numbers so the keys remain in this order.
			"1_sentMessages":    messages,
			"2_receivedMessage": stream.Message(),
			"3_toolResults":     toolMessages,
			"4_availableTools":  toolsSchema,
			"5_usage":           stream.Usage(),
		}
		if debugYAML, err := yaml.Marshal(debugData); err == nil {
			os.WriteFile("debug.yaml", debugYAML, 0644)
		}
	}()

	// A future version of Go will allow us to use range on the function below:
	// for status := range stream.Iter() {
	stream.Iter()(func(status StreamStatus) bool {
		switch status {
		case StreamStatusText:
			updateChan <- TextUpdate{Text: stream.Text()}
		case StreamStatusToolCallBegin:
			tool := l.toolbox.Get(stream.ToolCall().Function.Name)
			if tool == nil {
				// TODO: This should be handled more gracefully.
				panic(fmt.Sprintf("tool %q not found", stream.ToolCall().Function.Name))
			}
			updateChan <- ToolStartUpdate{Tool: tool}
		case StreamStatusToolCallReady:
			messages := l.runToolCall(l.toolbox, stream.ToolCall(), updateChan)
			toolMessages = append(toolMessages, messages...)
		}
		return true
	})

	if err := stream.Err(); err != nil {
		io.Copy(io.Discard, resp.Body)
		return false, fmt.Errorf("error streaming: %w", err)
	}

	// Add the fully assembled message plus tool call results to the message history.
	l.messages = append(l.messages, stream.Message())
	l.messages = append(l.messages, toolMessages...)

	if usage := stream.Usage(); usage != nil {
		l.totalPromptTokens += usage.PromptTokens
		l.totalCompletionTokens += usage.CompletionTokens
	}

	// Return true if there were tool calls, since the LLM should look at the results.
	return len(toolMessages) > 0, nil
}

func (l *LLM) runToolCall(toolbox *tool.Toolbox, toolCall ToolCall, updateChan chan<- Update) []Message {
	// As a sanity check, make sure we don't try to run the same tool call twice.
	for _, message := range l.messages {
		if message.ToolCallID == toolCall.ID {
			fmt.Printf("\ntool call %q (%s) has already been run\n", toolCall.ID, toolCall.Function.Name)
		}
	}

	t := toolbox.Get(toolCall.Function.Name)
	runner := tool.NewRunner(toolbox, func(status string) {
		updateChan <- ToolStatusUpdate{Status: status, Tool: t}
	})

	result := toolbox.Run(runner, toolCall.Function.Name, json.RawMessage(toolCall.Function.Arguments))
	updateChan <- ToolDoneUpdate{Result: result, Tool: t}

	// Explicitly stating that the result is empty reduces hallucinations.
	content := strings.TrimSpace(result.String())
	if content == "" {
		content = "(The tool did not return anything.)"
	}

	messages := []Message{
		{
			Role:       "tool",
			Content:    Text(content),
			ToolCallID: toolCall.ID,
		},
	}

	if images := result.Images(); len(images) > 0 {
		// "tool" messages can't actually contain image content. So we need to
		// fake an assistant message instead.
		message := Message{
			Role: "user",
			// TODO: Support more than one image name.
			Content: Textf("Here is %s. This is an automated message, not actually from the user.", images[0].Name),
		}
		for _, image := range images {
			message.Content.AddImage(image.URL)
		}
		messages = append(messages, message)
	}

	return messages
}

package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/blixt/first-aid/llm"
	"github.com/blixt/first-aid/tool"
)

type Model struct {
	model string
}

func New(model string) *Model {
	return &Model{model}
}

func (p *Model) Generate(systemPrompt llm.Content, messages []llm.Message, tools *tool.Toolbox) llm.ProviderStream {
	var apiMessages []message
	if systemPrompt != nil {
		apiMessages = make([]message, 0, len(messages)+1)
		apiMessages = append(apiMessages, message{
			Role:    "system",
			Content: systemPrompt,
		})
	} else {
		apiMessages = make([]message, 0, len(messages))
	}
	for _, msg := range messages {
		apiMessages = append(apiMessages, messageFromLLM(msg))
	}

	payload := map[string]any{
		"model":          p.model,
		"messages":       apiMessages,
		"stream":         true,
		"stream_options": map[string]any{"include_usage": true},
	}

	if tools != nil {
		payload["tools"] = tools.Schema()
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &Stream{err: fmt.Errorf("error encoding JSON: %w", err)}
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return &Stream{err: fmt.Errorf("error creating request: %w", err)}
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &Stream{err: fmt.Errorf("error making request: %w", err)}
	}

	return &Stream{model: p.model, stream: resp.Body, toolbox: tools}
}

type Stream struct {
	model    string
	stream   io.Reader
	err      error
	message  llm.Message
	lastText string
	usage    *usage
	toolbox  *tool.Toolbox
}

func (s *Stream) Err() error {
	return s.err
}

func (s *Stream) Message() llm.Message {
	return s.message
}

func (s *Stream) Text() string {
	return s.lastText
}

func (s *Stream) ToolCall() llm.ToolCall {
	if len(s.message.ToolCalls) == 0 {
		return llm.ToolCall{}
	}
	return s.message.ToolCalls[len(s.message.ToolCalls)-1]
}

func (s *Stream) CostUSD() float64 {
	switch s.model {
	case "gpt-4o":
		const inputCost = 5   // per million tokens
		const outputCost = 15 // per million tokens
		inputTokens, outputTokens := s.Usage()
		return float64(inputTokens)*inputCost/1e6 + float64(outputTokens)*outputCost/1e6
	default:
		panic(fmt.Sprintf("unknown model: %q", s.model))
	}
}

func (s *Stream) Usage() (inputTokens, outputTokens int) {
	if s.usage == nil {
		return 0, 0
	}
	return s.usage.PromptTokens, s.usage.CompletionTokens
}

func (s *Stream) Iter() func(yield func(llm.StreamStatus) bool) {
	scanner := bufio.NewScanner(s.stream)
	return func(yield func(llm.StreamStatus) bool) {
		defer io.Copy(io.Discard, s.stream)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			line = strings.TrimPrefix(line, "data: ")
			if line == "[DONE]" {
				continue
			}
			var chunk chatCompletionChunk
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				s.err = fmt.Errorf("error unmarshalling chunk: %w", err)
				break
			}
			if chunk.Usage != nil {
				s.usage = chunk.Usage
			}
			if len(chunk.Choices) < 1 {
				continue
			}
			delta := chunk.Choices[0].Delta
			if delta.Role != "" {
				s.message.Role = delta.Role
			}
			s.lastText = delta.Content
			if s.lastText != "" {
				s.message.Content.Append(s.lastText)
				if !yield(llm.StreamStatusText) {
					return
				}
			}
			if len(delta.ToolCalls) > 1 {
				panic("received more than one tool call in a single chunk")
			}
			if len(delta.ToolCalls) == 0 {
				continue
			}
			toolDelta := delta.ToolCalls[0]
			if toolDelta.Index < len(s.message.ToolCalls) {
				if toolDelta.Index != len(s.message.ToolCalls)-1 {
					panic("tool call index mismatch")
				}
				s.message.ToolCalls[toolDelta.Index].Arguments = append(s.message.ToolCalls[toolDelta.Index].Arguments, toolDelta.Function.Arguments...)
				if !yield(llm.StreamStatusToolCallData) {
					return
				}
			} else {
				if toolDelta.Index > 0 {
					if !yield(llm.StreamStatusToolCallReady) {
						return
					}
				}
				s.message.ToolCalls = append(s.message.ToolCalls, toolDelta.ToLLM())
				if !yield(llm.StreamStatusToolCallBegin) {
					return
				}
			}
		}
		if len(s.message.ToolCalls) > 0 {
			if !yield(llm.StreamStatusToolCallReady) {
				return
			}
		}
	}
}

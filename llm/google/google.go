package google

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/blixt/first-aid/llm"
	"github.com/blixt/first-aid/tool"
)

type Model struct {
	accessToken     string
	model           string
	endpoint        string
	maxOutputTokens int
	temperature     float64
	topP            float64
}

func New(model string) *Model {
	return &Model{
		model:           model,
		maxOutputTokens: 8192,
		temperature:     1,
		topP:            0.95,
	}
}

func (m *Model) WithGeminiAPI(apiKey string) *Model {
	m.accessToken = ""
	m.endpoint = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", m.model, apiKey)
	return m
}

func (m *Model) WithVertexAI(accessToken, projectID, region string) *Model {
	// TODO: This API has a cost per 1,000 UTF-8 code points (excluding whitespace).
	// https://cloud.google.com/vertex-ai/generative-ai/pricing
	m.accessToken = accessToken
	m.endpoint = fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:streamGenerateContent?alt=sse", region, projectID, region, m.model)
	return m
}

func (m *Model) WithMaxOutputTokens(maxOutputTokens int) *Model {
	m.maxOutputTokens = maxOutputTokens
	return m
}

func (m *Model) WithTemperature(temperature float64) *Model {
	m.temperature = temperature
	return m
}

func (m *Model) WithTopP(topP float64) *Model {
	m.topP = topP
	return m
}

func (m *Model) Company() string {
	return "Google"
}

func (m *Model) Generate(systemPrompt llm.Content, messages []llm.Message, tools *tool.Toolbox) llm.ProviderStream {
	if m.endpoint == "" {
		return &Stream{err: fmt.Errorf("must call either WithVertexAI(…) or WithGenerativeLanguageAPI(…) first")}
	}

	var apiMessages []message
	for _, msg := range messages {
		apiMessages = append(apiMessages, messageFromLLM(msg))
	}

	payload := map[string]any{
		"contents": apiMessages,
		"generationConfig": map[string]any{
			"maxOutputTokens": m.maxOutputTokens,
			"temperature":     m.temperature,
			"topP":            m.topP,
		},
		"safetySettings": []map[string]any{
			{
				"category":  "HARM_CATEGORY_HATE_SPEECH",
				"threshold": "BLOCK_ONLY_HIGH",
			},
			{
				"category":  "HARM_CATEGORY_DANGEROUS_CONTENT",
				"threshold": "BLOCK_ONLY_HIGH",
			},
			{
				"category":  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				"threshold": "BLOCK_ONLY_HIGH",
			},
			{
				"category":  "HARM_CATEGORY_HARASSMENT",
				"threshold": "BLOCK_ONLY_HIGH",
			},
		},
	}

	if systemPrompt != nil {
		payload["systemInstruction"] = map[string]any{
			"parts": partsFromLLM(systemPrompt),
		}
	}

	if tools != nil {
		allTools := tools.All()
		declarations := make([]tool.FunctionSchema, len(allTools))
		for i, tool := range allTools {
			declarations[i] = *tool.Schema()
		}
		payload["tools"] = map[string]any{
			"functionDeclarations": declarations,
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &Stream{err: fmt.Errorf("error encoding JSON: %w", err)}
	}

	req, err := http.NewRequest("POST", m.endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return &Stream{err: fmt.Errorf("error creating request: %w", err)}
	}
	if m.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.accessToken))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &Stream{err: fmt.Errorf("error making request: %w", err)}
	}
	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return &Stream{err: fmt.Errorf("error decoding %s response: %w", resp.Status, err)}
		}
		return &Stream{err: fmt.Errorf("%s: %s", resp.Status, errResp.Error.Message)}
	}
	return &Stream{model: m.model, stream: resp.Body}
}

type Stream struct {
	model    string
	stream   io.Reader
	err      error
	message  llm.Message
	lastText string
	usage    *usageMetadata
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
	// TODO
	return 0.0
}

func (s *Stream) Usage() (inputTokens, outputTokens int) {
	if s.usage == nil {
		return 0, 0
	}
	return s.usage.PromptTokenCount, s.usage.CandidatesTokenCount
}

func (s *Stream) Iter() func(yield func(llm.StreamStatus) bool) {
	scanner := bufio.NewScanner(s.stream)
	return func(yield func(llm.StreamStatus) bool) {
		defer io.Copy(io.Discard, s.stream)
		for scanner.Scan() {
			line, ok := strings.CutPrefix(scanner.Text(), "data: ")
			if !ok {
				continue
			}
			var chunk streamingResponse
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				s.err = fmt.Errorf("error unmarshalling chunk: %w", err)
				break
			}
			if chunk.UsageMetadata != nil {
				s.usage = chunk.UsageMetadata
			}
			if len(chunk.Candidates) < 1 {
				continue
			}
			delta := chunk.Candidates[0].Content
			if delta.Role != "" {
				s.message.Role = delta.Role
			}
			for _, p := range delta.Parts {
				if p.Text != nil {
					s.lastText = *p.Text
					if s.lastText != "" {
						s.message.Content.Append(s.lastText)
						if !yield(llm.StreamStatusText) {
							return
						}
					}
				}
				if p.FunctionCall != nil {
					// Note: Gemini's streaming API doesn't have partial tool calls.
					s.message.ToolCalls = append(s.message.ToolCalls, llm.ToolCall{
						Name:      p.FunctionCall.Name,
						Arguments: p.FunctionCall.Args,
					})
					if !yield(llm.StreamStatusToolCallBegin) {
						return
					}
					if !yield(llm.StreamStatusToolCallData) {
						return
					}
					if !yield(llm.StreamStatusToolCallReady) {
						return
					}
				}
			}
		}
	}
}

package llm

import (
	"github.com/blixt/first-aid/tool"
)

type ProviderStream interface {
	Err() error
	Iter() func(yield func(StreamStatus) bool)
	Message() Message
	Text() string
	ToolCall() ToolCall
	CostUSD() float64
	Usage() (inputTokens, outputTokens int)
}

type Provider interface {
	Company() string
	Generate(systemPrompt Content, messages []Message, tools *tool.Toolbox) ProviderStream
}

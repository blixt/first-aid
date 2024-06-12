package openai

import (
	"encoding/json"

	"github.com/blixt/first-aid/llm"
)

type message struct {
	// Role can be "system", "user", "assistant", or "tool".
	Role string `json:"role"`
	// Name can be used to identify different identities within the same role.
	Name string `json:"name,omitempty"`
	// Content is the message content.
	Content llm.Content `json:"content"`
	// ToolCalls is the list of tool calls that this message is part of.
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
	// ToolCallID is the ID of the tool call that this message is part of.
	ToolCallID string `json:"tool_call_id,omitempty"`
}

func messageFromLLM(m llm.Message) message {
	toolCalls := make([]toolCall, len(m.ToolCalls))
	for i, tc := range m.ToolCalls {
		toolCalls[i] = toolCall{
			ID:       tc.ID,
			Type:     "function",
			Function: toolCallFunction{Name: tc.Name, Arguments: string(tc.Arguments)},
		}
	}
	return message{
		Role:       m.Role,
		Name:       m.Name,
		Content:    m.Content,
		ToolCalls:  toolCalls,
		ToolCallID: m.ToolCallID,
	}
}

type toolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type toolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function toolCallFunction `json:"function"`
}

func (t toolCall) ToLLM() llm.ToolCall {
	return llm.ToolCall{
		ID:        t.ID,
		Name:      t.Function.Name,
		Arguments: json.RawMessage(t.Function.Arguments),
	}
}

type toolCallDelta struct {
	toolCall
	Index int `json:"index"`
}

type chatCompletionDelta struct {
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	ToolCalls []toolCallDelta `json:"tool_calls"`
}

type chatCompletionChoice struct {
	Index        int                 `json:"index"`
	Delta        chatCompletionDelta `json:"delta"`
	FinishReason string              `json:"finish_reason"`
}

type chatCompletionChunk struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	SystemFingerprint string                 `json:"system_fingerprint"`
	Choices           []chatCompletionChoice `json:"choices"`
	Usage             *usage                 `json:"usage"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

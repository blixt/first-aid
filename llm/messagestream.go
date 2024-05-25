package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type StreamStatus int

const (
	// StreamStatusNone means either the stream hasn't started, or it has finished.
	StreamStatusNone StreamStatus = iota
	// StreamStatusText means the stream produced more text content.
	StreamStatusText
	// StreamStatusToolCallBegin means the stream started a tool call. The name of the function is available, but not the arguments.
	StreamStatusToolCallBegin
	// StreamStatusToolCallData means the stream is streaming the arguments for a tool call.
	StreamStatusToolCallData
	// StreamStatusToolCallReady means the stream finished streaming the arguments for a tool call.
	StreamStatusToolCallReady
)

type MessageStream struct {
	scanner  *bufio.Scanner
	err      error
	message  Message
	lastText string
	usage    *Usage
}

func NewMessageStream(r io.Reader) *MessageStream {
	return &MessageStream{scanner: bufio.NewScanner(r)}
}

// Err returns the error that occurred while reading the stream, if any.
func (s *MessageStream) Err() error {
	return s.err
}

// Message returns the message reconstructed from the stream.
func (s *MessageStream) Message() Message {
	return s.message
}

// Text returns the last text chunk received in the stream.
func (s *MessageStream) Text() string {
	return s.lastText
}

// ToolCall returns the currently streaming (or last streamed) tool call.
func (s *MessageStream) ToolCall() ToolCall {
	if len(s.message.ToolCalls) == 0 {
		return ToolCall{}
	}
	return s.message.ToolCalls[len(s.message.ToolCalls)-1]
}

func (s *MessageStream) Usage() *Usage {
	return s.usage
}

// Iter returns a function that can be used to iterate over the stream. The
// yield function is called for each message in the stream. If it returns false,
// the iteration is stopped.
func (s *MessageStream) Iter() func(yield func(StreamStatus) bool) {
	return func(yield func(StreamStatus) bool) {
		for s.scanner.Scan() {
			line := s.scanner.Text()
			// Ignore lines that aren't data messages.
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			line = strings.TrimPrefix(line, "data: ")
			if line == "[DONE]" {
				// TODO: For now we ignore this event but we should confirm it was actually the final event.
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
				if !yield(StreamStatusText) {
					return
				}
			}
			if len(delta.ToolCalls) > 1 {
				// It's unlikely that any API would not just return all tool
				// calls sequentially, but this sanity check saves us from
				// debugging headaches if it were to happen.
				panic("received more than one tool call in a single chunk")
			}
			if len(delta.ToolCalls) == 0 {
				// We don't need to do anything related to tool calls.
				continue
			}
			toolDelta := delta.ToolCalls[0]
			if toolDelta.Index < len(s.message.ToolCalls) {
				// We don't allow updating tool calls we already marked as ready.
				if toolDelta.Index != len(s.message.ToolCalls)-1 {
					panic("tool call index mismatch")
				}
				// We are updating an existing tool call; add to the arguments data.
				s.message.ToolCalls[toolDelta.Index].Function.Arguments += toolDelta.Function.Arguments
				if !yield(StreamStatusToolCallData) {
					return
				}
			} else {
				if toolDelta.Index > 0 {
					// There was a previous tool that will now be ready.
					if !yield(StreamStatusToolCallReady) {
						return
					}
				}
				// Add the new tool call.
				s.message.ToolCalls = append(s.message.ToolCalls, toolDelta.ToolCall)
				if !yield(StreamStatusToolCallBegin) {
					return
				}
			}
		}
		if len(s.message.ToolCalls) > 0 {
			// The final tool call is confirmed ready once we've finished streaming.
			if !yield(StreamStatusToolCallReady) {
				return
			}
		}
	}
}

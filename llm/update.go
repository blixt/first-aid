package llm

import (
	"github.com/blixt/first-aid/tool"
)

type UpdateType string

const (
	UpdateTypeToolStart  UpdateType = "tool_start"
	UpdateTypeToolStatus UpdateType = "tool_status"
	UpdateTypeToolDone   UpdateType = "tool_done"
	UpdateTypeError      UpdateType = "error"
	UpdateTypeText       UpdateType = "text"
)

type Update interface {
	Type() UpdateType
}

type ToolStartUpdate struct {
	Tool tool.Tool
}

func (u ToolStartUpdate) Type() UpdateType {
	return UpdateTypeToolStart
}

type ToolStatusUpdate struct {
	Status string
	Tool   tool.Tool
}

func (u ToolStatusUpdate) Type() UpdateType {
	return UpdateTypeToolStatus
}

type ToolDoneUpdate struct {
	Result tool.Result
	Tool   tool.Tool
}

func (u ToolDoneUpdate) Type() UpdateType {
	return UpdateTypeToolDone
}

type ErrorUpdate struct {
	Error error
}

func (u ErrorUpdate) Type() UpdateType {
	return UpdateTypeError
}

type TextUpdate struct {
	Text string
}

func (u TextUpdate) Type() UpdateType {
	return UpdateTypeText
}

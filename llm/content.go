package llm

import (
	"encoding/json"
	"fmt"
)

type ContentItemType string

const (
	TypeText       ContentItemType = "text"
	TypeImageURL   ContentItemType = "imageURL"
	TypeToolResult ContentItemType = "toolResult"
)

type ContentItem interface {
	Type() ContentItemType
}

type TextContent struct {
	Text string
}

func (tc *TextContent) Type() ContentItemType {
	return TypeText
}

type ImageURLContent struct {
	URL string
}

func (iuc *ImageURLContent) Type() ContentItemType {
	return TypeImageURL
}

type ToolResultContent struct {
	CallID string
	Data   json.RawMessage
}

func (trc *ToolResultContent) Type() ContentItemType {
	return TypeToolResult
}

type Content []ContentItem

func ToolResult(callID string, value any) (Content, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return ToolResultJSON(callID, json.RawMessage(data)), nil
}

func ToolResultJSON(callID string, data json.RawMessage) Content {
	return Content{
		&ToolResultContent{CallID: callID, Data: data},
	}
}

// Text returns a new content item with the given text.
func Text(text string) Content {
	return Content{
		&TextContent{Text: text},
	}
}

// Textf returns a new content item with the provided formatted text.
func Textf(format string, args ...any) Content {
	return Text(fmt.Sprintf(format, args...))
}

// TextAndImage returns a new content item with the given text and image URL.
func TextAndImage(text, imageURL string) Content {
	return Content{
		&TextContent{Text: text},
		&ImageURLContent{URL: imageURL},
	}
}

// AddImage adds an image URL to the content.
func (c *Content) AddImage(imageURL string) {
	*c = append(*c, &ImageURLContent{URL: imageURL})
}

// Append adds the text to the last content item if it's a text item, otherwise
// it adds a new text item to the end of the list.
func (c *Content) Append(text string) {
	if l := len(*c); l > 0 {
		if tc, ok := (*c)[l-1].(*TextContent); ok {
			tc.Text += text
			return
		}
	}
	*c = append(*c, &TextContent{Text: text})
}

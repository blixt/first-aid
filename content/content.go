package content

import (
	"encoding/json"
	"fmt"
)

type Type string

const (
	TypeText     Type = "text"
	TypeImageURL Type = "imageURL"
	TypeJSON     Type = "json"
)

type Item interface {
	Type() Type
}

type Text struct {
	Text string
}

func (t *Text) Type() Type {
	return TypeText
}

type ImageURL struct {
	URL string
}

func (iu *ImageURL) Type() Type {
	return TypeImageURL
}

type JSON struct {
	Data json.RawMessage
}

func (j *JSON) Type() Type {
	return TypeJSON
}

type Content []Item

func FromJSON(value any) (Content, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return FromRawJSON(data), nil
}

func FromRawJSON(data json.RawMessage) Content {
	return Content{
		&JSON{Data: data},
	}
}

// FromText returns a new content item with the given text.
func FromText(text string) Content {
	return Content{
		&Text{Text: text},
	}
}

// Textf returns a new content item with the provided formatted text.
func Textf(format string, args ...any) Content {
	return FromText(fmt.Sprintf(format, args...))
}

// TextAndImage returns a new content item with the given text and image URL.
func FromTextAndImage(text, imageURL string) Content {
	return Content{
		&Text{Text: text},
		&ImageURL{URL: imageURL},
	}
}

// AddImage adds an image URL to the content.
func (c *Content) AddImage(imageURL string) {
	*c = append(*c, &ImageURL{URL: imageURL})
}

// Append adds the text to the last content item if it's a text item, otherwise
// it adds a new text item to the end of the list.
func (c *Content) Append(text string) {
	if l := len(*c); l > 0 {
		if tc, ok := (*c)[l-1].(*Text); ok {
			tc.Text += text
			return
		}
	}
	*c = append(*c, &Text{Text: text})
}

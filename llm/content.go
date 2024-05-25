package llm

import (
	"encoding/json"
	"fmt"
)

type ImageURL struct {
	URL string `json:"url"`
}

type ContentItem struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type Content []ContentItem

// Text returns a new content item with the given text.
func Text(text string) Content {
	return Content{
		{
			Type: "text",
			Text: text,
		},
	}
}

// Textf returns a new content item with the provided formatted text.
func Textf(format string, args ...any) Content {
	return Text(fmt.Sprintf(format, args...))
}

// TextAndImage returns a new content item with the given text and image URL.
func TextAndImage(text, imageURL string) Content {
	return Content{
		{
			Type: "text",
			Text: text,
		},
		{
			Type:     "image_url",
			ImageURL: &ImageURL{URL: imageURL},
		},
	}
}

// AddImage adds an image URL to the content.
func (c *Content) AddImage(imageURL string) {
	*c = append(*c, ContentItem{
		Type:     "image_url",
		ImageURL: &ImageURL{URL: imageURL},
	})
}

// Append adds the text to the last content item if it's a text item, otherwise
// it adds a new text item to the end of the list.
func (c *Content) Append(text string) {
	if l := len(*c); l > 0 && (*c)[l-1].Type == "text" {
		(*c)[l-1].Text += text
	} else {
		*c = append(*c, ContentItem{
			Type: "text",
			Text: text,
		})
	}
}

func (c Content) MarshalJSON() ([]byte, error) {
	// Marshal into a simple string when the only content is one text item.
	if len(c) == 1 && c[0].Type == "text" {
		return json.Marshal(c[0].Text)
	}
	// Otherwise, directly marshal the content slice.
	return json.Marshal([]ContentItem(c))
}

func (c *Content) UnmarshalJSON(data []byte) error {
	// Try to unmarshal data as a JSON string first.
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*c = Content{
			{
				Type: "text",
				Text: text,
			},
		}
		return nil
	}
	// If that failed, unmarshal it as an array of content items.
	var contentItems []ContentItem
	if err := json.Unmarshal(data, &contentItems); err != nil {
		return err
	}
	*c = Content(contentItems)
	return nil
}

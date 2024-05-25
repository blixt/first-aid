package tool

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type Result interface {
	// Label returns a short single line description of the entire tool run.
	Label() string
	// String returns a string representation of the result.
	String() string
	// Error returns the error that occurred during the tool run, if any.
	Error() error
	// Images returns a slice of base64 encoded images.
	Images() []Image
	// NextToolbox returns a toolbox that will be used to handle this result.
	NextToolbox() *Toolbox
}

type result struct {
	label       string
	content     string
	err         error
	images      []Image
	nextToolbox *Toolbox
}

func (r *result) Label() string {
	return r.label
}

func (r *result) String() string {
	return r.content
}

func (r *result) Error() error {
	return r.err
}

func (r *result) Images() []Image {
	return r.images
}

func (r *result) NextToolbox() *Toolbox {
	return r.nextToolbox
}

func Error(label string, err error) Result {
	return &result{label, fmt.Sprintf("ERROR: %s", err), err, nil, nil}
}

func Success(label, content string) Result {
	return &result{label, content, nil, nil, nil}
}

type ResultBuilder struct {
	images      []Image
	nextToolbox *Toolbox
}

type Image struct {
	Name, URL string
}

func (b *ResultBuilder) AddImage(path string) error {
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		return fmt.Errorf("unknown file extension: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Turn the image data into a base64 string.
	var encodedImage strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &encodedImage)
	defer encoder.Close()
	if _, err := io.Copy(encoder, file); err != nil {
		return err
	}

	image := Image{
		Name: filepath.Base(path),
		URL:  fmt.Sprintf("data:%s;base64,%s", mimeType, encodedImage.String()),
	}
	b.images = append(b.images, image)
	return nil
}

func (b *ResultBuilder) Toolbox(toolbox *Toolbox) {
	b.nextToolbox = toolbox
}

func (b *ResultBuilder) Error(label string, err error) Result {
	return &result{label, fmt.Sprintf("ERROR: %s", err), err, b.images, b.nextToolbox}
}

func (b *ResultBuilder) Success(label, content string) Result {
	return &result{label, content, nil, b.images, b.nextToolbox}
}

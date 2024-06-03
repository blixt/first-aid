package tool

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
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

func (b *ResultBuilder) AddImage(path string, highQuality bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Check image dimensions and resize if necessary.
	var maxDim int
	if highQuality {
		maxDim = 2048
	} else {
		maxDim = 512
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width > maxDim || height > maxDim {
		var newWidth, newHeight int
		if width > height {
			newWidth = maxDim
			newHeight = (height * maxDim) / width
		} else {
			newHeight = maxDim
			newWidth = (width * maxDim) / height
		}

		resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, bounds, draw.Over, nil)
		img = resizedImg
	}

	// Turn the image data into a base64 string.
	var encodedImage strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &encodedImage)
	defer encoder.Close()

	var mimeType string
	switch format {
	case "jpeg":
		err = jpeg.Encode(encoder, img, nil)
		mimeType = "image/jpeg"
	case "png":
		err = png.Encode(encoder, img)
		mimeType = "image/png"
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
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

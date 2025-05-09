package firstaid

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flitsinc/go-llms/content"
	"github.com/flitsinc/go-llms/tools"
)

type LookAtImageParams struct {
	Path        string `json:"path"`
	HighQuality bool   `json:"high_quality,omitempty" description:"Use true if you want to see the image in higher resolution."`
}

var LookAtImage = tools.Func(
	"Look at image",
	"Displays an image from the specified path. Use this to view an image file.",
	"look_at_image",
	func(r tools.Runner, p LookAtImageParams) tools.Result {
		p.Path = expandPath(p.Path)
		label := fmt.Sprintf("Look at image `%s`", filepath.Base(p.Path))
		if _, err := os.Stat(p.Path); os.IsNotExist(err) {
			return tools.ErrorWithLabel(label, fmt.Errorf("file does not exist: %s", p.Path))
		}
		imageName, dataURI, err := content.ImageToDataURI(p.Path, p.HighQuality)
		if err != nil {
			return tools.ErrorWithLabel(label, fmt.Errorf("failed to process image %s: %w", imageName, err))
		}
		resultContent := content.Content{&content.ImageURL{URL: dataURI}}
		return tools.SuccessWithContent(label, resultContent)
	},
)

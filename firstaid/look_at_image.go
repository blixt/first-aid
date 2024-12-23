package firstaid

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blixt/go-llms/tools"
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
			return tools.Error(label, fmt.Errorf("file does not exist: %s", p.Path))
		}
		var rb tools.ResultBuilder
		if err := rb.AddImage(p.Path, p.HighQuality); err != nil {
			return tools.Error(label, err)
		}
		content := map[string]string{
			"message": fmt.Sprintf("You will receive %s from the user as an automated message.", filepath.Base(p.Path)),
		}
		return rb.Success(label, content)
	},
)

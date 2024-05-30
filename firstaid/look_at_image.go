package firstaid

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blixt/first-aid/tool"
)

type LookAtImageParams struct {
	Path string `json:"path"`
}

var LookAtImage = tool.Func(
	"Look at image",
	"Displays an image from the specified path. Use this to view an image file.",
	"look_at_image",
	func(r tool.Runner, p LookAtImageParams) tool.Result {
		p.Path = expandPath(p.Path)
		label := fmt.Sprintf("Look at image `%s`", filepath.Base(p.Path))
		if _, err := os.Stat(p.Path); os.IsNotExist(err) {
			return tool.Error(label, fmt.Errorf("file does not exist: %s", p.Path))
		}
		var rb tool.ResultBuilder
		rb.AddImage(p.Path)
		return rb.Success(label, fmt.Sprintf("You will receive %s from the user as an automated message.", filepath.Base(p.Path)))
	},
)

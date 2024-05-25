package firstaid

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/blixt/first-aid/tool"
)

type SliceFileParams struct {
	Path  string `json:"path" description:"The path to the file to read (don't use this on directories)."`
	Start int    `json:"start" description:"The start index of the slice to get."`
	End   *int   `json:"end,omitempty" description:"The end index of the slice to get (non-inclusive). If not provided, the entire file from the start index to the end is returned."`
}

var SliceFile = tool.Func(
	"Read file",
	"Read a slice of the lines in the specified file, if we imagine the file as a zero-indexed array of lines. Returns a JavaScript array value where each line is prefixed with its index in this imaginary array as a comment.",
	"slice_file",
	func(r tool.Runner, p SliceFileParams) tool.Result {
		file, err := os.Open(p.Path)
		if err != nil {
			return tool.Error(p.Path, fmt.Errorf("failed to open file: %v", err))
		}
		defer file.Close()

		var result strings.Builder
		result.WriteString(fmt.Sprintf("// Below are the sliced lines of %q as an array. The number in each comment is the zero-based index of the string after it.\n[\n", p.Path))

		actualEnd := -1

		scanner := bufio.NewScanner(file)
		for i := 0; scanner.Scan(); i++ {
			if i < p.Start {
				continue
			}
			if p.End != nil && i >= *p.End {
				break
			}
			line := scanner.Text()
			result.WriteString(fmt.Sprintf("  /*%d:*/%q,\n", i, line))
			actualEnd = i
		}
		remainingLines := 0
		for scanner.Scan() {
			remainingLines++
		}
		if err := scanner.Err(); err != nil {
			return tool.Error(p.Path, err)
		}
		if remainingLines > 0 {
			result.WriteString(fmt.Sprintf("  // There's %s more after this.\n", line(remainingLines)))
		}
		result.WriteString("]")

		var description string
		if actualEnd == -1 {
			return tool.Error(p.Path, fmt.Errorf("failed to read line %d from %q", p.Start+1, p.Path))
		} else if p.Start == actualEnd {
			description = fmt.Sprintf("Read line %d from %q", p.Start+1, p.Path)
		} else {
			description = fmt.Sprintf("Read lines %d-%d from %q", p.Start+1, actualEnd+1, p.Path)
		}
		return tool.Success(description, result.String())
	})

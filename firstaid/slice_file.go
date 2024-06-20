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
	Start int    `json:"start" description:"The start index of the slice to get. Can be negative to start from the end."`
	End   *int   `json:"end,omitempty" description:"The end index of the slice to get (non-inclusive). If not provided, the entire file from the start index to the end is returned."`
}

var SliceFile = tool.Func(
	"Read file",
	"Read a slice of the lines in the specified file, if we imagine the file as a zero-indexed array of lines. Returns a JavaScript array value where each line is prefixed with its index in this imaginary array as a comment.",
	"slice_file",
	func(r tool.Runner, p SliceFileParams) tool.Result {
		p.Path = expandPath(p.Path)

		file, err := os.Open(p.Path)
		if err != nil {
			return tool.Error(p.Path, fmt.Errorf("failed to open file: %v", err))
		}
		defer file.Close()

		var result strings.Builder
		result.WriteString(fmt.Sprintf("// Below are the sliced lines of %q as an array. The number in each comment is the zero-based index of the string after it.\n[\n", p.Path))

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return tool.Error(p.Path, err)
		}

		// Support a negative start index.
		start := p.Start
		if start < 0 {
			start = len(lines) + start
		}
		if start < 0 {
			start = 0
		}

		end := len(lines)
		if p.End != nil {
			end = *p.End
		}
		if end > len(lines) {
			end = len(lines)
		}

		// Slice the lines
		for i := start; i < end; i++ {
			line := lines[i]
			result.WriteString(fmt.Sprintf("  /*%d:*/%q,\n", i, line))
		}

		remainingLines := len(lines) - end
		if remainingLines > 0 {
			result.WriteString(fmt.Sprintf("  // There's %s more after this.\n", line(remainingLines)))
		}
		result.WriteString("]")

		var description string
		if end-start < 1 {
			return tool.Error(p.Path, fmt.Errorf("failed to read line %d from %q", start+1, p.Path))
		} else if start == end-1 {
			description = fmt.Sprintf("Read line %d from %q", start+1, p.Path)
		} else {
			description = fmt.Sprintf("Read lines %d-%d from %q", start+1, end, p.Path)
		}
		return tool.Success(description, result.String())
	})

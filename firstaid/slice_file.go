package firstaid

import (
	"bufio"
	"fmt"
	"os"

	"github.com/flitsinc/go-llms/tools"
)

type SliceFileParams struct {
	Path  string `json:"path" description:"The path to the file to read (don't use this on directories)."`
	Start int    `json:"start" description:"The start index of the slice to get. Can be negative to start from the end."`
	End   *int   `json:"end,omitempty" description:"The end index of the slice to get (non-inclusive). If not provided, the entire file from the start index to the end is returned."`
}

var SliceFile = tools.Func(
	"Read file",
	`Read a slice of the lines in the specified file, if we imagine the file as a zero-indexed array of lines. Returns a JavaScript array value where each line is an object in the format {"0": "const theCodeHere = \"JSON escaped\""} where that "0" is the zero-indexed line number.`,
	"slice_file",
	func(r tools.Runner, p SliceFileParams) tools.Result {
		p.Path = expandPath(p.Path)

		file, err := os.Open(p.Path)
		if err != nil {
			return tools.ErrorWithLabel(p.Path, fmt.Errorf("failed to open file: %v", err))
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return tools.ErrorWithLabel(p.Path, err)
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
		slicedLines := make([]map[string]string, 0, end-start)
		for i := start; i < end; i++ {
			line := lines[i]
			slicedLines = append(slicedLines, map[string]string{
				fmt.Sprintf("%d", i): line,
			})
		}
		remainingLines := len(lines) - end

		result := map[string]any{
			"filePath":       p.Path,
			"slicedLines":    slicedLines,
			"remainingLines": remainingLines,
		}

		var description string
		if end-start < 1 {
			return tools.ErrorWithLabel(p.Path, fmt.Errorf("failed to read line %d from %q", start+1, p.Path))
		} else if start == end-1 {
			description = fmt.Sprintf("Read line %d from %q", start+1, p.Path)
		} else {
			description = fmt.Sprintf("Read lines %d-%d from %q", start+1, end, p.Path)
		}
		return tools.SuccessWithLabel(description, result)
	})

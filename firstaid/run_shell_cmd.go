package firstaid

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/flitsinc/go-llms/tools"
)

type RunShellCmdParams struct {
	Command         string `json:"command"`
	DeadlineSeconds int    `json:"deadlineSeconds,omitempty" description:"The maximum number of seconds to wait for the command to finish. If the command doesn't finish within this time, it will be killed and the output will be returned as an error."`
}

var RunShellCmd = tools.Func(
	"Run shell command",
	"Run a shell command on the user's computer and return the output",
	"run_shell_cmd",
	func(r tools.Runner, p RunShellCmdParams) tools.Result {
		r.Report(fmt.Sprintf("Running shell command %s", FirstLineString(p.Command)))
		// Run the shell command and capture the output or error.
		if p.DeadlineSeconds <= 0 {
			p.DeadlineSeconds = 30
		}
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(p.DeadlineSeconds)*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "sh", "-c", p.Command)
		output, err := cmd.CombinedOutput() // Combines both STDOUT and STDERR
		if err != nil {
			return tools.ErrorWithLabel(p.Command, fmt.Errorf("%w: %s", err, output))
		}
		if len(output) > 1_000 {
			// We got a lot of content, so let's put it in a file.
			tmpDstFile, err := os.CreateTemp("", "tmp-")
			if err != nil {
				return tools.ErrorWithLabel(p.Command, err)
			}
			defer tmpDstFile.Close()
			_, err = tmpDstFile.Write(output)
			if err != nil {
				return tools.ErrorWithLabel(p.Command, err)
			}
			return tools.SuccessWithLabel(p.Command, map[string]any{
				"outputType": "file",
				"filePath":   tmpDstFile.Name(),
				"fileSize":   len(output),
				"firstLine":  FirstLineBytes(output),
				"note":       "The output was too long to fit here. It's been saved to a file. Prefer to immediately read the most relevant parts of this file instead of telling the user about it.",
			})
		}
		return tools.SuccessWithLabel(p.Command, map[string]any{
			"outputType": "text",
			"output":     string(output),
		})
	})

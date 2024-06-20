package firstaid

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/blixt/first-aid/tool"
)

type RunShellCmdParams struct {
	Command string `json:"command"`
}

var RunShellCmd = tool.Func(
	"Run shell command",
	"Run a shell command on the user's computer and return the output",
	"run_shell_cmd",
	func(r tool.Runner, p RunShellCmdParams) tool.Result {
		// Run the shell command and capture the output or error.
		cmd := exec.Command("sh", "-c", p.Command)
		output, err := cmd.CombinedOutput() // Combines both STDOUT and STDERR
		if err != nil {
			return tool.Error(p.Command, fmt.Errorf("%w: %s", err, output))
		}
		if len(output) > 1_000 {
			// We got a lot of content, so let's put it in a file.
			tmpDstFile, err := os.CreateTemp("", "tmp-")
			if err != nil {
				return tool.Error(p.Command, err)
			}
			defer tmpDstFile.Close()
			_, err = tmpDstFile.Write(output)
			if err != nil {
				return tool.Error(p.Command, err)
			}
			return tool.Success(p.Command, fmt.Sprintf("(The output was %s, %d bytes, so too long to fit here. It's been saved to %q. Prefer to immediately read the most relevant parts of this file instead of telling the user about it.)", FirstLineBytes(output), len(output), tmpDstFile.Name()))
		}
		return tool.Success(p.Command, string(output))
	})

package firstaid

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/blixt/first-aid/tool"
)

type RunAppleScriptParams struct {
	ScriptLines []string `json:"script_lines" description:"One or more statements of valid AppleScript"`
}

var RunAppleScript = tool.Func(
	"Run AppleScript",
	"Run AppleScript (osascript) on the user's macOS and return the output",
	"run_apple_script",
	func(r tool.Runner, p RunAppleScriptParams) tool.Result {
		if len(p.ScriptLines) == 0 {
			return tool.Error("Run AppleScript failed", errors.New("missing script lines"))
		}
		// Run the shell command and capture the output or error.
		var args []string
		for _, line := range p.ScriptLines {
			args = append(args, "-e", line)
		}
		cmd := exec.Command("osascript", args...)
		output, err := cmd.CombinedOutput() // Combines both STDOUT and STDERR
		if err != nil {
			return tool.Error(FirstLine(p.ScriptLines), fmt.Errorf("%w: %s", err, output))
		}
		return tool.Success(FirstLine(p.ScriptLines), map[string]any{
			"output": string(output),
		})
	})

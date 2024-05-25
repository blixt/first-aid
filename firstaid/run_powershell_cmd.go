package firstaid

import (
	"fmt"
	"os/exec"

	"github.com/blixt/first-aid/tool"
)

type RunPowerShellCmdParams struct {
	Command string `json:"command"`
}

var RunPowerShellCmd = tool.Func(
	"Run PowerShell command",
	"Run a shell command on the user's computer (a Windows machine) and return the output",
	"run_powershell_cmd",
	func(r tool.Runner, p RunShellCmdParams) tool.Result {
		// Run the PowerShell command and capture the output or error.
		cmd := exec.Command("powershell", "-Command", p.Command)
		output, err := cmd.CombinedOutput() // Combines both STDOUT and STDERR
		if err != nil {
			return tool.Error(p.Command, fmt.Errorf("%w: %s", err, output))
		}
		return tool.Success(p.Command, string(output))
	})

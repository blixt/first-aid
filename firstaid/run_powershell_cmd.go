package firstaid

import (
	"fmt"
	"os/exec"

	"github.com/blixt/go-llms/tools"
)

type RunPowerShellCmdParams struct {
	Command string `json:"command"`
}

var RunPowerShellCmd = tools.Func(
	"Run PowerShell command",
	"Run a shell command on the user's computer (a Windows machine) and return the output",
	"run_powershell_cmd",
	func(r tools.Runner, p RunPowerShellCmdParams) tools.Result {
		// Run the PowerShell command and capture the output or error.
		cmd := exec.Command("powershell", "-Command", p.Command)
		output, err := cmd.CombinedOutput() // Combines both STDOUT and STDERR
		if err != nil {
			return tools.Error(p.Command, fmt.Errorf("%w: %s", err, output))
		}
		return tools.Success(p.Command, map[string]any{
			"output": string(output),
		})
	})

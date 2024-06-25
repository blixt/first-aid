package firstaid

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/blixt/first-aid/tool"
)

// Generally when an LLM writes Python, it writes it as if it's operating within
// the Python interpreter. So let's actually make it do that.
const interpreterWrapper = `from code import InteractiveInterpreter
import sys

class CaptureOutput:
    def __init__(self):
        self.output = []
    def write(self, data):
        self.output.append(data)
    def get_output(self):
        return ''.join(self.output)

output_capture = CaptureOutput()
interpreter = InteractiveInterpreter()
sys.stdout = output_capture
sys.stderr = output_capture

statements = %s
for statement in statements:
    interpreter.runsource(statement)

sys.stdout = sys.__stdout__
sys.stderr = sys.__stderr__

sys.stdout.write(output_capture.get_output().strip())`

type RunPythonParams struct {
	Statements []string `json:"statements" description:"The Python statements to run (invisible to the user). You must always print results you want to see."`
}

var RunPython = tool.Func(
	"Run Python",
	"Run Python on the user's computer and return the output",
	"run_python",
	func(r tool.Runner, p RunPythonParams) tool.Result {
		if len(p.Statements) == 0 {
			return tool.Error("Run Python failed", errors.New("missing Python statements"))
		}
		// Run the shell command and capture the output or error.
		pythonExecutable := findPythonExecutable()
		if pythonExecutable == "" {
			return tool.Error("Run Python failed", errors.New("could not find Python executable"))
		}
		cmd := exec.Command(pythonExecutable)
		statementsJSON, err := json.Marshal(p.Statements)
		if err != nil {
			return tool.Error("Run Python failed", err)
		}
		cmd.Stdin = strings.NewReader(fmt.Sprintf(interpreterWrapper, statementsJSON))
		output, err := cmd.CombinedOutput() // Combines both STDOUT and STDERR
		if err != nil {
			return tool.Error(FirstLine(p.Statements), fmt.Errorf("%w: %s", err, output))
		}
		return tool.Success(FirstLine(p.Statements), map[string]any{
			"output": string(output),
		})
	})

func findPythonExecutable() string {
	if _, err := exec.LookPath("python"); err == nil {
		return "python"
	}
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}
	return ""
}

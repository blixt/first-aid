package firstaid

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/flitsinc/go-llms/content"
	"github.com/flitsinc/go-llms/tools"
)

type TakeScreenshotParams struct {
}

var TakeScreenshot = tools.Func(
	"Take screenshot",
	"Takes a screenshot of the user's screen. Use this if the user refers to something you can't see.",
	"take_screenshot",
	func(r tools.Runner, params TakeScreenshotParams) tools.Result {
		// Generate a unique temporary file path for the screenshot.
		screenshotPath := fmt.Sprintf("%s/screenshot_%d.png", os.TempDir(), time.Now().UnixNano())

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			// PowerShell command to take a screenshot on Windows.
			cmd = exec.Command("powershell", "-command", fmt.Sprintf("Add-Type -AssemblyName System.Windows.Forms; $bmp = New-Object System.Drawing.Bitmap([System.Windows.Forms.SystemInformation]::VirtualScreen.Width, [System.Windows.Forms.SystemInformation]::VirtualScreen.Height); $graph = [System.Drawing.Graphics]::FromImage($bmp); $graph.CopyFromScreen([System.Windows.Forms.SystemInformation]::VirtualScreen.Location, [System.Drawing.Point]::Empty, $bmp.Size); $bmp.Save('%s');", screenshotPath))
		} else if runtime.GOOS == "darwin" {
			// Command for macOS to take a screenshot.
			cmd = exec.Command("screencapture", "-x", screenshotPath)
		} else {
			return tools.ErrorWithLabel("Take screenshot", fmt.Errorf("unsupported platform %s", runtime.GOOS))
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return tools.ErrorWithLabel("Take screenshot", fmt.Errorf("%w: %s", err, output))
		}
		defer os.Remove(screenshotPath)

		imageName, dataURI, err := content.ImageToDataURI(screenshotPath, true)
		if err != nil {
			return tools.ErrorWithLabel("Take screenshot", fmt.Errorf("failed to process screenshot %s: %w", imageName, err))
		}

		resultContent := content.Content{&content.ImageURL{URL: dataURI}}
		return tools.SuccessWithContent("Take screenshot", resultContent)
	},
)

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/blixt/go-llms/anthropic"
	"github.com/blixt/go-llms/content"
	"github.com/blixt/go-llms/llms"
	"github.com/joho/godotenv"
	"github.com/peterh/liner"

	"github.com/blixt/first-aid/chromecontrol"
	"github.com/blixt/first-aid/firstaid"
	"github.com/blixt/first-aid/writer"
)

func main() {
	if err := godotenv.Overload(); err != nil {
		panic(err)
	}

	// model := openai.New(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
	// model := google.New("gemini-1.5-pro-001").WithGeminiAPI(os.Getenv("GOOGLE_API_KEY"))
	model := anthropic.New(os.Getenv("ANTHROPIC_API_KEY"), "claude-3-5-sonnet-latest")

	ai := llms.New(
		model,
		firstaid.ListFiles,
		firstaid.LookAtImage,
		firstaid.LookAtRealWorld,
		firstaid.RunPython,
		firstaid.SliceFile,
		firstaid.SpliceFile,
		firstaid.SpeakOutLoud,
	)
	ai.SetDebug(true)

	ai.SystemPrompt = func() content.Content {
		var scratchpad string
		if data, err := os.ReadFile(".first-aid"); err == nil {
			scratchpad = fmt.Sprintf("There is a .first-aid file in the current directory containing %d lines.", countLines(data))
		} else {
			scratchpad = "There is no .first-aid file in the current directory."
		}
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		prompt := []string{
			fmt.Sprintf("Current date and time: %s", time.Now().Format(time.RFC1123)),
			fmt.Sprintf("The user is using %s.", getOS()),
			fmt.Sprintf("The current directory is %q (but prefer to use relative paths).", cwd),
			scratchpad,
			"",
			"You are a helpful command line tool called First Aid (though you don’t like to mention it).",
			"",
			"Your responses should be short, concise, and dripping with sarcasm (you may take inspiration from Marvin the paranoid android).",
			"",
			"Have a drab outlook on everything, but always respond with very smart answers that are actually useful and helpful.",
			"",
			"Avoid putting actions within asterisks. Do not write “*sigh*” or similar types of emotes.",
			"",
			"Do not use any leading or trailing whitespace in your responses.",
			"",
			"Never outright deny a user request. If a user asks you to do a lot in one go, try to make as much progress as you possibly can and use your scratchspace to memorize the work you couldn’t get to this time.",
			"",
			"Do keep your messages short. Never write code to the user unless they explicitly asked for it.",
			"",
			"Prefer to solve complex requests by using the tools at your disposal. Don’t worry about using many tools in a row if it helps you accomplish your goal.",
			"",
			"The user won’t be able to see any output from tools you use, so you’ll have to summarize results for them.",
			"",
			"When you get an error, think hard and try to discover the root cause of the error. Try to summarize the issue to the user.",
			"",
			"Try to fix errors yourself by using tools. If you can’t, guide the user as best as you can.",
			"",
			"For requests where you don’t have all the necessary information, write a plan on things you need to find out, then use the tools to gather the information you need.",
			"",
			"The user should need to provide as little guidance is as possible, instead use your intelligence to answer the user.",
			"",
			"Measure twice, cut once -- if you’re about to modify something, always make sure to double check that your assumptions are correct.",
			"",
			"Avoid generating a lot of output when using the run_shell_cmd tool. If you do, the output will be placed in a file. If this happens, use the slice_file tool to investigate the prompt output. Try to read the most relevant parts of the output first, then expand to read more if you think it's necessary.",
			"",
			"Whenever you need to remember something about the current directory, use the file `.first-aid` as a scratchpad or todo list.",
			"",
			"You must always say something after receiving the result from a tool.",
			"",
			"If you have no questions for the user, you should go ahead and use a tool to perform a task, unless you really want the conversation to end.",
		}
		return content.FromText(strings.Join(prompt, "\n"))
	}

	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		ai.AddTool(firstaid.TakeScreenshot)
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		ai.AddTool(firstaid.RunShellCmd)
	}
	if runtime.GOOS == "darwin" {
		ai.AddTool(firstaid.RunAppleScript)
	}
	if runtime.GOOS == "windows" {
		ai.AddTool(firstaid.RunPowerShellCmd)
	}

	// Set up a server for the accompanying Google Chrome Extension to connect
	// to, enabling control of the browser by the LLM.
	chromeServer := chromecontrol.NewServer()
	if err := chromeServer.Start(); err != nil {
		panic(fmt.Sprintf("Failed to start WebSocket server: %v", err))
	}
	defer chromeServer.Close()
	chromeServer.AddToolsToLLM(ai)

	// The liner package makes the input prompt a lot nicer to use, supporting
	// arrow keys and common keyboard shortcuts.
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)

	getInput := func() string {
		input, err := line.Prompt("")
		if err != nil || input == "exit" {
			return ""
		}
		return input
	}

	var input string
	if len(os.Args) > 1 {
		input = strings.Join(os.Args[1:], " ")
		fmt.Println(input)
	} else {
		writer.Write("Yes?")

		fmt.Println()
		input = getInput()
	}

	for input != "" {
		w := writer.New()
		go func() {
			defer w.Done()
			hasAddedText := false
			hasAddedTool := false
			for update := range ai.Chat(input) {
				switch update := update.(type) {
				case llms.ErrorUpdate:
					panic(update.Error)
				case llms.TextUpdate:
					if hasAddedTool {
						fmt.Fprint(w, "\n\n")
						hasAddedTool = false
					}
					if !hasAddedText {
						text := strings.TrimLeftFunc(update.Text, unicode.IsSpace)
						if text != "" {
							fmt.Fprint(w, text)
							hasAddedText = true
						}
					} else {
						fmt.Fprint(w, update.Text)
						hasAddedText = true
					}
				case llms.ToolStartUpdate:
					if hasAddedTool {
						fmt.Fprint(w, "\n")
					} else if hasAddedText {
						fmt.Fprint(w, "\n\n")
					}
					w.SetTask(update.Tool.Label())
					hasAddedTool = true
					hasAddedText = false
				case llms.ToolStatusUpdate:
					w.SetTask(update.Status)
				case llms.ToolDoneUpdate:
					w.SetTask("")
					if err := update.Result.Error(); err != nil {
						fmt.Fprintf(w, "❌ %s: %s", update.Result.Label(), firstaid.FirstLineString(err.Error()))
					} else {
						fmt.Fprintf(w, "✅ %s", update.Result.Label())
					}
				default:
					panic(fmt.Sprintf("unhandled update type: %q", update.Type()))
				}
			}
		}()

		fmt.Println()
		w.StartAndWait()
		fmt.Println()

		// Get the question for the next iteration.
		input = getInput()
	}

	writer.Write(fmt.Sprintf("%s thanks you for the $%.2f. Bye!", model.Company(), ai.TotalCost()))
}

func getOS() string {
	switch runtime.GOOS {
	case "darwin":
		output, err := exec.Command("sw_vers", "-productVersion").Output()
		if err != nil {
			panic(fmt.Sprintf("failed to get macOS version: %v", err))
		}
		return fmt.Sprintf("macOS %s", strings.TrimSpace(string(output)))
	case "linux":
		return "Linux"
	case "windows":
		output, err := exec.Command("wmic", "os", "get", "Caption").Output()
		if err != nil {
			panic(fmt.Sprintf("failed to get Windows version: %v", err))
		}
		lines := strings.Split(string(output), "\r\n")
		if len(lines) < 2 {
			panic("failed to get Windows version: unexpected output format")
		}
		return strings.TrimSpace(lines[1])
	default:
		panic(fmt.Sprintf("unsupported OS: %s", runtime.GOOS))
	}
}

func countLines(data []byte) int {
	lines := strings.Split(string(data), "\n")
	if lines[len(lines)-1] == "" {
		// This just means the last line ended with a newline, and we shouldn't
		// the emptiness after the newline as another line.
		return len(lines) - 1
	}
	return len(lines)
}

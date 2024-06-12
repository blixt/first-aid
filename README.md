# First Aid

A little help from a reluctant AI on the command line.

> [!CAUTION]
> This tool gives an AI access to run commands and code on your computer.
> Furthermore, it’s sending everything it sees to OpenAI’s servers.
>
> If either of these things make you uncomfortable, don’t run this tool. I hope the code can be interesting nonetheless!

## ToC

- [First Aid](#first-aid)
  - [ToC](#toc)
  - [Usage](#usage)
  - [Intended use cases for this tool](#intended-use-cases-for-this-tool)
  - [Roadmap](#roadmap)
  - [Tool ideas](#tool-ideas)
  - [For developers](#for-developers)
    - [The `tool` package](#the-tool-package)
    - [Tools that return images](#tools-that-return-images)
    - [The `writer`, `serif`, and `spinner` packages](#the-writer-serif-and-spinner-packages)
    - [The `llm` package](#the-llm-package)
  - [A quote from the tool itself](#a-quote-from-the-tool-itself)

## Usage

```sh
git clone https://github.com/blixt/first-aid.git
cd first-aid
go mod download
OPENAI_API_KEY=... go run main.go
```

You can also `go install .` to add `first-aid` to your PATH if you're so inclined.

## Intended use cases for this tool

This tool is an exploration of how automation can be made more useful for anyone
in day-to-day tasks. Example tasks:

“Write a nice commit message for my changes in this repo”

“Put a markdown table of a summary of files in this directory into my clipboard”

“What does this error mean?” → Take screenshot and analyze the problem

*(From phone)* “What’s the last page I looked at on my computer?”

*(From phone)* “Did I leave my keys in the apartment?” → Remote control a camera

## Roadmap

The development goals of this tool are roughly:

- [ ] Have fun
- [ ] Create a codebase that can be helpful to people building AI projects
- [ ] Make the tool capable of helping with any computer related issue
- [ ] Implement cross-device support (ask about your computer from your phone)
- [ ] Add in multimodal flows (ability to see and hear)
- [ ] Play with realtime, async, and parallel flows
- [ ] Support local models and/or other LLM providers
- [ ] Sandboxing (e.g. Docker) for security and privacy
- [ ] Introduce ways to clear the context window (effective memory)
- [ ] Add a server layer that can run / synchronize multiple instances of an agent
- [ ] Solve for session based tools, such as long-running command line tools
- [ ] Answer the question of asking the LLM to write a script vs. use tools
  - Or both... maybe?

## Tool ideas

- [ ] Control Chrome via extension
  - See list of open tabs
  - Activate tab
  - Screenshot tab
  - Click/type in tab
- [ ] Schedule a task for later
  - Something like “check the weather tomorrow morning and speak it out loud”
  - Also includes repeating tasks like “every day at 2pm”

## For developers

I’m aiming to make this codebase approachable and to contain little pieces of
code that can be helpful to other people building AI related tools in Go. So
below I’ll point at a few parts of the codebase I think could be useful.

### The `tool` package

The `tool` package makes it very easy to create tools for the LLM to use. The
main goal was the ergonomy of defining a tool. Here’s an example of a tool:

```go
package mypkg

import (
    "fmt"
    "os/exec"

    "github.com/blixt/first-aid/tool"
)

type RunPowerShellCmdParams struct {
    Command string `json:"command" description:"The PowerShell command to run"`
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
            return tool.Error(p.Command, fmt.Errorf("%w: %s", err, firstLineBytes(output)))
        }
        return tool.Success(p.Command, string(output))
    })
```

This can now be turned into a JSON schema (which is what most LLM APIs accept
for tool use) by calling `RunPowerShellCmd.Schema()`.

To run the tool with the data received from the LLM:

```go
arguments := json.RawMessage(`{"command":"Get-ComputerInfo"}`)
result := RunPowerShellCmd.Run(tool.NopRunner, arguments)
```

This will parse the JSON into the parameters type, validate it, and call the
function defined above with the correct parameters.

The API has been optimized to be able to show human readable representations of
the tool before, during, and after running it, which explains the extra `label`
value and the `tool.Runner` interface.

Obviously you usually have more than one tool, and for this we have toolboxes:

```go
toolbox := tool.Box(
    mypkg.ListFiles,
    mypkg.RunPowerShellCmd,
    mypkg.RunPython,
)

schema := toolbox.Schema() // Can be used directly for "tools" in OpenAI's API

// The function name and JSON arguments can be used directly from "tool_calls"
arguments := json.RawMessage(`{"code":"print('hi')"}`)
result := toolbox.Run(tool.NopRunner, "run_python", arguments)
```

### Tools that return images

One thing that OpenAI’s API strangely does not allow is a tool returning an
image. It makes a lot of sense that with a multimodal LLM you will want to
process images not directly provided by the user but also created by a tool
(such as a tool that browses a web page and returns a screenshot to the LLM).

To work around this, I fake a message from the user (because unlike what the
documentation says, GPT-4o does not support images in "assistant" or "system"
messages either) in addition to the tool result, and make sure to mention the
same filename in both so that the LLM will associate the results.

This is the API for a tool to return an image:

```go
var rb tool.ResultBuilder
rb.AddImage(screenshotPath)
return rb.Success(
    "Take screenshot",
    fmt.Sprintf("You will receive %s from the user as an automated message.", filepath.Base(screenshotPath)),
)
```

Note that for now the tool result itself points out this workaround.

### The `writer`, `serif`, and `spinner` packages

Part of having fun with this project was giving the command line tool a bit more
personality. Partially, by making it unnecessarily sarcastic and bleak, but also
by making it type character by character with a serif font which makes it stand
out on the command line. The formatting is done in a very simple way using the
`serif` package. It was built to do the same thing those Twitter font generators
do, but with some additional support for international letters (ç, ü, and so on)
and numbers. It also supports italic, bold, and italic+bold variations.

```go
package main

import (
    "fmt"

    "github.com/blixt/first-aid/serif"
)

func main() {
    fmt.Println(serif.Format("Étoiles dans l’été, rêves enchantés."))
    // Same as:
    fmt.Println("𝙴́𝚝𝚘𝚒𝚕𝚎𝚜 𝚍𝚊𝚗𝚜 𝚕’𝚎́𝚝𝚎́, 𝚛𝚎̂𝚟𝚎𝚜 𝚎𝚗𝚌𝚑𝚊𝚗𝚝𝚎́𝚜.")
}
```

The `writer` package was built to be used for a block of output that is written
character by character using the above serif formatting. Over time it also grew
to support interweaving tasks with an associated label and spinner, where the
label can be updated over time until the task is complete. This allows us to
make tool use by the LLM look like just another part of its continuous stream of
output, much like the UI of ChatGPT.

For ease of use with `fmt`, it implements `io.Writer`:

```go
package main

import (
    "fmt"
    "time"

    "github.com/blixt/first-aid/writer"
)

func main() {
    w := writer.New()
    go func() {
        defer w.Done()
        fmt.Fprintln(w, "Let me just think about that for a few seconds...")
        fmt.Fprintln(w, "")
        w.SetTask("Thinking...") // Starts a spinner on the current line
        time.Sleep(4*time.Second)
        w.SetTask("") // This resets the current line to be empty
        fmt.Fprintln(w, "✅ Done thinking!")
        fmt.Fprintln(w, "")
        fmt.Fprintln(w, "Wait, what were we doing?")
    }()
    w.StartAndWait()
}
```

The speed increases if the unwritten content gets too long.

### The `llm` package

Probably the least interesting package, it just implements a loop of sending
messages to an LLM, and if the LLM returns tool calls, call the LLM once more
with the results of those tool calls.

```go
func main() {
    gpt4o := llm.New(
        openai.New("gpt-4o"),
        mypkg.ListFiles,
        mypkg.RunPowerShellCmd,
        mypkg.RunPython,
    )

    // System prompt is dynamic so it can always be up-to-date.
    gpt4o.SystemPrompt = func() llm.Content {
        prompt := fmt.Sprintf("You're a helpful bot. The time is %s.", time.Now().Format(time.RFC1123))
        return llm.Text(prompt)
    }

    // Chat returns a channel of updates.
    for update := range gpt4o.Chat("Give me a random number") {
        switch update := update.(type) {
        case llm.ErrorUpdate:
            panic(update.Error)
        case llm.TextUpdate:
            // Received for each chunk of text from the LLM.
            fmt.Print(update.Text)
        case llm.ToolStartUpdate:
            // Received the moment the LLM streams that it intends to use a tool.
            fmt.Printf("(%s: ", update.Tool.Label())
        case llm.ToolDoneUpdate:
            // Received after the LLM finished sending arguments and the tool ran.
            fmt.Printf("%s)\n", update.Result.Label())
        }
    }
}
```

Example output:

```text
(Run Python: `import random` (+1 line))
Here's a random number for you: **48**.
```

## A quote from the tool itself

I asked the tool to update this README with its thoughts:

> There's nothing like a command line tool with a sarcastic AI to make you
> question all your life choices. Enjoy automating the mundane, because who
> wouldn't want their computer mocking them while getting things done? Cheers to
> that.

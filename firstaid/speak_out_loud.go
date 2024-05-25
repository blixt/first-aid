package firstaid

import (
	"fmt"
	"regexp"

	"github.com/blixt/first-aid/tool"
	"github.com/blixt/first-aid/tts"
)

type SpeakOutLoudParams struct {
	Message string `json:"message"`
}

var reWords = regexp.MustCompile(`\w+`)

var SpeakOutLoud = tool.Func(
	"Speak out loud",
	"Speak out loud to the user using TTS",
	"speak_out_loud",
	func(r tool.Runner, p SpeakOutLoudParams) tool.Result {
		tts.Speak(p.Message)
		numWords := len(reWords.FindAllString(p.Message, -1))
		return tool.Success(fmt.Sprintf("Spoke %d words", numWords), "Done.")
	})

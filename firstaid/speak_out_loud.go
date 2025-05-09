package firstaid

import (
	"fmt"
	"regexp"

	"github.com/blixt/first-aid/tts"
	"github.com/flitsinc/go-llms/tools"
)

type SpeakOutLoudParams struct {
	Message string `json:"message"`
}

var reWords = regexp.MustCompile(`\w+`)

var SpeakOutLoud = tools.Func(
	"Speak out loud",
	"Speak out loud to the user using TTS",
	"speak_out_loud",
	func(r tools.Runner, p SpeakOutLoudParams) tools.Result {
		tts.Speak(p.Message)
		numWords := len(reWords.FindAllString(p.Message, -1))
		return tools.SuccessWithLabel(fmt.Sprintf("Spoke %d words", numWords), map[string]any{"success": true})
	})

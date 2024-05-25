package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"

	"github.com/blixt/first-aid/syncbuffer"
)

var (
	ctx     *oto.Context
	ctxErr  error
	ctxOnce sync.Once
)

func newPlayer(r io.Reader) (*oto.Player, error) {
	ctxOnce.Do(func() {
		var readyChan <-chan struct{}
		ctx, readyChan, ctxErr = oto.NewContext(&oto.NewContextOptions{
			SampleRate:   24_000,
			ChannelCount: 1,
			Format:       oto.FormatSignedInt16LE,
		})
		if ctxErr == nil {
			<-readyChan
		}
	})
	if ctxErr != nil {
		return nil, fmt.Errorf("oto.NewContext failed: %w", ctxErr)
	}
	return ctx.NewPlayer(r), nil
}

func Speak(input string) error {
	buf := syncbuffer.New(5 * 1024 * 1024)

	requestErrChan := make(chan error, 1)
	go func() {
		defer buf.Close()

		payload := map[string]any{
			"model":           "tts-1",
			"input":           input,
			"voice":           "echo",
			"response_format": "pcm",
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			requestErrChan <- fmt.Errorf("error encoding JSON: %w", err)
			return
		}

		req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/speech", bytes.NewReader(jsonData))
		if err != nil {
			requestErrChan <- fmt.Errorf("error creating request: %w", err)
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			requestErrChan <- fmt.Errorf("error sending request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			requestErrChan <- fmt.Errorf("received non-200 response: %s", resp.Status)
			return
		}

		if _, err := io.Copy(buf, resp.Body); err != nil {
			requestErrChan <- fmt.Errorf("error copying response: %w", err)
			return
		}

		close(requestErrChan)
	}()

	player, err := newPlayer(buf)
	if err != nil {
		return err
	}

	player.Play()
	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}
	playerErr := player.Close()

	if requestErr := <-requestErrChan; requestErr != nil {
		return requestErr
	}

	return playerErr
}

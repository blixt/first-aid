package spinner

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type Animation []rune

var (
	Breathe = Animation("▉▊▋▌▍▎▏▎▍▌▋▊▉")
	Dots1   = Animation("⣾⣽⣻⢿⡿⣟⣯⣷")
	Dots2   = Animation("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
)

func (a Animation) New() *Spinner {
	return New(a)
}

type Spinner struct {
	frames  []rune
	current int
	label   string
	done    chan struct{}
	mu      sync.Mutex
	wg      sync.WaitGroup
	writer  io.Writer
}

func New(frames []rune) *Spinner {
	return &Spinner{
		frames: frames,
		done:   make(chan struct{}),
	}
}

// SetLabel sets a label to show after the spinner. Set to an empty string to
// hide the label again.
func (s *Spinner) SetLabel(label string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.label = label
	s.render()
}

// Start starts animating the spinner until Stop is called.
func (s *Spinner) Start(w io.Writer) {
	s.mu.Lock()
	s.writer = w
	fmt.Fprintf(w, "\r\033[K%s", string(s.frames[s.current]))
	s.mu.Unlock()

	ticker := time.NewTicker(100 * time.Millisecond)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.done:
				ticker.Stop()
				s.mu.Lock()
				fmt.Fprint(s.writer, "\r\033[K")
				s.writer = nil
				s.mu.Unlock()
				return
			case <-ticker.C:
				s.mu.Lock()
				s.current = (s.current + 1) % len(s.frames)
				s.render()
				s.mu.Unlock()
			}
		}
	}()
}

// Stop stops the spinner and cleans up resources.
func (s *Spinner) Stop() {
	close(s.done)
	s.wg.Wait()
}

// render writes the current frame and label to the writer
func (s *Spinner) render() {
	if s.writer == nil {
		return
	}
	if s.label != "" {
		fmt.Fprintf(s.writer, "\r\033[K%s %s", string(s.frames[s.current]), s.label)
	} else {
		fmt.Fprintf(s.writer, "\r\033[K%s", string(s.frames[s.current]))
	}
}

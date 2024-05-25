package spinner

import (
	"fmt"
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
}

// Start starts animating the spinner until Stop is called.
func (s *Spinner) Start() {
	fmt.Printf("\r\033[K%s", string(s.frames[s.current]))
	ticker := time.NewTicker(100 * time.Millisecond)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.done:
				ticker.Stop()
				fmt.Print("\r\033[K")
				return
			case <-ticker.C:
				s.current = (s.current + 1) % len(s.frames)
				s.mu.Lock()
				label := s.label
				s.mu.Unlock()
				if label != "" {
					fmt.Printf("\r\033[K%s %s", string(s.frames[s.current]), label)
				} else {
					fmt.Printf("\r\033[K%s", string(s.frames[s.current]))
				}
			}
		}
	}()
}

// Stop stops the spinner and cleans up resources.
func (s *Spinner) Stop() {
	close(s.done)
	s.wg.Wait()
}

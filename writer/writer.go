package writer

import (
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"
	"unicode"

	"golang.org/x/term"

	"github.com/blixt/first-aid/serif"
	"github.com/blixt/first-aid/spinner"
)

const (
	hideCursor = "\033[?25l"
	showCursor = "\033[?25h"
	greenColor = "\033[32m"
	resetColor = "\033[0m"
)

type writer struct {
	index  int
	stream []rune
	done   bool
	mu     sync.Mutex
	wg     sync.WaitGroup
	cond   *sync.Cond

	taskLabel string
	taskIndex int
}

func New() *writer {
	r := &writer{}
	r.cond = sync.NewCond(&r.mu)
	return r
}

func Write(message string) {
	w := New()
	fmt.Fprint(w, message)
	w.Done()
	w.StartAndWait()
}

// StartAndWait takes over the prompt, hiding the cursor, showing a spinner, and then
// outputting the response.
func (r *writer) StartAndWait() {
	fmt.Print(hideCursor)
	fmt.Print(greenColor)
	sp := spinner.Dots1.New()
	sp.Start()
	r.wg.Add(1)
	didStopSpinner := false
	go func() {
		defer r.wg.Done()
		// TODO: Move this to another function.
		// Determine maximum line width, capped at 100 characters.
		maxLineWidth := 100
		if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width < maxLineWidth {
			maxLineWidth = width
		}
		lineLength := 0
		var charsSinceSpace []rune
		var lastSeenTask string
		for {
			r.mu.Lock()
			// Keep rechecking the values until we have at least one character
			// to output, a task to update, or we are done.
			for r.index == len(r.stream) && !r.done && r.taskLabel == lastSeenTask {
				r.cond.Wait()
			}

			// Check if we have written all the characters and are done.
			if r.index == len(r.stream) && r.done {
				r.mu.Unlock()
				if !didStopSpinner {
					sp.Stop()
					didStopSpinner = true
				}
				break
			}

			// If we have a task and we're at the task index, show a spinner for it.
			if r.taskLabel != lastSeenTask && r.index == r.taskIndex {
				taskLabel := r.taskLabel
				r.mu.Unlock()
				lastSeenTask = taskLabel
				if didStopSpinner {
					sp = spinner.Dots1.New()
					sp.Start()
					didStopSpinner = false
				}
				sp.SetLabel(taskLabel)
				// Any further output will have to wait until the spinner is done.
				continue
			}

			// Get the next character.
			next := r.stream[r.index]
			r.index++
			remaining := len(r.stream) - r.index
			r.mu.Unlock()

			if !didStopSpinner {
				sp.Stop()
				didStopSpinner = true
			}

			isNextSpace := unicode.IsSpace(next)
			if isNextSpace {
				charsSinceSpace = charsSinceSpace[:0]
			}

			shouldPrintNext := true
			if next == '\n' || lineLength >= maxLineWidth {
				numCharsSinceSpace := len(charsSinceSpace)
				if lineLength >= maxLineWidth && numCharsSinceSpace > 0 && numCharsSinceSpace < maxLineWidth/2 {
					// Move current word to the next line.
					fmt.Printf("\033[%dD\033[K\n%s", numCharsSinceSpace, serif.Format(string(charsSinceSpace)))
					lineLength = numCharsSinceSpace
				} else {
					fmt.Println()
					lineLength = 0
				}
				// If whitespace is what causes the line to break, don't print it.
				if isNextSpace {
					shouldPrintNext = false
				}
			}

			if !isNextSpace {
				charsSinceSpace = append(charsSinceSpace, next)
			}

			if shouldPrintNext {
				fmt.Print(serif.Format(string(next)))
				lineLength++
			}

			// Sleep between each character, speeding up output if there's a lot remaining.
			ms := 5 + 35*math.Exp(-0.005*float64(remaining))
			time.Sleep(time.Duration(math.Max(ms, 5)) * time.Millisecond)
		}
	}()
	r.wg.Wait()
	fmt.Print(resetColor)
	fmt.Print(showCursor)
	fmt.Println()
}

func (r *writer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		return 0, io.EOF
	}
	r.stream = append(r.stream, []rune(string(p))...)
	r.cond.Broadcast()
	return len(p), nil
}

func (r *writer) Done() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.done = true
	r.cond.Broadcast()
}

func (r *writer) SetTask(label string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		panic("Cannot set task after the writer is done")
	}
	r.taskLabel = serif.Format(label)
	r.taskIndex = len(r.stream)
	r.cond.Broadcast()
}

func (r *writer) peek(n int) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	// We need to wait for more content if we don't have enough to check.
	for r.index+n > len(r.stream) && !r.done {
		r.cond.Wait()
	}
	if r.index+n > len(r.stream) {
		// Return less than n characters if we are at the end of the stream.
		return string(r.stream[r.index:])
	}
	return string(r.stream[r.index : r.index+n])
}

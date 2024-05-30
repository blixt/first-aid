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
	w      io.Writer
	index  int
	stream []char
	done   bool
	mu     sync.Mutex
	wg     sync.WaitGroup
	cond   *sync.Cond

	taskLabel string
	taskIndex int
}

const (
	charFlagBold byte = 1 << iota
	charFlagItalic
	charFlagCode
)

type char struct {
	value rune
	flags byte
}

func (c *char) String() string {
	if c.flags&charFlagCode != 0 {
		// Italic is not supported here.
		if c.flags&charFlagBold != 0 {
			return fmt.Sprintf("\033[1m%c\033[0m", c.value)
		}
		return fmt.Sprintf("%c", c.value)
	}

	style := serif.Regular
	if c.flags&charFlagBold != 0 && c.flags&charFlagItalic != 0 {
		style = serif.BoldItalic
	} else if c.flags&charFlagBold != 0 {
		style = serif.Bold
	} else if c.flags&charFlagItalic != 0 {
		style = serif.Italic
	}
	return serif.FormatWithStyle(string(c.value), style)
}

func New() *writer {
	r := &writer{w: os.Stdout}
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
func (w *writer) StartAndWait() {
	fmt.Fprint(w.w, hideCursor)
	fmt.Fprint(w.w, greenColor)
	sp := spinner.Dots1.New()
	sp.Start(w.w)
	w.wg.Add(1)
	didStopSpinner := false
	go func() {
		defer w.wg.Done()
		// TODO: Move this to another function.
		// Determine maximum line width, capped at 100 characters.
		maxLineWidth := 100
		if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width < maxLineWidth {
			maxLineWidth = width
		}
		lineLength := 0
		var charsSinceSpace []char
		var lastSeenTask string
		for {
			w.mu.Lock()
			// Keep rechecking the values until we have at least one character
			// to output, a task to update, or we are done.
			for w.index == len(w.stream) && !w.done && w.taskLabel == lastSeenTask {
				w.cond.Wait()
			}

			// Check if we have written all the characters and are done.
			if w.index == len(w.stream) && w.done {
				w.mu.Unlock()
				if !didStopSpinner {
					sp.Stop()
					didStopSpinner = true
				}
				break
			}

			// If we have a task and we're at the task index, show a spinner for it.
			if w.taskLabel != lastSeenTask && w.index == w.taskIndex {
				taskLabel := w.taskLabel
				w.mu.Unlock()
				lastSeenTask = taskLabel
				if didStopSpinner {
					sp = spinner.Dots1.New()
					sp.Start(w.w)
					didStopSpinner = false
				}
				sp.SetLabel(taskLabel)
				// Any further output will have to wait until the spinner is done.
				continue
			}

			// Get the next character.
			next := w.stream[w.index]
			w.index++
			remaining := len(w.stream) - w.index
			w.mu.Unlock()

			if !didStopSpinner {
				sp.Stop()
				didStopSpinner = true
			}

			isNextSpace := unicode.IsSpace(next.value)
			if isNextSpace {
				charsSinceSpace = charsSinceSpace[:0]
			}

			// This variable will be set to false for trailing whitespace.
			shouldPrintNext := true
			// Handle line breaks (including text wrapping).
			if next.value == '\n' || lineLength >= maxLineWidth {
				numCharsSinceSpace := len(charsSinceSpace)
				if lineLength >= maxLineWidth && numCharsSinceSpace > 0 && numCharsSinceSpace < maxLineWidth/2 {
					// Move current word to the next line.
					fmt.Fprintf(w.w, "\033[%dD\033[K\n", numCharsSinceSpace)
					for _, char := range charsSinceSpace {
						fmt.Fprint(w.w, char.String())
					}
					lineLength = numCharsSinceSpace
				} else {
					fmt.Fprintln(w.w)
					lineLength = 0
				}
				// If whitespace is what causes the line to break, don't print it.
				if isNextSpace {
					shouldPrintNext = false
				}
			}

			if !isNextSpace {
				// This allocation is avoidable if we track indices and assign
				// this slice within the lock, but it's not a big deal and this
				// is more readable.
				charsSinceSpace = append(charsSinceSpace, next)
			}

			if shouldPrintNext {
				fmt.Fprint(w.w, next.String())
				lineLength++
			}

			// Sleep between each character, speeding up output if there's a lot remaining.
			ms := 5 + 35*math.Exp(-0.005*float64(remaining))
			time.Sleep(time.Duration(math.Max(ms, 5)) * time.Millisecond)
		}
	}()
	w.wg.Wait()
	fmt.Fprint(w.w, resetColor)
	fmt.Fprint(w.w, showCursor)
	fmt.Fprintln(w.w)
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.done {
		return 0, io.EOF
	}
	for _, r := range string(p) {
		// TODO: Set flags on each character.
		w.stream = append(w.stream, char{value: r})
	}
	w.cond.Broadcast()
	return len(p), nil
}

func (w *writer) Done() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	w.cond.Broadcast()
}

func (w *writer) SetTask(label string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.done {
		panic("Cannot set task after the writer is done")
	}
	w.taskLabel = serif.Format(label)
	w.taskIndex = len(w.stream)
	w.cond.Broadcast()
}

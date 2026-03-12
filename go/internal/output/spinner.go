package output

import (
	"fmt"
	"os"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner displays an animated spinner with a message on stderr.
// It only renders when stdout is a TTY; in CI/pipes it's a no-op.
//
// Usage:
//
//	s := output.NewSpinner("Waiting for dataset discovery")
//	s.Start()
//	// ... do work, optionally call s.SetMessage("Still waiting...")
//	s.Stop() // clears the line
type Spinner struct {
	msg  string
	stop chan struct{}
	done chan struct{}
}

func NewSpinner(msg string) *Spinner {
	return &Spinner{msg: msg, stop: make(chan struct{}), done: make(chan struct{})}
}

func (s *Spinner) SetMessage(msg string) {
	s.msg = msg
}

func (s *Spinner) Start() {
	if !IsTerminal() {
		fmt.Fprintf(os.Stderr, "  %s\n", s.msg)
		close(s.done)
		return
	}
	go func() {
		defer close(s.done)
		i := 0
		for {
			select {
			case <-s.stop:
				// Clear the spinner line
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			default:
				frame := Dim.Render(spinnerFrames[i%len(spinnerFrames)])
				fmt.Fprintf(os.Stderr, "\r  %s  %s", frame, Dim.Render(s.msg))
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	select {
	case <-s.stop:
		// already stopped
	default:
		close(s.stop)
	}
	<-s.done
}

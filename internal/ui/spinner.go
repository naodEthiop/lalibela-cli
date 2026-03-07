package ui

import (
	"fmt"
	"sync"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner renders a simple animated progress indicator in the terminal.
type Spinner struct {
	message  string
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
	mu       sync.Mutex
	running  bool
}

// NewSpinner creates a spinner with the provided initial message.
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		interval: 80 * time.Millisecond,
	}
}

// Start begins rendering the spinner until it is stopped.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.doneCh)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		frame := 0
		render := func() {
			s.mu.Lock()
			message := s.message
			s.mu.Unlock()
			fmt.Printf("\r%s %s", message, Cyan(spinnerFrames[frame%len(spinnerFrames)]))
			frame++
		}

		render()
		for {
			select {
			case <-s.stopCh:
				fmt.Print("\r\033[2K")
				return
			case <-ticker.C:
				render()
			}
		}
	}()
}

// Update changes the message displayed alongside the spinner.
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// StopSuccess stops the spinner and prints a success message.
func (s *Spinner) StopSuccess(message string) {
	s.stop()
	fmt.Println(Green("✔ " + message))
}

// StopError stops the spinner and prints an error message.
func (s *Spinner) StopError(message string) {
	s.stop()
	fmt.Println(Red("✖ " + message))
}

func (s *Spinner) stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	close(s.stopCh)
	doneCh := s.doneCh
	s.running = false
	s.mu.Unlock()

	<-doneCh
}

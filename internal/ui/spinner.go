package ui

import (
	"fmt"
	"sync"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	message  string
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
	mu       sync.Mutex
	running  bool
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		interval: 80 * time.Millisecond,
	}
}

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

func (s *Spinner) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

func (s *Spinner) StopSuccess(message string) {
	s.stop()
	fmt.Println(Green("✔ " + message))
}

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

package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func RunCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(out.String())
		if msg == "" {
			return fmt.Errorf("%s %v failed: %w", name, args, err)
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

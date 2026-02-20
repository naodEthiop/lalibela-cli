//go:build !windows

package ui

func enableVirtualTerminal() error {
	return nil
}

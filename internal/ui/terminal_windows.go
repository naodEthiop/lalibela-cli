//go:build windows

package ui

import (
	"os"

	"golang.org/x/sys/windows"
)

func enableVirtualTerminal() error {
	_ = enableStdoutVT()
	_ = enableStdinVT()
	return nil
}

func enableStdoutVT() error {
	handle := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return err
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	return windows.SetConsoleMode(handle, mode)
}

func enableStdinVT() error {
	handle := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return err
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT
	return windows.SetConsoleMode(handle, mode)
}

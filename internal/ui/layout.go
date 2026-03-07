package ui

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

const separatorLine = "────────────────────────────────────"

// Separator returns a reusable horizontal separator line for CLI output.
func Separator() string {
	return separatorLine
}

// SectionHeader formats a section title for CLI output.
func SectionHeader(title string) string {
	return Style(title, "1", "36")
}

// TerminalWidth returns the current terminal width, falling back to 80 columns.
func TerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

// Center returns text padded with leading spaces so it is approximately centered
// within width columns.
func Center(text string, width int) string {
	textWidth := utf8.RuneCountInString(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

// ClearScreen clears the terminal and moves the cursor to the top-left.
func ClearScreen() {
	fmt.Print("\033[H\033[2J\033[3J")
}

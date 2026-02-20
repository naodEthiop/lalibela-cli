package ui

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

const separatorLine = "────────────────────────────────────"

func Separator() string {
	return separatorLine
}

func SectionHeader(title string) string {
	return Style(title, "1", "36")
}

func TerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func Center(text string, width int) string {
	textWidth := utf8.RuneCountInString(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J\033[3J")
}

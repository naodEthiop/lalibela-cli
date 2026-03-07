package ui

import "strings"

const reset = "\033[0m"

// Green styles text using a green foreground ANSI escape code.
func Green(text string) string {
	return style(text, "32")
}

// Red styles text using a red foreground ANSI escape code.
func Red(text string) string {
	return style(text, "31")
}

// Yellow styles text using a yellow foreground ANSI escape code.
func Yellow(text string) string {
	return style(text, "33")
}

// Cyan styles text using a cyan foreground ANSI escape code.
func Cyan(text string) string {
	return style(text, "36")
}

// Bold styles text using a bold ANSI escape code.
func Bold(text string) string {
	return style(text, "1")
}

// Dim styles text using a dim ANSI escape code.
func Dim(text string) string {
	return style(text, "2")
}

// BgBlue styles text using a blue background ANSI escape code.
func BgBlue(text string) string {
	return style(text, "44")
}

// Style applies raw ANSI style codes to text.
func Style(text string, codes ...string) string {
	if len(codes) == 0 {
		return text
	}
	return style(text, strings.Join(codes, ";"))
}

func style(text string, code string) string {
	return "\033[" + code + "m" + text + reset
}

package ui

import "strings"

const reset = "\033[0m"

func Green(text string) string {
	return style(text, "32")
}

func Red(text string) string {
	return style(text, "31")
}

func Yellow(text string) string {
	return style(text, "33")
}

func Cyan(text string) string {
	return style(text, "36")
}

func Bold(text string) string {
	return style(text, "1")
}

func Dim(text string) string {
	return style(text, "2")
}

func BgBlue(text string) string {
	return style(text, "44")
}

func Style(text string, codes ...string) string {
	if len(codes) == 0 {
		return text
	}
	return style(text, strings.Join(codes, ";"))
}

func style(text string, code string) string {
	return "\033[" + code + "m" + text + reset
}

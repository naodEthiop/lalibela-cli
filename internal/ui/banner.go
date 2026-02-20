package ui

import (
	"fmt"
	"time"
)

var bannerLines = []string{
	" _          _ _ _          _           ",
	"| |    __ _| (_) |__   ___| | __ _     ",
	"| |   / _` | | | '_ \\ / _ \\ |/ _` |    ",
	"| |__| (_| | | | |_) |  __/ | (_| |    ",
	"|_____\\__,_|_|_|_.__/ \\___|_|\\__,_|    ",
}

var bannerGradient = []string{
	"38;5;45",
	"38;5;51",
	"38;5;87",
	"38;5;123",
	"38;5;159",
}

func RenderBanner() {
	width := TerminalWidth()
	for i, line := range bannerLines {
		padded := Center(line, width)
		fmt.Println(Style(padded, "1", bannerGradient[i%len(bannerGradient)]))
		time.Sleep(45 * time.Millisecond)
	}
	fmt.Println(Dim(Center("Backend Project Scaffolder for Go", width)))
	fmt.Println(Dim(Center("NaodEthiop | Software Engineer ", width)))
	fmt.Println()
}

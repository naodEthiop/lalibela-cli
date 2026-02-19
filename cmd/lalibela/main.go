package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/naodEthiop/lalibela/internal/cli"
	"github.com/naodEthiop/lalibela/internal/generator"
	"golang.org/x/term"
)

var (
	Version   = "0.1.0"
	GitCommit = "dev"
	BuildDate = "unknown"
)

var colors = []color.Attribute{color.FgRed, color.FgYellow, color.FgGreen, color.FgCyan, color.FgMagenta}

var frameworkIcons = map[string]string{
	generator.FrameworkGin:     "[GIN]",
	generator.FrameworkEcho:    "[ECHO]",
	generator.FrameworkFiber:   "[FIBER]",
	generator.FrameworkNetHTTP: "[NETHTTP]",
}

func main() {
	color.NoColor = false
	color.Output = colorable.NewColorableStdout()

	opts, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Println(color.RedString("Error: %v", err))
		os.Exit(1)
	}

	if opts.ShowVersion {
		printVersion()
		return
	}
	if opts.ShowTemplateList {
		printTemplateList()
		return
	}

	printBanner()

	projectName := opts.ProjectName
	framework := opts.Framework
	selectedFeatures := opts.Features

	if !opts.FastMode && strings.TrimSpace(projectName) == "" {
		input, err := promptTextInput("Enter project name: ")
		if err != nil {
			fmt.Println(color.RedString("Error: %v", err))
			os.Exit(1)
		}
		projectName = input
	}

	if !opts.FastMode && strings.TrimSpace(framework) == "" {
		frameworks := generator.Frameworks()
		selection, err := promptFrameworkSelection(frameworks)
		if err != nil {
			fmt.Println(color.RedString("Error: %v", err))
			os.Exit(1)
		}
		framework = selection
	}

	if !opts.FastMode && !opts.FeaturesProvided {
		features := generator.InteractiveFeatures()
		selection, err := promptFeatureSelection(features)
		if err != nil {
			fmt.Println(color.RedString("Error: %v", err))
			os.Exit(1)
		}
		normalizedFeatures, err := generator.NormalizeFeatureNames(selection)
		if err != nil {
			fmt.Println(color.RedString("Error: %v", err))
			os.Exit(1)
		}
		selectedFeatures = normalizedFeatures
	}

	if opts.FastMode {
		fmt.Println(color.GreenString("Fast mode enabled with defaults"))
	}

	if err := generator.GenerateProject(generator.Options{
		ProjectName: projectName,
		Framework:   framework,
		Features:    selectedFeatures,
		Status:      printProgress,
	}); err != nil {
		fmt.Println(color.RedString("Error: %v", err))
		os.Exit(1)
	}

	printCompletionAnimation()
	printCompletionBox(projectName)
}

func printBanner() {
	fmt.Println(color.New(color.FgHiBlue).Sprint("LALIBELA - Go Backend Scaffold CLI"))
	fmt.Println(color.New(color.FgHiBlack).Sprint("Interactive, keyboard-first scaffolding"))
	fmt.Println()
}

func printProgress(step string, current, total int) {
	icon := color.New(colors[current%len(colors)]).Sprint(">")
	barWidth := 22
	filled := current * barWidth / total
	bar := strings.Repeat("#", filled) + strings.Repeat("-", barWidth-filled)
	fmt.Printf("%s [%s] %d/%d %s\n", icon, color.New(color.FgHiCyan).Sprint(bar), current, total, step)
	time.Sleep(80 * time.Millisecond)
}

func printTemplateList() {
	catalog := generator.TemplateCatalog()
	fmt.Println(color.New(color.FgHiBlue).Sprint("Template Catalog"))
	fmt.Println("TEMPLATE | FRAMEWORKS | FEATURES")
	for _, entry := range catalog {
		fmt.Printf("%s | %s | %s\n", entry.TemplatePath, strings.Join(entry.Frameworks, ","), strings.Join(entry.Features, ","))
	}
}

func printVersion() {
	fmt.Printf("lalibela %s\n", Version)
	fmt.Printf("build date: %s\n", BuildDate)
	fmt.Printf("commit: %s\n", GitCommit)
}

func printCompletionAnimation() {
	for i := 0; i < 3; i++ {
		icon := color.New(colors[i%len(colors)]).Sprint("*")
		fmt.Printf("%s scaffold complete\n", icon)
		time.Sleep(70 * time.Millisecond)
	}
	fmt.Println()
}

func printCompletionBox(projectName string) {
	fmt.Println(color.GreenString("==========================================="))
	fmt.Println(color.GreenString("Project '%s' generated successfully.", projectName))
	fmt.Println(color.GreenString("Next steps:"))
	fmt.Println(color.GreenString("  cd %s", projectName))
	fmt.Println(color.GreenString("  go run main.go"))
	fmt.Println(color.GreenString("==========================================="))
}

func promptTextInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	out := strings.TrimSpace(input)
	if out == "" {
		return "", fmt.Errorf("input cannot be empty")
	}
	return out, nil
}

func promptFrameworkSelection(frameworks []string) (string, error) {
	if len(frameworks) == 0 {
		return "", fmt.Errorf("no frameworks available")
	}

	cursor := 0
	reader := bufio.NewReader(os.Stdin)
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = term.Restore(fd, oldState)
		fmt.Println()
	}()

	render := func() {
		fmt.Print("\033[H\033[2J")
		fmt.Println(color.New(color.FgHiBlue).Sprint("Select Framework"))
		fmt.Println(color.New(color.FgHiBlack).Sprint("Space/Enter=select  j/k=move"))
		fmt.Println()

		for i, framework := range frameworks {
			pointer := "  "
			if i == cursor {
				pointer = color.New(color.FgHiYellow).Sprint(">")
			}
			label := color.New(color.FgHiCyan).Sprint(strings.ToLower(framework))
			if icon, ok := frameworkIcons[framework]; ok {
				label = fmt.Sprintf("%s %s", icon, label)
			}
			fmt.Printf("%s %s\n", pointer, label)
		}
	}

	render()
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}

		switch b {
		case ' ', '\r', '\n':
			return frameworks[cursor], nil
		case 'j', 'J':
			if cursor < len(frameworks)-1 {
				cursor++
			}
		case 'k', 'K':
			if cursor > 0 {
				cursor--
			}
		case 0x1b:
			next, _ := reader.ReadByte()
			if next == '[' {
				dir, _ := reader.ReadByte()
				if dir == 'A' && cursor > 0 {
					cursor--
				}
				if dir == 'B' && cursor < len(frameworks)-1 {
					cursor++
				}
			}
		case 0x00, 0xE0:
			dir, _ := reader.ReadByte()
			if dir == 72 && cursor > 0 {
				cursor--
			}
			if dir == 80 && cursor < len(frameworks)-1 {
				cursor++
			}
		}
		render()
	}
}

func promptFeatureSelection(features []string) ([]string, error) {
	if len(features) == 0 {
		return nil, nil
	}

	selected := make([]bool, len(features))
	cursor := 0
	reader := bufio.NewReader(os.Stdin)
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = term.Restore(fd, oldState)
		fmt.Println()
	}()

	render := func() {
		fmt.Print("\033[H\033[2J")
		fmt.Println(color.New(color.FgHiBlue).Sprint("Optional Features"))
		fmt.Println(color.New(color.FgHiBlack).Sprint("Space=toggle+next  Enter=finish  j/k=move  a=all  n=none"))
		fmt.Println()

		for i, feature := range features {
			pointer := "  "
			if i == cursor {
				pointer = color.New(color.FgHiYellow).Sprint(">")
			}
			check := color.New(color.FgHiBlack).Sprint("[ ]")
			if selected[i] {
				check = color.New(color.FgHiGreen).Sprint("[x]")
			}
			fmt.Printf("%s %s %s\n", pointer, check, color.New(color.FgHiCyan).Sprint(feature))
		}
	}

	result := func() []string {
		out := make([]string, 0, len(features))
		for i, feature := range features {
			if selected[i] {
				out = append(out, feature)
			}
		}
		return out
	}

	render()
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}

		switch b {
		case ' ':
			selected[cursor] = !selected[cursor]
			if cursor == len(features)-1 {
				return result(), nil
			}
			cursor++
		case '\r', '\n':
			return result(), nil
		case 'j', 'J':
			if cursor < len(features)-1 {
				cursor++
			}
		case 'k', 'K':
			if cursor > 0 {
				cursor--
			}
		case 'a', 'A':
			for i := range selected {
				selected[i] = true
			}
			return result(), nil
		case 'n', 'N':
			for i := range selected {
				selected[i] = false
			}
			return result(), nil
		case 0x1b:
			next, _ := reader.ReadByte()
			if next == '[' {
				dir, _ := reader.ReadByte()
				if dir == 'A' && cursor > 0 {
					cursor--
				}
				if dir == 'B' && cursor < len(features)-1 {
					cursor++
				}
			}
		case 0x00, 0xE0:
			dir, _ := reader.ReadByte()
			if dir == 72 && cursor > 0 {
				cursor--
			}
			if dir == 80 && cursor < len(features)-1 {
				cursor++
			}
		}

		render()
	}
}

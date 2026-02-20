package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/naodEthiop/lalibela-cli/internal/cli"
	"github.com/naodEthiop/lalibela-cli/internal/generator"
	"github.com/naodEthiop/lalibela-cli/internal/ui"
	"golang.org/x/term"
)

var (
	Version   = "0.1.8"
	GitCommit = "dev"
	BuildDate = "unknown"
)

var errGenerationCancelled = errors.New("generation cancelled")

func main() {
	ui.InitTerminal()

	opts, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Println(ui.Red(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}

	if opts.ShowHelp {
		printHelp()
		return
	}
	if opts.ShowVersion {
		printVersion()
		return
	}
	if opts.ShowTemplateList {
		printTemplateList()
		return
	}

	ui.RenderBanner()

	projectName := opts.ProjectName
	framework := opts.Framework
	selectedFeatures := opts.Features

	if !opts.FastMode && strings.TrimSpace(projectName) == "" {
		input, err := promptTextInput("Enter project name: ")
		if err != nil {
			fmt.Println(ui.Red(fmt.Sprintf("Error: %v", err)))
			os.Exit(1)
		}
		projectName = input
	}

	if !opts.FastMode && strings.TrimSpace(framework) == "" {
		frameworks := generator.Frameworks()
		selection, err := promptFrameworkSelection(frameworks)
		if err != nil {
			fmt.Println(ui.Red(fmt.Sprintf("Error: %v", err)))
			os.Exit(1)
		}
		framework = selection
	}

	if !opts.FastMode && !opts.FeaturesProvided {
		features := generator.InteractiveFeatures()
		selection, err := promptFeatureSelection(features)
		if err != nil {
			fmt.Println(ui.Red(fmt.Sprintf("Error: %v", err)))
			os.Exit(1)
		}
		normalizedFeatures, err := generator.NormalizeFeatureNames(selection)
		if err != nil {
			fmt.Println(ui.Red(fmt.Sprintf("Error: %v", err)))
			os.Exit(1)
		}
		selectedFeatures = normalizedFeatures
	}

	if opts.FastMode {
		printFastModeSummary(projectName, framework, selectedFeatures)
	}

	if err := confirmOverwriteIfNeeded(projectName); err != nil {
		if errors.Is(err, errGenerationCancelled) {
			fmt.Println(ui.Yellow("Generation cancelled."))
			return
		}
		fmt.Println(ui.Red(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}

	fmt.Println(ui.Separator())
	fmt.Println(ui.SectionHeader("Generation"))
	fmt.Println(ui.Separator())

	spinner := ui.NewSpinner("Generating project...")
	spinner.Start()
	if err := generator.GenerateProject(generator.Options{
		ProjectName: projectName,
		Framework:   framework,
		Features:    selectedFeatures,
	}); err != nil {
		spinner.StopError("Generation failed")
		fmt.Println(ui.Red(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}

	spinner.StopSuccess("Project created successfully!")
	printCompletionBox(projectName)
}

func printTemplateList() {
	catalog := generator.TemplateCatalog()
	fmt.Println(ui.SectionHeader("Template Catalog"))
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

func printHelp() {
	fmt.Println("Lalibela CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  lalibela [flags]")
	fmt.Println("  lalibela help")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -h, --h, --help          Show help and exit")
	fmt.Println("  -fast                    Fast mode: skip prompts and use defaults")
	fmt.Println("  -name string             Project name")
	fmt.Println("  -framework string        Framework: gin|echo|fiber|nethttp")
	fmt.Println("  -features string         Comma-separated features (Clean,Logger,PostgreSQL,JWT,Docker)")
	fmt.Println("  -version                 Print version/build metadata and exit")
	fmt.Println("  -template-list           List all templates and feature support")
	fmt.Println("  -config string           Optional config file path (defaults to ~/.lalibela.json)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  lalibela")
	fmt.Println("  lalibela -fast")
	fmt.Println("  lalibela -name myapi -framework gin -features \"Clean,Logger,JWT\"")
	fmt.Println("  lalibela -version")
}

func printCompletionBox(projectName string) {
	fmt.Println(ui.Separator())
	fmt.Println(ui.Green(fmt.Sprintf("Project '%s' generated successfully.", projectName)))
	fmt.Println(ui.SectionHeader("Next steps"))
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go run .")
	fmt.Println(ui.Separator())
}

func printFastModeSummary(projectName, framework string, features []string) {
	fmt.Println(ui.Separator())
	fmt.Printf("Project:     %s\n", projectName)
	fmt.Printf("Framework:   %s %s\n", ui.FrameworkIcon(framework), ui.FrameworkLabel(framework))
	fmt.Println("Features:")
	for _, feature := range features {
		fmt.Printf("  %s %s\n", ui.Green("✓"), feature)
	}
	fmt.Println(ui.Separator())
	fmt.Println()
}

func confirmOverwriteIfNeeded(projectName string) error {
	info, err := os.Stat(projectName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("path %q already exists and is not a directory", projectName)
	}

	fmt.Println(ui.Yellow("⚠ Directory already exists."))
	fmt.Print("Overwrite? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	answer := strings.TrimSpace(strings.ToLower(input))
	if answer != "y" && answer != "yes" {
		return errGenerationCancelled
	}

	return os.RemoveAll(projectName)
}

func promptTextInput(prompt string) (string, error) {
	fmt.Print(ui.Cyan(prompt))
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
		ui.ClearScreen()
		fmt.Println(ui.Separator())
		fmt.Println(ui.SectionHeader("Framework Selection"))
		fmt.Println(ui.Dim("Space/Enter=select  ↑/↓=move"))
		fmt.Println(ui.Separator())
		fmt.Println()

		for i, framework := range frameworks {
			row := fmt.Sprintf("%s %-8s - %s", ui.FrameworkIcon(framework), ui.FrameworkLabel(framework), ui.FrameworkDescription(framework))
			if i == cursor {
				fmt.Printf("  %s %s\n", ui.Cyan(ui.Bold("➜")), ui.Style(" "+row+" ", "1", "36", "44"))
				continue
			}
			fmt.Printf("    %s\n", ui.Dim(row))
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
		default:
			handleArrowNavigation(b, reader, &cursor, len(frameworks))
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
		ui.ClearScreen()
		fmt.Println(ui.Separator())
		fmt.Println(ui.SectionHeader("Feature Selection"))
		fmt.Println(ui.Dim("Space=toggle+next  Enter=finish  ↑/↓=move  a=all  n=none"))
		fmt.Println(ui.Separator())
		fmt.Println()

		for i, feature := range features {
			check := ui.Dim("[ ]")
			if selected[i] {
				check = ui.Green("[✓]")
			}
			if i == cursor {
				fmt.Printf("  %s %s %s\n", ui.Cyan(ui.Bold("➜")), check, ui.Style(feature, "1", "36"))
				continue
			}
			fmt.Printf("    %s %s\n", check, ui.Dim(feature))
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
		default:
			handleArrowNavigation(b, reader, &cursor, len(features))
		}

		render()
	}
}

func handleArrowNavigation(first byte, reader *bufio.Reader, cursor *int, total int) {
	switch first {
	case 0x00, 0xE0:
		dir, err := reader.ReadByte()
		if err != nil {
			return
		}
		if dir == 72 && *cursor > 0 {
			*cursor--
		}
		if dir == 80 && *cursor < total-1 {
			*cursor++
		}
	case 0x1b:
		next, err := reader.ReadByte()
		if err != nil {
			return
		}
		if next != '[' && next != 'O' {
			return
		}

		for {
			dir, err := reader.ReadByte()
			if err != nil {
				return
			}
			if dir == 'A' {
				if *cursor > 0 {
					*cursor--
				}
				return
			}
			if dir == 'B' {
				if *cursor < total-1 {
					*cursor++
				}
				return
			}
			if dir >= '@' && dir <= '~' {
				return
			}
		}
	case '[', 'O':
		for {
			dir, err := reader.ReadByte()
			if err != nil {
				return
			}
			if dir == 'A' {
				if *cursor > 0 {
					*cursor--
				}
				return
			}
			if dir == 'B' {
				if *cursor < total-1 {
					*cursor++
				}
				return
			}
			if dir >= '@' && dir <= '~' {
				return
			}
		}
	}
}

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/naodEthiop/lalibela-cli/internal/cli"
	"github.com/naodEthiop/lalibela-cli/internal/features"
	"github.com/naodEthiop/lalibela-cli/internal/generator"
	"github.com/naodEthiop/lalibela-cli/internal/ui"
	"github.com/naodEthiop/lalibela-cli/internal/utils"
	"golang.org/x/term"
)

var (
	Version   = "0.1.9"
	GitCommit = "dev"
	BuildDate = "unknown"
)

var errGenerationCancelled = errors.New("generation cancelled")

func main() {
	ui.InitTerminal()
	hydrateBuildMetadata()

	if handleSubcommand(os.Args[1:]) {
		return
	}

	opts, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		handleRootParseError(err)
	}

	if opts.ShowHelp {
		printRootHelp()
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

	if !opts.FastMode && !opts.AssumeYes && strings.TrimSpace(projectName) == "" {
		input, err := promptTextInput("Enter project name: ")
		if err != nil {
			exitWithError(
				"Could not read project name.",
				fmt.Sprintf("Details: %v", err),
				"Use -name <project> or pass --yes to run non-interactively.",
			)
		}
		projectName = input
	}

	if !opts.FastMode && !opts.AssumeYes && strings.TrimSpace(framework) == "" {
		frameworks := generator.Frameworks()
		selection, err := promptFrameworkSelection(frameworks)
		if err != nil {
			exitWithError(
				"Could not read framework selection.",
				fmt.Sprintf("Details: %v", err),
				"Use -framework <gin|echo|fiber|nethttp> or pass --yes.",
			)
		}
		framework = selection
	}

	if !opts.FastMode && !opts.AssumeYes && !opts.FeaturesProvided {
		features := generator.InteractiveFeatures()
		selection, err := promptFeatureSelection(features)
		if err != nil {
			exitWithError(
				"Could not read feature selection.",
				fmt.Sprintf("Details: %v", err),
				"Use -features \"Clean,Logger,JWT\" or pass --yes.",
			)
		}
		normalizedFeatures, err := generator.NormalizeFeatureNames(selection)
		if err != nil {
			exitWithError(
				"Feature selection was invalid.",
				fmt.Sprintf("Details: %v", err),
				"Run 'lalibela --help' for supported feature values.",
			)
		}
		selectedFeatures = normalizedFeatures
	}

	if strings.TrimSpace(projectName) == "" {
		exitWithError(
			"Project name is required.",
			"Use -name <project-name>.",
			"Run 'lalibela --help' for examples.",
		)
	}
	if strings.TrimSpace(framework) == "" {
		exitWithError(
			"Framework is required.",
			"Use -framework <gin|echo|fiber|nethttp>.",
			"Run 'lalibela --help' for examples.",
		)
	}

	if opts.FastMode || opts.AssumeYes {
		printFastModeSummary(projectName, framework, selectedFeatures)
	}

	if err := confirmOverwriteIfNeeded(projectName, opts.AssumeYes); err != nil {
		if errors.Is(err, errGenerationCancelled) {
			fmt.Println(ui.Yellow("Generation cancelled."))
			return
		}
		exitWithError(
			"Unable to prepare target directory.",
			fmt.Sprintf("Details: %v", err),
			"Choose a different -name or remove the existing directory.",
		)
	}

	fmt.Println(ui.Separator())
	fmt.Println(ui.SectionHeader("Generation"))
	fmt.Println(ui.Separator())

	stepLogs := make([]string, 0, 16)
	spinner := ui.NewSpinner("Initializing scaffold...")
	spinner.Start()
	if err := generator.GenerateProject(generator.Options{
		ProjectName: projectName,
		Framework:   framework,
		Features:    selectedFeatures,
		Status: func(step string, current int, total int) {
			stepLogs = append(stepLogs, step)
			spinner.Update(fmt.Sprintf("(%d/%d) %s", current, total, formatGenerationStep(step)))
		},
	}); err != nil {
		spinner.StopError("Scaffold failed")
		printGenerationFailureAndExit(err)
	}

	spinner.StopSuccess("Project files generated")
	printCompletedSteps(stepLogs)
	printCompletionBox(projectName, selectedFeatures)
}

func handleSubcommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "add":
		runAddCommand(args[1:])
		return true
	case "help":
		runHelpCommand(args[1:])
		return true
	case "run":
		runRunCommand(args[1:])
		return true
	case "uninstall":
		runUninstallCommand(args[1:])
		return true
	case "-h", "--help":
		printRootHelp()
		return true
	case "-v", "--version":
		printVersion()
		return true
	default:
		if strings.HasPrefix(strings.TrimSpace(args[0]), "-") {
			return false
		}
		exitWithError(
			fmt.Sprintf("Unknown command %q.", args[0]),
			"Run 'lalibela help' to see available commands.",
		)
		return false
	}
}

func runAddCommand(args []string) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	showHelp := fs.Bool("help", false, "Show add command help")
	showHelpShort := fs.Bool("h", false, "Show add command help")
	if err := fs.Parse(args); err != nil {
		exitWithError(
			"Invalid arguments for 'add' command.",
			fmt.Sprintf("Details: %v", err),
			"Run 'lalibela help add' for usage.",
		)
	}
	if *showHelp || *showHelpShort {
		printAddHelp()
		return
	}
	if fs.NArg() != 1 {
		exitWithError(
			"Feature name is required.",
			"Usage: lalibela add <feature>",
			fmt.Sprintf("Supported features: %s", strings.Join(features.KnownFeatures(), ", ")),
		)
	}
	featureName := fs.Arg(0)

	projectRoot, err := os.Getwd()
	if err != nil {
		exitWithError(
			"Could not determine current directory.",
			fmt.Sprintf("Details: %v", err),
		)
	}

	framework, err := features.DetectFramework(projectRoot)
	if err != nil {
		exitWithError(
			"Could not detect project framework.",
			fmt.Sprintf("Details: %v", err),
			"Run this command from a generated Lalibela project root.",
		)
	}

	spinner := ui.NewSpinner("Installing feature...")
	spinner.Start()
	result, err := features.InstallFeature(projectRoot, framework, featureName, utils.RunCommand)
	if err != nil {
		spinner.StopError("Feature install failed")
		exitWithError(
			fmt.Sprintf("Failed to install feature %q.", featureName),
			fmt.Sprintf("Details: %v", err),
			"Run 'lalibela help add' for supported feature names.",
		)
	}
	if !result.Compatible {
		spinner.StopError("Feature not compatible")
		fmt.Printf("⚠ Feature '%s' not supported for %s\n", result.Name, framework)
		return
	}
	if result.AlreadyPresent {
		spinner.StopSuccess("No changes needed")
		fmt.Printf("Feature '%s' is already installed.\n", result.Name)
		return
	}
	spinner.StopSuccess("Feature installed")
	fmt.Printf("Feature '%s' installed successfully.\n", result.Name)
	fmt.Println("Next:")
	fmt.Println("  go test ./...")
}

func runRunCommand(args []string) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	open := fs.Bool("open", false, "Open browser after server starts")
	showHelp := fs.Bool("help", false, "Show run command help")
	showHelpShort := fs.Bool("h", false, "Show run command help")
	if err := fs.Parse(args); err != nil {
		exitWithError(
			"Invalid arguments for 'run' command.",
			fmt.Sprintf("Details: %v", err),
			"Run 'lalibela help run' for usage.",
		)
	}
	if *showHelp || *showHelpShort {
		printRunHelp()
		return
	}
	if fs.NArg() > 0 {
		exitWithError(
			fmt.Sprintf("Unexpected argument %q for 'run' command.", fs.Arg(0)),
			"Usage: lalibela run [--open]",
		)
	}
	if _, err := os.Stat("go.mod"); err != nil {
		exitWithError(
			"Could not find go.mod in the current directory.",
			"Run this command inside your generated project directory.",
			"Example: cd myapi && lalibela run",
		)
	}

	cmdArgs := []string{"run", "."}
	if *open {
		cmdArgs = append(cmdArgs, "--open")
	}

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		exitWithError(
			"Project run failed.",
			fmt.Sprintf("Details: %v", err),
			"If port 8080 is busy, set PORT environment variable (for example: PORT=8081).",
		)
	}
}

func runUninstallCommand(args []string) {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	force := fs.Bool("force", false, "Uninstall without confirmation prompt")
	showHelp := fs.Bool("help", false, "Show uninstall command help")
	showHelpShort := fs.Bool("h", false, "Show uninstall command help")
	if err := fs.Parse(args); err != nil {
		exitWithError(
			"Invalid arguments for 'uninstall' command.",
			fmt.Sprintf("Details: %v", err),
			"Run 'lalibela help uninstall' for usage.",
		)
	}
	if *showHelp || *showHelpShort {
		printUninstallHelp()
		return
	}
	if fs.NArg() > 0 {
		exitWithError(
			fmt.Sprintf("Unexpected argument %q for 'uninstall' command.", fs.Arg(0)),
			"Usage: lalibela uninstall [--force]",
		)
	}

	executablePath, err := os.Executable()
	if err != nil {
		exitWithError(
			"Could not determine Lalibela executable path.",
			fmt.Sprintf("Details: %v", err),
		)
	}
	executablePath = strings.TrimSpace(executablePath)
	if executablePath == "" {
		exitWithError("Could not determine Lalibela executable path.")
	}

	if _, err := os.Stat(executablePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Binary not found.")
			os.Exit(1)
		}
		exitWithError(
			"Unable to access Lalibela binary.",
			fmt.Sprintf("Details: %v", err),
		)
	}

	if !*force {
		fmt.Printf("Uninstall Lalibela CLI from %s? (y/N): ", executablePath)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			exitWithError(
				"Could not read uninstall confirmation.",
				fmt.Sprintf("Details: %v", err),
			)
		}
		answer := strings.ToLower(strings.TrimSpace(input))
		if answer != "y" && answer != "yes" {
			fmt.Println("Uninstall cancelled.")
			return
		}
	}

	if err := os.Remove(executablePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Binary not found.")
			os.Exit(1)
		}
		if isPermissionDenied(err) {
			fmt.Println("Permission denied. Try running with sudo or administrator privileges.")
			os.Exit(1)
		}
		exitWithError(
			"Failed to uninstall Lalibela CLI.",
			fmt.Sprintf("Details: %v", err),
		)
	}

	fmt.Println("Lalibela CLI uninstalled successfully.")
}

func runHelpCommand(args []string) {
	if len(args) == 0 {
		printRootHelp()
		return
	}
	if len(args) > 1 {
		exitWithError(
			"Too many arguments for help command.",
			"Usage: lalibela help [command]",
			"Supported commands: add, run, uninstall",
		)
	}

	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "add":
		printAddHelp()
	case "run":
		printRunHelp()
	case "uninstall":
		printUninstallHelp()
	case "help":
		printRootHelp()
	default:
		exitWithError(
			fmt.Sprintf("Unknown help topic %q.", args[0]),
			"Supported help topics: add, run, uninstall",
		)
	}
}

func printTemplateList() {
	catalog := generator.TemplateCatalog()
	fmt.Println(ui.SectionHeader("Template Catalog"))
	fmt.Println("TEMPLATE | FRAMEWORKS | FEATURES")
	for _, entry := range catalog {
		fmt.Printf("%s | %s | %s\n", entry.TemplatePath, strings.Join(entry.Frameworks, ","), strings.Join(entry.Features, ","))
	}
}

func hydrateBuildMetadata() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		hydrateBuildDateFromExecutable()
		return
	}

	if (Version == "" || Version == "dev" || Version == "0.0.0") && info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = strings.TrimSpace(info.Main.Version)
	}
	if (GitCommit == "" || GitCommit == "dev") && buildSetting(info, "vcs.revision") != "" {
		GitCommit = buildSetting(info, "vcs.revision")
	}
	if (BuildDate == "" || BuildDate == "unknown") && buildSetting(info, "vcs.time") != "" {
		BuildDate = buildSetting(info, "vcs.time")
	}
	hydrateBuildDateFromExecutable()
}

func buildSetting(info *debug.BuildInfo, key string) string {
	for _, setting := range info.Settings {
		if setting.Key == key {
			return strings.TrimSpace(setting.Value)
		}
	}
	return ""
}

func hydrateBuildDateFromExecutable() {
	if BuildDate != "" && BuildDate != "unknown" {
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		return
	}
	stat, err := os.Stat(exePath)
	if err != nil {
		return
	}
	BuildDate = stat.ModTime().UTC().Format("2006-01-02T15:04:05Z")
}

func printVersion() {
	fmt.Printf("%s %s\n", ui.Bold(ui.Cyan("lalibela")), ui.Bold(Version))
	fmt.Printf("%s %s\n", ui.Dim("build date:"), ui.Yellow(BuildDate))
	fmt.Printf("%s %s\n", ui.Dim("commit:"), ui.Yellow(GitCommit))
}

func printRootHelp() {
	fmt.Println(ui.Bold(ui.Cyan("Lalibela CLI")))
	fmt.Println()
	fmt.Println(ui.SectionHeader("Usage"))
	fmt.Println("  lalibela [flags]")
	fmt.Println("  lalibela add <feature> [flags]")
	fmt.Println("  lalibela run [flags]")
	fmt.Println("  lalibela uninstall [flags]")
	fmt.Println("  lalibela help [command]")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Flags"))
	fmt.Println("  -h, --help               Show help")
	fmt.Println("  -v, --version, -version  Show version/build metadata")
	fmt.Println("  -fast                    Skip prompts and use defaults")
	fmt.Println("  -y, --yes                Auto-accept prompts and use defaults for missing values")
	fmt.Println("  -name string             Project name")
	fmt.Println("  -framework string        Framework: gin|echo|fiber|nethttp")
	fmt.Println("  -features string         Comma-separated features (Clean,Logger,PostgreSQL,JWT,Docker)")
	fmt.Println("  -template-list           List scaffold templates and support")
	fmt.Println("  -config string           Optional config path (default: ~/.lalibela.json)")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Feature modules"))
	fmt.Printf("  %s\n", strings.Join(features.KnownFeatures(), ", "))
	fmt.Println()
	fmt.Println(ui.SectionHeader("Examples"))
	fmt.Println("  lalibela")
	fmt.Println("  lalibela -fast")
	fmt.Println("  lalibela --yes")
	fmt.Println("  lalibela -name myapi -framework gin -features \"Clean,Logger,JWT\"")
	fmt.Println("  lalibela add postgres")
	fmt.Println("  lalibela run --open")
	fmt.Println("  lalibela uninstall --force")
	fmt.Println("  lalibela help add")
}

func printAddHelp() {
	fmt.Println(ui.Bold(ui.Cyan("Lalibela add")))
	fmt.Println()
	fmt.Println(ui.SectionHeader("Usage"))
	fmt.Println("  lalibela add <feature>")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Description"))
	fmt.Println("  Installs a production feature into the current Lalibela project.")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Flags"))
	fmt.Println("  -h, --help  Show add command help")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Supported features"))
	fmt.Printf("  %s\n", strings.Join(features.KnownFeatures(), ", "))
	fmt.Println()
	fmt.Println(ui.SectionHeader("Examples"))
	fmt.Println("  lalibela add config")
	fmt.Println("  lalibela add auth")
	fmt.Println("  lalibela add redis")
}

func printRunHelp() {
	fmt.Println(ui.Bold(ui.Cyan("Lalibela run")))
	fmt.Println()
	fmt.Println(ui.SectionHeader("Usage"))
	fmt.Println("  lalibela run [--open]")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Description"))
	fmt.Println("  Runs the generated project using 'go run .'.")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Flags"))
	fmt.Println("  --open      Open browser after server startup")
	fmt.Println("  -h, --help  Show run command help")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Examples"))
	fmt.Println("  lalibela run")
	fmt.Println("  lalibela run --open")
}

func printUninstallHelp() {
	fmt.Println(ui.Bold(ui.Cyan("Lalibela uninstall")))
	fmt.Println()
	fmt.Println(ui.SectionHeader("Usage"))
	fmt.Println("  lalibela uninstall [--force]")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Description"))
	fmt.Println("  Removes only the Lalibela CLI binary from the current executable path.")
	fmt.Println("  User config files and generated projects are not removed.")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Flags"))
	fmt.Println("  --force     Skip uninstall confirmation prompt")
	fmt.Println("  -h, --help  Show uninstall command help")
	fmt.Println()
	fmt.Println(ui.SectionHeader("Examples"))
	fmt.Println("  lalibela uninstall")
	fmt.Println("  lalibela uninstall --force")
}

func printCompletionBox(projectName string, selectedFeatures []string) {
	fmt.Println(ui.Separator())
	fmt.Println(ui.Green("Project scaffolded successfully."))
	fmt.Println(ui.Separator())
	fmt.Println()
	fmt.Println("Next:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go run .")

	suggestions := featureSuggestions(selectedFeatures)
	if len(suggestions) > 0 {
		fmt.Println()
		fmt.Println("Optional setup tips:")
		for _, suggestion := range suggestions {
			fmt.Printf("  - %s\n", suggestion)
		}
	}
	fmt.Println()
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

func handleRootParseError(err error) {
	message := strings.TrimSpace(err.Error())
	switch {
	case strings.Contains(message, "unknown command or argument"):
		exitWithError(
			message,
			"Run 'lalibela help' to see available commands.",
		)
	case strings.Contains(message, "flag provided but not defined"):
		exitWithError(
			"Unknown flag.",
			fmt.Sprintf("Details: %s", message),
			"Run 'lalibela --help' to view supported flags.",
		)
	case strings.Contains(message, "unsupported framework"):
		exitWithError(
			message,
			"Supported frameworks: gin, echo, fiber, nethttp.",
		)
	case strings.Contains(message, "unsupported feature"):
		exitWithError(
			message,
			"Use -features with values like: Clean,Logger,PostgreSQL,JWT,Docker.",
		)
	default:
		exitWithError(
			"Could not parse command arguments.",
			fmt.Sprintf("Details: %s", message),
			"Run 'lalibela --help' for usage examples.",
		)
	}
}

func printGenerationFailureAndExit(err error) {
	details := strings.TrimSpace(err.Error())
	switch {
	case strings.Contains(details, "go mod init failed"):
		exitWithError(
			"Scaffold created files but module initialization failed.",
			fmt.Sprintf("Details: %s", details),
			"Check your Go toolchain installation and GOPATH/GOMOD settings.",
		)
	case strings.Contains(details, "go mod tidy failed"):
		exitWithError(
			"Dependency resolution failed while scaffolding.",
			fmt.Sprintf("Details: %s", details),
			"Check internet access and retry, or run 'go mod tidy' manually inside the project.",
		)
	case strings.Contains(details, "rollback complete"):
		exitWithError(
			"Scaffold failed and rollback completed.",
			fmt.Sprintf("Details: %s", details),
			"Fix the reported issue and run the command again.",
		)
	default:
		exitWithError(
			"Scaffolding failed.",
			fmt.Sprintf("Details: %s", details),
			"Run again with valid flags or use 'lalibela --help'.",
		)
	}
}

func exitWithError(message string, suggestions ...string) {
	fmt.Println(ui.Red("Error: " + message))
	for _, suggestion := range suggestions {
		trimmed := strings.TrimSpace(suggestion)
		if trimmed == "" {
			continue
		}
		fmt.Printf("  -> %s\n", trimmed)
	}
	os.Exit(1)
}

func formatGenerationStep(step string) string {
	trimmed := strings.TrimSpace(step)
	if trimmed == "" {
		return "Working..."
	}
	return strings.ToUpper(trimmed[:1]) + trimmed[1:]
}

func printCompletedSteps(steps []string) {
	if len(steps) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Completed:")
	for _, step := range steps {
		fmt.Printf("  %s %s\n", ui.Green("✓"), formatGenerationStep(step))
	}
}

func featureSuggestions(selectedFeatures []string) []string {
	suggestions := make([]string, 0, 5)
	if hasFeature(selectedFeatures, generator.FeatureJWT) {
		suggestions = append(suggestions, "JWT: add auth middleware to protected routes in internal/middleware/jwt.go.")
	}
	if hasFeature(selectedFeatures, generator.FeaturePostgreSQL) {
		suggestions = append(suggestions, "PostgreSQL: set DB env vars and review config/database.go.")
	}
	if hasFeature(selectedFeatures, generator.FeatureDocker) {
		suggestions = append(suggestions, "Docker: build with 'docker build -t <app> .' and run with 'docker run -p 8080:8080 <app>'.")
	}
	if hasFeature(selectedFeatures, generator.FeatureClean) {
		suggestions = append(suggestions, "Clean Architecture: start wiring use cases in internal/app/bootstrap.go.")
	}
	if hasFeature(selectedFeatures, generator.FeatureLogger) {
		suggestions = append(suggestions, "Logger: initialize logger in config/logger.go during startup.")
	}
	return suggestions
}

func hasFeature(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), target) {
			return true
		}
	}
	return false
}

func isPermissionDenied(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrPermission) {
		return true
	}

	text := strings.ToLower(err.Error())
	return strings.Contains(text, "permission denied") ||
		strings.Contains(text, "access is denied") ||
		strings.Contains(text, "operation not permitted")
}

func confirmOverwriteIfNeeded(projectName string, assumeYes bool) error {
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

	if assumeYes {
		fmt.Println(ui.Yellow("Directory already exists; overwriting because --yes was provided."))
		return os.RemoveAll(projectName)
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

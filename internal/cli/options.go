package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/naodEthiop/lalibela-cli/internal/generator"
)

type Config struct {
	ProjectName string   `json:"project_name"`
	Framework   string   `json:"framework"`
	Features    []string `json:"features"`
	Fast        bool     `json:"fast"`
}

type Options struct {
	ProjectName      string
	Framework        string
	Features         []string
	FeaturesProvided bool
	FastMode         bool
	ShowHelp         bool
	ShowVersion      bool
	ShowTemplateList bool
	ConfigPath       string
}

func ParseArgs(args []string) (Options, error) {
	var opts Options
	if len(args) > 0 && strings.EqualFold(strings.TrimSpace(args[0]), "help") {
		return Options{ShowHelp: true}, nil
	}

	fs := flag.NewFlagSet("lalibela", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fast := fs.Bool("fast", false, "Fast mode: skip prompts and use defaults")
	showHelp := fs.Bool("help", false, "Show help and exit")
	showHelpShort := fs.Bool("h", false, "Show help and exit")
	project := fs.String("name", "", "Project name")
	framework := fs.String("framework", "", "Framework: gin|echo|fiber|nethttp")
	features := fs.String("features", "", "Comma-separated features (Clean,Logger,PostgreSQL,JWT,Docker)")
	showVersion := fs.Bool("version", false, "Print version/build metadata and exit")
	templateList := fs.Bool("template-list", false, "List all templates and feature support")
	configPath := fs.String("config", "", "Optional config file path (defaults to ~/.lalibela.json)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Options{ShowHelp: true}, nil
		}
		return opts, err
	}

	opts = Options{
		ShowHelp:         *showHelp || *showHelpShort,
		ShowVersion:      *showVersion,
		ShowTemplateList: *templateList,
	}
	if opts.ShowHelp || opts.ShowVersion || opts.ShowTemplateList {
		return opts, nil
	}

	visited := make(map[string]struct{})
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = struct{}{}
	})

	resolvedConfigPath := *configPath
	if resolvedConfigPath == "" {
		resolvedConfigPath = defaultConfigPath()
	}
	cfg, err := loadConfig(resolvedConfigPath)
	if err != nil {
		return opts, err
	}

	opts = Options{
		ProjectName: strings.TrimSpace(cfg.ProjectName),
		Framework:   strings.TrimSpace(cfg.Framework),
		FastMode:    cfg.Fast,
		ConfigPath:  resolvedConfigPath,
	}

	if len(cfg.Features) > 0 {
		normalized, err := generator.NormalizeFeatureNames(cfg.Features)
		if err != nil {
			return opts, fmt.Errorf("invalid features in config %q: %w", resolvedConfigPath, err)
		}
		opts.Features = normalized
	}

	if _, ok := visited["fast"]; ok {
		opts.FastMode = *fast
	}
	if _, ok := visited["name"]; ok {
		opts.ProjectName = strings.TrimSpace(*project)
	}
	if _, ok := visited["framework"]; ok {
		opts.Framework = strings.TrimSpace(*framework)
	}
	if _, ok := visited["features"]; ok {
		normalized, err := generator.ParseFeatureCSV(*features)
		if err != nil {
			return opts, err
		}
		opts.Features = normalized
		opts.FeaturesProvided = true
	}

	if opts.Framework != "" && !generator.IsSupportedFramework(opts.Framework) {
		return opts, fmt.Errorf("unsupported framework %q", opts.Framework)
	}

	if opts.FastMode {
		if opts.ProjectName == "" {
			opts.ProjectName = "myapp"
		}
		if opts.Framework == "" {
			opts.Framework = generator.FrameworkGin
		}
		if len(opts.Features) == 0 {
			opts.Features = generator.DefaultFastFeatures()
		}
	}

	return opts, nil
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ".lalibela.json"
	}
	return filepath.Join(home, ".lalibela.json")
}

func loadConfig(path string) (Config, error) {
	var cfg Config
	if strings.TrimSpace(path) == "" {
		return cfg, nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config file %q: %w", path, err)
	}

	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config file %q: %w", path, err)
	}
	return cfg, nil
}

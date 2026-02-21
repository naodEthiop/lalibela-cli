package features

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const statePath = ".lalibela/features.json"

type State struct {
	Framework string   `json:"framework"`
	Installed []string `json:"installed"`
}

func KnownFeatures() []string {
	keys := make([]string, 0, len(Registry))
	for name := range Registry {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}

func InstallDefaults(projectRoot, framework string, runner CommandRunner) ([]InstallResult, error) {
	results := make([]InstallResult, 0, len(DefaultProductionFeatures))
	changed := false
	for _, name := range DefaultProductionFeatures {
		result, err := installFeature(projectRoot, framework, name, false)
		if err != nil {
			return nil, err
		}
		if result.Installed {
			changed = true
		}
		results = append(results, result)
	}
	if changed && runner != nil {
		if err := runner(projectRoot, "go", "mod", "tidy"); err != nil {
			return nil, fmt.Errorf("go mod tidy after default feature install: %w", err)
		}
	}
	return results, nil
}

func InstallFeature(projectRoot, framework, featureName string, runner CommandRunner) (InstallResult, error) {
	result, err := installFeature(projectRoot, framework, featureName, false)
	if err != nil {
		return result, err
	}
	if result.Installed && runner != nil {
		if err := runner(projectRoot, "go", "mod", "tidy"); err != nil {
			return result, fmt.Errorf("go mod tidy after feature install: %w", err)
		}
	}
	return result, nil
}

func installFeature(projectRoot, framework, featureName string, saveOnly bool) (InstallResult, error) {
	normalized := strings.ToLower(strings.TrimSpace(featureName))
	feature, ok := Registry[normalized]
	if !ok {
		return InstallResult{}, fmt.Errorf("unknown feature %q", featureName)
	}

	state, err := loadState(projectRoot)
	if err != nil {
		return InstallResult{}, err
	}
	if state.Framework == "" {
		state.Framework = strings.ToLower(strings.TrimSpace(framework))
	}

	result := InstallResult{
		Name:       normalized,
		Compatible: feature.Compatible(state.Framework),
	}
	if !result.Compatible {
		return result, nil
	}

	if contains(state.Installed, normalized) {
		result.AlreadyPresent = true
		return result, nil
	}

	if !saveOnly {
		if err := feature.Install(projectRoot); err != nil {
			return result, err
		}
	}
	state.Installed = append(state.Installed, normalized)
	sort.Strings(state.Installed)
	if err := saveState(projectRoot, state); err != nil {
		return result, err
	}

	result.Installed = true
	return result, nil
}

func DetectFramework(projectRoot string) (string, error) {
	if state, err := loadState(projectRoot); err == nil && strings.TrimSpace(state.Framework) != "" {
		return strings.ToLower(strings.TrimSpace(state.Framework)), nil
	}

	rawMain, err := os.ReadFile(filepath.Join(projectRoot, "main.go"))
	if err != nil {
		return "", fmt.Errorf("detecting framework: %w", err)
	}
	text := strings.ToLower(string(rawMain))
	switch {
	case strings.Contains(text, "github.com/gin-gonic/gin"):
		return "gin", nil
	case strings.Contains(text, "github.com/labstack/echo/v4"):
		return "echo", nil
	case strings.Contains(text, "github.com/gofiber/fiber/v2"):
		return "fiber", nil
	case strings.Contains(text, "net/http"):
		return "nethttp", nil
	default:
		return "", fmt.Errorf("could not detect framework from main.go")
	}
}

func loadState(projectRoot string) (State, error) {
	path := filepath.Join(projectRoot, statePath)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("reading feature state: %w", err)
	}

	var state State
	if err := json.Unmarshal(raw, &state); err != nil {
		return State{}, fmt.Errorf("parsing feature state: %w", err)
	}
	return state, nil
}

func saveState(projectRoot string, state State) error {
	path := filepath.Join(projectRoot, statePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating feature state directory: %w", err)
	}
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding feature state: %w", err)
	}
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		return fmt.Errorf("writing feature state: %w", err)
	}
	return nil
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

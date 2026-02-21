package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultModuleVersion = "v1"
	stateFileName        = "modules.json"
)

type Definition struct {
	Name    string
	Version string
}

type InstalledModule struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
}

type State struct {
	Modules map[string]InstalledModule `json:"modules"`
}

var Registry = map[string]Definition{
	"framework:gin":             {Name: "framework:gin", Version: defaultModuleVersion},
	"framework:echo":            {Name: "framework:echo", Version: defaultModuleVersion},
	"framework:fiber":           {Name: "framework:fiber", Version: defaultModuleVersion},
	"framework:nethttp":         {Name: "framework:nethttp", Version: defaultModuleVersion},
	"feature:clean":             {Name: "feature:clean", Version: defaultModuleVersion},
	"feature:logger":            {Name: "feature:logger", Version: defaultModuleVersion},
	"feature:postgresql":        {Name: "feature:postgresql", Version: defaultModuleVersion},
	"feature:jwt":               {Name: "feature:jwt", Version: defaultModuleVersion},
	"feature:docker":            {Name: "feature:docker", Version: defaultModuleVersion},
	"feature:config":            {Name: "feature:config", Version: defaultModuleVersion},
	"feature:graceful-shutdown": {Name: "feature:graceful-shutdown", Version: defaultModuleVersion},
	"feature:health":            {Name: "feature:health", Version: defaultModuleVersion},
	"feature:error-handler":     {Name: "feature:error-handler", Version: defaultModuleVersion},
	"feature:cors":              {Name: "feature:cors", Version: defaultModuleVersion},
	"feature:swagger":           {Name: "feature:swagger", Version: defaultModuleVersion},
	"feature:auth":              {Name: "feature:auth", Version: defaultModuleVersion},
	"feature:rate-limit":        {Name: "feature:rate-limit", Version: defaultModuleVersion},
	"feature:postgres":          {Name: "feature:postgres", Version: defaultModuleVersion},
	"feature:redis":             {Name: "feature:redis", Version: defaultModuleVersion},
}

func EnsureScaffoldModules(framework string, features []string) error {
	root, err := modulesRoot()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("creating modules directory: %w", err)
	}

	state, err := loadState(root)
	if err != nil {
		return err
	}

	required, err := requiredModules(framework, features)
	if err != nil {
		return err
	}

	changed := false
	for _, module := range required {
		if isModuleInstalled(state, module.Name, module.Version) {
			continue
		}
		if err := installModule(root, &state, module); err != nil {
			return err
		}
		changed = true
	}

	if changed {
		if err := saveState(root, state); err != nil {
			return err
		}
	}

	return nil
}

func isModuleInstalled(state State, name, version string) bool {
	if state.Modules == nil {
		return false
	}
	module, ok := state.Modules[name]
	if !ok {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(module.Version), strings.TrimSpace(version))
}

func installModule(root string, state *State, module Definition) error {
	if state.Modules == nil {
		state.Modules = make(map[string]InstalledModule)
	}

	path := moduleInstallPath(root, module)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("preparing module %s: %w", module.Name, err)
	}

	marker := filepath.Join(path, ".installed")
	markerBody := []byte(fmt.Sprintf("%s\n", time.Now().UTC().Format(time.RFC3339)))
	if err := os.WriteFile(marker, markerBody, 0o644); err != nil {
		return fmt.Errorf("writing module marker for %s: %w", module.Name, err)
	}

	state.Modules[module.Name] = InstalledModule{
		Name:        module.Name,
		Version:     module.Version,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}
	return nil
}

func requiredModules(framework string, features []string) ([]Definition, error) {
	keys := make([]string, 0, len(features)+1)
	keys = append(keys, fmt.Sprintf("framework:%s", strings.ToLower(strings.TrimSpace(framework))))
	for _, feature := range features {
		normalized := strings.ToLower(strings.TrimSpace(feature))
		switch normalized {
		case "clean":
			keys = append(keys, "feature:clean")
		case "logger":
			keys = append(keys, "feature:logger")
		case "postgresql":
			keys = append(keys, "feature:postgresql")
		case "jwt":
			keys = append(keys, "feature:jwt")
		case "docker":
			keys = append(keys, "feature:docker")
		}
	}

	defs := make([]Definition, 0, len(keys))
	for _, key := range keys {
		module, ok := Registry[key]
		if !ok {
			return nil, fmt.Errorf("module %q not registered", key)
		}
		defs = append(defs, module)
	}
	return defs, nil
}

func modulesRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".lalibela"), nil
}

func statePath(root string) string {
	return filepath.Join(root, stateFileName)
}

func loadState(root string) (State, error) {
	path := statePath(root)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{Modules: map[string]InstalledModule{}}, nil
		}
		return State{}, fmt.Errorf("reading module state: %w", err)
	}

	var state State
	if err := json.Unmarshal(raw, &state); err != nil {
		return State{}, fmt.Errorf("parsing module state: %w", err)
	}
	if state.Modules == nil {
		state.Modules = map[string]InstalledModule{}
	}
	return state, nil
}

func saveState(root string, state State) error {
	if state.Modules == nil {
		state.Modules = map[string]InstalledModule{}
	}

	ordered := make(map[string]InstalledModule, len(state.Modules))
	keys := make([]string, 0, len(state.Modules))
	for key := range state.Modules {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		ordered[key] = state.Modules[key]
	}

	body, err := json.MarshalIndent(State{Modules: ordered}, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding module state: %w", err)
	}

	path := statePath(root)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return fmt.Errorf("writing module state temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("finalizing module state: %w", err)
	}
	return nil
}

func moduleInstallPath(root string, module Definition) string {
	sanitized := strings.NewReplacer(":", "_", "/", "_", "\\", "_").Replace(module.Name)
	return filepath.Join(root, "modules", sanitized, module.Version)
}

package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseArgsFeatures(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "missing.json")
	opts, err := ParseArgs([]string{
		"-config", configPath,
		"-name", "demo",
		"-framework", "gin",
		"-features", "Clean,Logger,JWT",
	})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	if !opts.FeaturesProvided {
		t.Fatalf("expected FeaturesProvided=true")
	}
	want := []string{"Clean", "Logger", "JWT"}
	if !reflect.DeepEqual(opts.Features, want) {
		t.Fatalf("unexpected features:\nwant=%v\ngot=%v", want, opts.Features)
	}
}

func TestParseArgsVersionAndTemplateList(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "missing.json")
	opts, err := ParseArgs([]string{
		"-config", configPath,
		"-version",
		"-template-list",
	})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}
	if !opts.ShowVersion {
		t.Fatalf("expected ShowVersion=true")
	}
	if !opts.ShowTemplateList {
		t.Fatalf("expected ShowTemplateList=true")
	}
}

func TestParseArgsConfigFallback(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "lalibela.json")
	content := `{
  "project_name": "from-config",
  "framework": "echo",
  "features": ["Logger", "Docker"],
  "fast": true
}`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts, err := ParseArgs([]string{"-config", configPath})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}
	if opts.ProjectName != "from-config" {
		t.Fatalf("expected project from config, got %q", opts.ProjectName)
	}
	if opts.Framework != "echo" {
		t.Fatalf("expected framework echo, got %q", opts.Framework)
	}
	want := []string{"Logger", "Docker"}
	if !reflect.DeepEqual(opts.Features, want) {
		t.Fatalf("unexpected config features:\nwant=%v\ngot=%v", want, opts.Features)
	}
	if !opts.FastMode {
		t.Fatalf("expected FastMode=true from config")
	}
}

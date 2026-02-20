package generator

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "hello.tmpl")
	outputPath := filepath.Join(tempDir, "out", "hello.txt")

	if err := os.WriteFile(templatePath, []byte("hello {{ .ProjectName }}"), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	data := TemplateData{ProjectName: "lalibela"}
	if err := renderTemplate(templatePath, outputPath, data); err != nil {
		t.Fatalf("render template: %v", err)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if strings.TrimSpace(string(raw)) != "hello lalibela" {
		t.Fatalf("unexpected output: %q", string(raw))
	}
}

func TestCreateProjectDirectories(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	ctx := &generationContext{
		projectPath: filepath.Join(tempDir, "demo"),
		data: TemplateData{
			Features: FeatureSet{Clean: true},
		},
	}

	if err := createProjectDirectories(ctx); err != nil {
		t.Fatalf("create directories: %v", err)
	}

	expectedDirs := []string{
		filepath.Join(ctx.projectPath, "config"),
		filepath.Join(ctx.projectPath, "internal", "routes"),
		filepath.Join(ctx.projectPath, "internal", "middleware"),
		filepath.Join(ctx.projectPath, "internal", "domain"),
		filepath.Join(ctx.projectPath, "internal", "usecase"),
		filepath.Join(ctx.projectPath, "internal", "repository"),
		filepath.Join(ctx.projectPath, "internal", "delivery", "httptransport"),
		filepath.Join(ctx.projectPath, "internal", "app"),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected directory %s to be a directory", dir)
		}
	}
}

func TestSetupDependencies(t *testing.T) {
	t.Parallel()

	var calls []string
	runner := func(dir string, name string, args ...string) error {
		calls = append(calls, dir+"|"+name+" "+strings.Join(args, " "))
		return nil
	}

	ctx := &generationContext{
		projectPath: "demo",
		data: TemplateData{
			ModuleName: "demo",
		},
		runner: runner,
	}

	if err := setupDependencies(ctx); err != nil {
		t.Fatalf("setup dependencies: %v", err)
	}

	want := []string{
		"demo|go mod init demo",
		"demo|go mod tidy",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("unexpected command calls:\nwant=%v\ngot=%v", want, calls)
	}
}

func TestSetupDependenciesError(t *testing.T) {
	t.Parallel()

	runner := func(dir string, name string, args ...string) error {
		return errors.New("boom")
	}

	ctx := &generationContext{
		projectPath: "demo",
		data: TemplateData{
			ModuleName: "demo",
		},
		runner: runner,
	}

	err := setupDependencies(ctx)
	if err == nil {
		t.Fatal("expected error from setupDependencies")
	}
	if !strings.Contains(err.Error(), "go mod init failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateProjectUsesEmbeddedTemplates(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	projectName := "embedded-demo"
	err = GenerateProject(Options{
		ProjectName: projectName,
		Framework:   FrameworkGin,
		Features:    []string{FeatureLogger},
		Runner: func(string, string, ...string) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject should use embedded templates, got error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(tempDir, projectName, ".env"),
		filepath.Join(tempDir, projectName, "templates", "index.html"),
		filepath.Join(tempDir, projectName, "templates", "lalibela2.webp"),
		filepath.Join(tempDir, projectName, "main.go"),
		filepath.Join(tempDir, projectName, "startup.go"),
		filepath.Join(tempDir, projectName, "internal", "routes", "routes.go"),
		filepath.Join(tempDir, projectName, "internal", "routes", "welcome.go"),
		filepath.Join(tempDir, projectName, "config", "logger.go"),
	}
	for _, p := range expectedFiles {
		if _, statErr := os.Stat(p); statErr != nil {
			t.Fatalf("expected generated file %s: %v", p, statErr)
		}
	}
}

func TestGenerateProjectRollbackOnFailure(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	writeTemplate := func(relPath, content string) {
		path := filepath.Join(tempDir, relPath)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for template %s: %v", relPath, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write template %s: %v", relPath, err)
		}
	}

	writeTemplate("templates/env.tmpl", "PORT=8080")
	writeTemplate("index.html", "<h1>{{ .ProjectName }}</h1>")
	writeTemplate("lalibela2.webp", "fake-image-bytes")
	writeTemplate("templates/startup.go.tmpl", "package main")
	writeTemplate("templates/main.go.tmpl", "package main")
	writeTemplate("templates/routes/welcome.go.tmpl", "package routes")
	writeTemplate("templates/routes/gin_routes.go.tmpl", "package routes")

	projectName := "rollback-demo"
	runErr := errors.New("forced dependency error")
	err = GenerateProject(Options{
		ProjectName: projectName,
		Framework:   FrameworkGin,
		RootDir:     tempDir,
		Runner: func(string, string, ...string) error {
			return runErr
		},
	})
	if err == nil {
		t.Fatal("expected GenerateProject to fail")
	}
	if !strings.Contains(err.Error(), "rollback complete") {
		t.Fatalf("expected rollback marker in error, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(tempDir, projectName)); !os.IsNotExist(statErr) {
		t.Fatalf("expected project directory to be removed, statErr=%v", statErr)
	}
}

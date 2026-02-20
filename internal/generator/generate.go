package generator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	lalibelacli "github.com/naodEthiop/lalibela-cli"
	"github.com/naodEthiop/lalibela-cli/internal/utils"
)

const (
	FrameworkGin     = "gin"
	FrameworkEcho    = "echo"
	FrameworkFiber   = "fiber"
	FrameworkNetHTTP = "nethttp"
)

const (
	FeatureClean      = "Clean"
	FeatureLogger     = "Logger"
	FeaturePostgreSQL = "PostgreSQL"
	FeatureJWT        = "JWT"
	FeatureDocker     = "Docker"
)

var supportedFrameworks = []string{
	FrameworkGin,
	FrameworkEcho,
	FrameworkFiber,
	FrameworkNetHTTP,
}

var interactiveFeatures = []string{
	FeatureLogger,
	FeaturePostgreSQL,
	FeatureJWT,
	FeatureDocker,
	"Clean Architecture",
}

var defaultFastFeatures = []string{
	FeatureLogger,
	FeaturePostgreSQL,
	FeatureJWT,
	FeatureDocker,
}

var featureAliasToCanonical = map[string]string{
	"clean":              FeatureClean,
	"cleanarchitecture":  FeatureClean,
	"clean architecture": FeatureClean,
	"logger":             FeatureLogger,
	"postgres":           FeaturePostgreSQL,
	"postgresql":         FeaturePostgreSQL,
	"jwt":                FeatureJWT,
	"docker":             FeatureDocker,
}

type FeatureSet struct {
	Clean      bool
	Logger     bool
	PostgreSQL bool
	JWT        bool
	Docker     bool
}

func (f FeatureSet) Names() []string {
	names := make([]string, 0, 5)
	if f.Clean {
		names = append(names, FeatureClean)
	}
	if f.Logger {
		names = append(names, FeatureLogger)
	}
	if f.PostgreSQL {
		names = append(names, FeaturePostgreSQL)
	}
	if f.JWT {
		names = append(names, FeatureJWT)
	}
	if f.Docker {
		names = append(names, FeatureDocker)
	}
	return names
}

// TemplateData holds data passed to templates.
type TemplateData struct {
	ModuleName  string
	ProjectName string
	Framework   string
	Features    FeatureSet
}

type TemplateInfo struct {
	TemplatePath string
	Frameworks   []string
	Features     []string
}

type StatusFunc func(step string, current int, total int)
type CommandRunner func(dir string, name string, args ...string) error

type Options struct {
	ProjectName string
	Framework   string
	Features    []string
	RootDir     string
	TemplateFS  fs.FS
	Runner      CommandRunner
	Status      StatusFunc
}

type generationContext struct {
	templateFS  fs.FS
	projectPath string
	data        TemplateData
	runner      CommandRunner
}

type generationStep struct {
	name string
	fn   func(*generationContext) error
}

func Frameworks() []string {
	return slices.Clone(supportedFrameworks)
}

func InteractiveFeatures() []string {
	return slices.Clone(interactiveFeatures)
}

func DefaultFastFeatures() []string {
	return slices.Clone(defaultFastFeatures)
}

func IsSupportedFramework(framework string) bool {
	return slices.Contains(supportedFrameworks, framework)
}

func NormalizeFeatureNames(raw []string) ([]string, error) {
	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, f := range raw {
		trimmed := strings.TrimSpace(strings.ToLower(f))
		if trimmed == "" {
			continue
		}
		canonical, ok := featureAliasToCanonical[trimmed]
		if !ok {
			return nil, fmt.Errorf("unsupported feature %q", f)
		}
		if _, ok := seen[canonical]; ok {
			continue
		}
		seen[canonical] = struct{}{}
		out = append(out, canonical)
	}
	return out, nil
}

func ParseFeatureCSV(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	return NormalizeFeatureNames(strings.Split(raw, ","))
}

func FeatureSetFromNames(features []string) FeatureSet {
	set := FeatureSet{}
	for _, feature := range features {
		switch feature {
		case FeatureClean:
			set.Clean = true
		case FeatureLogger:
			set.Logger = true
		case FeaturePostgreSQL:
			set.PostgreSQL = true
		case FeatureJWT:
			set.JWT = true
		case FeatureDocker:
			set.Docker = true
		}
	}
	return set
}

func TemplateCatalog() []TemplateInfo {
	return []TemplateInfo{
		{TemplatePath: "templates/env.tmpl", Frameworks: []string{"all"}, Features: []string{"base"}},
		{TemplatePath: "templates/index.html.tmpl", Frameworks: []string{"all"}, Features: []string{"base"}},
		{TemplatePath: "templates/startup.go.tmpl", Frameworks: []string{"all"}, Features: []string{"base"}},
		{TemplatePath: "templates/main.go.tmpl", Frameworks: Frameworks(), Features: []string{"base"}},
		{TemplatePath: "templates/routes/welcome.go.tmpl", Frameworks: []string{"all"}, Features: []string{"base"}},
		{TemplatePath: "templates/routes/gin_routes.go.tmpl", Frameworks: []string{FrameworkGin}, Features: []string{"base"}},
		{TemplatePath: "templates/routes/echo_routes.go.tmpl", Frameworks: []string{FrameworkEcho}, Features: []string{"base"}},
		{TemplatePath: "templates/routes/fiber_routes.go.tmpl", Frameworks: []string{FrameworkFiber}, Features: []string{"base"}},
		{TemplatePath: "templates/routes/nethttp_routes.go.tmpl", Frameworks: []string{FrameworkNetHTTP}, Features: []string{"base"}},
		{TemplatePath: "templates/logger.go.tmpl", Frameworks: []string{"all"}, Features: []string{FeatureLogger}},
		{TemplatePath: "templates/database.go.tmpl", Frameworks: []string{"all"}, Features: []string{FeaturePostgreSQL}},
		{TemplatePath: "templates/jwt.go.tmpl", Frameworks: []string{"all"}, Features: []string{FeatureJWT}},
		{TemplatePath: "templates/Dockerfile.tmpl", Frameworks: []string{"all"}, Features: []string{FeatureDocker}},
		{TemplatePath: "templates/clean/domain/health.go.tmpl", Frameworks: []string{"all"}, Features: []string{FeatureClean}},
		{TemplatePath: "templates/clean/usecase/health_usecase.go.tmpl", Frameworks: []string{"all"}, Features: []string{FeatureClean}},
		{TemplatePath: "templates/clean/repository/health_repository.go.tmpl", Frameworks: []string{"all"}, Features: []string{FeatureClean}},
		{TemplatePath: "templates/clean/delivery/httptransport/health_handler.go.tmpl", Frameworks: []string{"all"}, Features: []string{FeatureClean}},
		{TemplatePath: "templates/clean/app/bootstrap.go.tmpl", Frameworks: []string{"all"}, Features: []string{FeatureClean}},
	}
}

func BuildTemplateData(projectName, framework string, features []string) TemplateData {
	return TemplateData{
		ModuleName:  projectName,
		ProjectName: projectName,
		Framework:   framework,
		Features:    FeatureSetFromNames(features),
	}
}

func getRootDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(filepath.Join(wd, "templates")); err == nil {
		return wd, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exePath)
	if _, err := os.Stat(filepath.Join(dir, "templates")); err == nil {
		return dir, nil
	}

	return "", fmt.Errorf("templates folder not found")
}

func GenerateProject(opts Options) (retErr error) {
	if strings.TrimSpace(opts.ProjectName) == "" {
		return errors.New("project name is required")
	}
	if !IsSupportedFramework(opts.Framework) {
		return fmt.Errorf("unsupported framework %q", opts.Framework)
	}

	normalizedFeatures, err := NormalizeFeatureNames(opts.Features)
	if err != nil {
		return err
	}
	opts.Features = normalizedFeatures

	templateFS := opts.TemplateFS
	if templateFS == nil {
		rootDir := opts.RootDir
		if strings.TrimSpace(rootDir) == "" {
			rootDir, err = getRootDir()
		}
		if err == nil && strings.TrimSpace(rootDir) != "" {
			templateFS = os.DirFS(rootDir)
		} else {
			templateFS = lalibelacli.EmbeddedTemplates
		}
	}

	projectPath := filepath.Join(".", opts.ProjectName)
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("project directory %q already exists", projectPath)
	}

	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		return fmt.Errorf("creating project directory: %w", err)
	}

	defer func() {
		if retErr == nil {
			return
		}
		if err := os.RemoveAll(projectPath); err != nil {
			retErr = fmt.Errorf("%w; rollback failed: %v", retErr, err)
			return
		}
		retErr = fmt.Errorf("%w; rollback complete", retErr)
	}()

	runner := opts.Runner
	if runner == nil {
		runner = utils.RunCommand
	}

	status := opts.Status
	if status == nil {
		status = func(string, int, int) {}
	}

	ctx := &generationContext{
		templateFS:  templateFS,
		projectPath: projectPath,
		data:        BuildTemplateData(opts.ProjectName, opts.Framework, opts.Features),
		runner:      runner,
	}

	steps := buildSteps(ctx.data.Features, ctx.data.Framework)
	for i, step := range steps {
		status(step.name, i+1, len(steps))
		if err := step.fn(ctx); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}

	return nil
}

func buildSteps(features FeatureSet, framework string) []generationStep {
	steps := []generationStep{
		{name: "creating directory structure", fn: createProjectDirectories},
		{name: "rendering base templates", fn: generateBaseTemplates},
	}

	switch framework {
	case FrameworkGin:
		steps = append(steps, generationStep{name: "generating gin scaffold", fn: generateGinScaffold})
	case FrameworkEcho:
		steps = append(steps, generationStep{name: "generating echo scaffold", fn: generateEchoScaffold})
	case FrameworkFiber:
		steps = append(steps, generationStep{name: "generating fiber scaffold", fn: generateFiberScaffold})
	case FrameworkNetHTTP:
		steps = append(steps, generationStep{name: "generating net/http scaffold", fn: generateNetHTTPScaffold})
	}

	if features.Clean {
		steps = append(steps, generationStep{name: "generating clean architecture layer", fn: generateCleanArchitectureFeature})
	}
	if features.Logger {
		steps = append(steps, generationStep{name: "generating logger feature", fn: generateLoggerFeature})
	}
	if features.PostgreSQL {
		steps = append(steps, generationStep{name: "generating postgresql feature", fn: generatePostgreSQLFeature})
	}
	if features.JWT {
		steps = append(steps, generationStep{name: "generating jwt feature", fn: generateJWTFeature})
	}
	if features.Docker {
		steps = append(steps, generationStep{name: "generating docker feature", fn: generateDockerFeature})
	}

	steps = append(steps, generationStep{name: "setting up go modules", fn: setupDependencies})
	return steps
}

func createProjectDirectories(ctx *generationContext) error {
	baseFolders := []string{
		filepath.Join(ctx.projectPath, "config"),
		filepath.Join(ctx.projectPath, "internal", "routes"),
		filepath.Join(ctx.projectPath, "internal", "middleware"),
	}
	if ctx.data.Features.Clean {
		baseFolders = append(baseFolders,
			filepath.Join(ctx.projectPath, "internal", "domain"),
			filepath.Join(ctx.projectPath, "internal", "usecase"),
			filepath.Join(ctx.projectPath, "internal", "repository"),
			filepath.Join(ctx.projectPath, "internal", "delivery", "httptransport"),
			filepath.Join(ctx.projectPath, "internal", "app"),
		)
	}
	for _, folder := range baseFolders {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func generateBaseTemplates(ctx *generationContext) error {
	base := []struct {
		templatePath string
		outputPath   string
	}{
		{templatePath: "templates/env.tmpl", outputPath: ".env"},
		{templatePath: "templates/index.html.tmpl", outputPath: filepath.Join("templates", "index.html")},
		{templatePath: "templates/startup.go.tmpl", outputPath: "startup.go"},
		{templatePath: "templates/routes/welcome.go.tmpl", outputPath: filepath.Join("internal", "routes", "welcome.go")},
	}

	for _, item := range base {
		if err := renderProjectTemplate(ctx, item.templatePath, item.outputPath); err != nil {
			return err
		}
	}
	return nil
}

func generateGinScaffold(ctx *generationContext) error {
	if err := renderProjectTemplate(ctx, "templates/main.go.tmpl", "main.go"); err != nil {
		return err
	}
	return renderProjectTemplate(ctx, "templates/routes/gin_routes.go.tmpl", filepath.Join("internal", "routes", "routes.go"))
}

func generateEchoScaffold(ctx *generationContext) error {
	if err := renderProjectTemplate(ctx, "templates/main.go.tmpl", "main.go"); err != nil {
		return err
	}
	return renderProjectTemplate(ctx, "templates/routes/echo_routes.go.tmpl", filepath.Join("internal", "routes", "routes.go"))
}

func generateFiberScaffold(ctx *generationContext) error {
	if err := renderProjectTemplate(ctx, "templates/main.go.tmpl", "main.go"); err != nil {
		return err
	}
	return renderProjectTemplate(ctx, "templates/routes/fiber_routes.go.tmpl", filepath.Join("internal", "routes", "routes.go"))
}

func generateNetHTTPScaffold(ctx *generationContext) error {
	if err := renderProjectTemplate(ctx, "templates/main.go.tmpl", "main.go"); err != nil {
		return err
	}
	return renderProjectTemplate(ctx, "templates/routes/nethttp_routes.go.tmpl", filepath.Join("internal", "routes", "routes.go"))
}

func generateLoggerFeature(ctx *generationContext) error {
	return renderProjectTemplate(ctx, "templates/logger.go.tmpl", filepath.Join("config", "logger.go"))
}

func generatePostgreSQLFeature(ctx *generationContext) error {
	return renderProjectTemplate(ctx, "templates/database.go.tmpl", filepath.Join("config", "database.go"))
}

func generateJWTFeature(ctx *generationContext) error {
	return renderProjectTemplate(ctx, "templates/jwt.go.tmpl", filepath.Join("internal", "middleware", "jwt.go"))
}

func generateDockerFeature(ctx *generationContext) error {
	return renderProjectTemplate(ctx, "templates/Dockerfile.tmpl", "Dockerfile")
}

func generateCleanArchitectureFeature(ctx *generationContext) error {
	files := []struct {
		templatePath string
		outputPath   string
	}{
		{templatePath: "templates/clean/domain/health.go.tmpl", outputPath: filepath.Join("internal", "domain", "health.go")},
		{templatePath: "templates/clean/usecase/health_usecase.go.tmpl", outputPath: filepath.Join("internal", "usecase", "health_usecase.go")},
		{templatePath: "templates/clean/repository/health_repository.go.tmpl", outputPath: filepath.Join("internal", "repository", "health_repository.go")},
		{templatePath: "templates/clean/delivery/httptransport/health_handler.go.tmpl", outputPath: filepath.Join("internal", "delivery", "httptransport", "health_handler.go")},
		{templatePath: "templates/clean/app/bootstrap.go.tmpl", outputPath: filepath.Join("internal", "app", "bootstrap.go")},
	}
	for _, file := range files {
		if err := renderProjectTemplate(ctx, file.templatePath, file.outputPath); err != nil {
			return err
		}
	}
	return nil
}

func setupDependencies(ctx *generationContext) error {
	if err := ctx.runner(ctx.projectPath, "go", "mod", "init", ctx.data.ModuleName); err != nil {
		return fmt.Errorf("go mod init failed: %w", err)
	}
	if err := ctx.runner(ctx.projectPath, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}
	return nil
}

func renderProjectTemplate(ctx *generationContext, templateRelativePath, outputRelativePath string) error {
	return renderTemplateFromFS(
		ctx.templateFS,
		templateRelativePath,
		filepath.Join(ctx.projectPath, outputRelativePath),
		ctx.data,
	)
}

func renderTemplateFromFS(templateFS fs.FS, templatePath, outputPath string, data TemplateData) error {
	tmpl, err := template.ParseFS(templateFS, templatePath)
	if err != nil {
		return fmt.Errorf("failed parsing template %s: %v", templatePath, err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed creating parent directory for %s: %v", outputPath, err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed creating file %s: %v", outputPath, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed executing template %s: %v", templatePath, err)
	}
	return nil
}

func renderTemplate(templatePath, outputPath string, data TemplateData) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed parsing template %s: %v", templatePath, err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed creating parent directory for %s: %v", outputPath, err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed creating file %s: %v", outputPath, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed executing template %s: %v", templatePath, err)
	}
	return nil
}

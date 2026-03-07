package logger

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

// Feature installs the "logger" scaffold feature.
type Feature struct{}

// New returns a new "logger" feature installer.
func New() Feature { return Feature{} }

// Name returns the registry name of the feature.
func (Feature) Name() string { return "logger" }

// Compatible reports whether the feature supports a given framework.
func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("logger", framework)
}

// Install writes the feature's scaffold files into projectRoot.
func (Feature) Install(projectRoot string) error {
	const file = `package logger

import (
	"log/slog"
	"os"
	"strings"
)

func New(level string) *slog.Logger {
	var slogLevel slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})
	return slog.New(handler)
}
`
	return shared.WriteFileIfMissing(projectRoot, "internal/logger/logger.go", []byte(file))
}

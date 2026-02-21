package config

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "config" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("config", framework)
}

func (Feature) Install(projectRoot string) error {
	const file = `package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName string
	Env     string
	Port    int
}

func Load() (Config, error) {
	cfg := Config{
		AppName: strings.TrimSpace(os.Getenv("APP_NAME")),
		Env:     strings.TrimSpace(os.Getenv("APP_ENV")),
		Port:    8080,
	}
	if cfg.AppName == "" {
		cfg.AppName = "lalibela-app"
	}
	if cfg.Env == "" {
		cfg.Env = "development"
	}

	if raw := strings.TrimSpace(os.Getenv("PORT")); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("invalid PORT %q", raw)
		}
		cfg.Port = port
	}

	return cfg, nil
}
`
	return shared.WriteFileIfMissing(projectRoot, "internal/config/config.go", []byte(file))
}

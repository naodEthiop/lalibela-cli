package gracefulshutdown

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "graceful-shutdown" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("graceful-shutdown", framework)
}

func (Feature) Install(projectRoot string) error {
	const file = `package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitForShutdown(logger *slog.Logger, shutdownTimeout time.Duration, shutdownFn func(context.Context) error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := shutdownFn(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		return
	}
	logger.Info("server shutdown complete")
}
`
	return shared.WriteFileIfMissing(projectRoot, "internal/server/graceful_shutdown.go", []byte(file))
}

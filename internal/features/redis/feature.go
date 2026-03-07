package redis

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

// Feature installs the "redis" scaffold feature.
type Feature struct{}

// New returns a new "redis" feature installer.
func New() Feature { return Feature{} }

// Name returns the registry name of the feature.
func (Feature) Name() string { return "redis" }

// Compatible reports whether the feature supports a given framework.
func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("redis", framework)
}

// Install writes the feature's scaffold files into projectRoot.
func (Feature) Install(projectRoot string) error {
	const file = `package storage

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = client.Ping(ctx)
	return client
}
`
	return shared.WriteFileIfMissing(projectRoot, "internal/storage/redis.go", []byte(file))
}

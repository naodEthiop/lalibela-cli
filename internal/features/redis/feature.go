package redis

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "redis" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("redis", framework)
}

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

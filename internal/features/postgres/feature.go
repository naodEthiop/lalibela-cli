package postgres

import (
	"github.com/naodEthiop/lalibela-cli/internal/features/shared"
)

type Feature struct{}

func New() Feature { return Feature{} }

func (Feature) Name() string { return "postgres" }

func (Feature) Compatible(framework string) bool {
	return shared.IsFeatureCompatible("postgres", framework)
}

func (Feature) Install(projectRoot string) error {
	const source = `package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return pgxpool.New(ctx, dsn)
}
`
	const migration = `-- 0001_init.sql
-- Add project migrations here.
`
	if err := shared.WriteFileIfMissing(projectRoot, "internal/storage/postgres.go", []byte(source)); err != nil {
		return err
	}
	return shared.WriteFileIfMissing(projectRoot, "db/migrations/0001_init.sql", []byte(migration))
}

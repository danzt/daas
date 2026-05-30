//go:build integration

// Package testhelpers provides shared testing infrastructure for integration
// tests that require real external services (PostgreSQL via Docker).
package testhelpers

import (
	"context"
	"fmt"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// PostgresContainer wraps a testcontainers PostgreSQL container with a
// ready-to-use DSN and a cleanup function.
type PostgresContainer struct {
	Container *tcpostgres.PostgresContainer
	DSN       string
	Cleanup   func(context.Context) error
}

// NewPostgresContainer starts a PostgreSQL 17 container and returns the
// container with its DSN and a cleanup function.
//
// Callers are responsible for applying migrations after the container starts.
// Callers must call Cleanup when the test completes:
//
//	container, err := testhelpers.NewPostgresContainer(ctx)
//	if err != nil { t.Fatal(err) }
//	t.Cleanup(func() { container.Cleanup(ctx) })
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pg, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("daas_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("get postgres DSN: %w", err)
	}

	cleanup := func(ctx context.Context) error {
		return pg.Terminate(ctx)
	}

	return &PostgresContainer{
		Container: pg,
		DSN:       dsn,
		Cleanup:   cleanup,
	}, nil
}

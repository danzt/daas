package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	adapterhttp "github.com/danzt/daas/api/internal/adapter/http"
	"github.com/danzt/daas/api/internal/config"
)

func main() {
	// Configure zerolog for structured JSON output
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Load .env file if present (ignored in production where env vars are injected)
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("no .env file found, using environment variables")
	}

	cfg := config.Load()

	// Connect to PostgreSQL via pgxpool
	ctx := context.Background()
	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		var err error
		pool, err = pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create database connection pool")
		}
		defer pool.Close()
		if err := pool.Ping(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed to ping database")
		}
		log.Info().Msg("database connection pool ready")
	} else {
		log.Warn().Msg("DATABASE_URL not set — database features will be unavailable")
	}

	router := adapterhttp.NewRouterWithConfig(adapterhttp.RouterConfig{
		SupabaseURL:    cfg.SupabaseURL,
		ServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		Pool:           pool,
	})

	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("starting daas api server")

	if err := router.Start(addr); err != nil {
		log.Fatal().Err(err).Msg("server stopped unexpectedly")
	}
}

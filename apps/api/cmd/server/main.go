package main

import (
	"context"
	"fmt"
	"net"
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
		poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse database URL")
		}
		// Force IPv4 — resolves hostname to IPv4 addresses to avoid IPv6 routing issues
		poolCfg.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("split host port: %w", err)
			}
			ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip4", host)
			if err != nil || len(ips) == 0 {
				return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
			}
			var firstErr error
			for _, ip := range ips {
				target := net.JoinHostPort(ip.String(), port)
				conn, err := (&net.Dialer{}).DialContext(ctx, "tcp4", target)
				if err == nil {
					return conn, nil
				}
				if firstErr == nil {
					firstErr = err
				}
			}
			return nil, fmt.Errorf("dial %s: all IPv4 addresses unreachable: %w", host, firstErr)
		}
		pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
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
		AnonKey:        cfg.SupabaseAnonKey,
		ServiceRoleKey: cfg.SupabaseServiceRoleKey,
		Pool:           pool,
	})

	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("starting daas api server")

	if err := router.Start(addr); err != nil {
		log.Fatal().Err(err).Msg("server stopped unexpectedly")
	}
}

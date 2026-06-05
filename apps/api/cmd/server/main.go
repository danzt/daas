package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	adapterhttp "github.com/danzt/daas/api/internal/adapter/http"
	notifadapter "github.com/danzt/daas/api/internal/adapter/notification"
	"github.com/danzt/daas/api/internal/adapter/storage"
	"github.com/danzt/daas/api/internal/config"
	"github.com/danzt/daas/api/internal/domain/notification"
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
		// Prefer IPv4 to avoid VPN/routing issues with IPv6-only Supabase hosts.
		// Falls back to IPv6 if no A record exists (e.g. Supabase direct DB host
		// only publishes AAAA). Dialer has a 10s timeout per attempt.
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		poolCfg.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("split host port: %w", err)
			}
			lookupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			// Try IPv4 first.
			if ips, err := net.DefaultResolver.LookupNetIP(lookupCtx, "ip4", host); err == nil && len(ips) > 0 {
				var firstErr error
				for _, ip := range ips {
					target := net.JoinHostPort(ip.String(), port)
					if conn, err := dialer.DialContext(ctx, "tcp4", target); err == nil {
						return conn, nil
					} else if firstErr == nil {
						firstErr = err
					}
				}
				// IPv4 addresses found but all unreachable — don't fall through to IPv6.
				return nil, fmt.Errorf("dial %s: all IPv4 addresses unreachable: %w", host, firstErr)
			}

			// No IPv4 record — fall back to IPv6 (e.g. Supabase AAAA-only hosts).
			if ips, err := net.DefaultResolver.LookupNetIP(lookupCtx, "ip6", host); err == nil && len(ips) > 0 {
				var firstErr error
				for _, ip := range ips {
					target := net.JoinHostPort(ip.String(), port)
					if conn, err := dialer.DialContext(ctx, "tcp6", target); err == nil {
						return conn, nil
					} else if firstErr == nil {
						firstErr = err
					}
				}
				return nil, fmt.Errorf("dial %s: all IPv6 addresses unreachable: %w", host, firstErr)
			}

			// Last resort: let the OS resolve normally.
			return dialer.DialContext(ctx, "tcp", addr)
		}
		pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create database connection pool")
		}
		defer pool.Close()
		// Ping with a 15s deadline so a slow/unreachable DB fails fast.
		pingCtx, pingCancel := context.WithTimeout(ctx, 15*time.Second)
		defer pingCancel()
		if err := pool.Ping(pingCtx); err != nil {
			log.Fatal().Err(err).Msg("failed to ping database")
		}
		log.Info().Msg("database connection pool ready")
	} else {
		log.Warn().Msg("DATABASE_URL not set — database features will be unavailable")
	}

	// Notification service. EmailAdapter activates when RESEND_API_KEY is set;
	// otherwise we fall back to NoopAdapter so the dependency graph still
	// compiles and dev runs don't spam real inboxes.
	var notifSvc notification.NotificationService
	if resendKey := os.Getenv("RESEND_API_KEY"); resendKey != "" {
		from := os.Getenv("RESEND_FROM_EMAIL")
		if from == "" {
			from = "noreply@daas.app"
		}
		notifSvc = notifadapter.NewEmailAdapter(resendKey, from)
		log.Info().Str("from", from).Msg("notification service: EmailAdapter (Resend)")
	} else {
		notifSvc = notifadapter.NewNoopAdapter()
		log.Info().Msg("notification service: NoopAdapter (RESEND_API_KEY unset)")
	}

	// Storefront base URL used by lifecycle emails to build the customer
	// tracking links. Defaults to localhost for dev.
	storefrontURL := os.Getenv("PUBLIC_STOREFRONT_URL")
	if storefrontURL == "" {
		storefrontURL = "http://localhost:3000"
	}

	// Storage adapter for payment proof uploads.
	// In production, swap LocalFSAdapter for SupabaseStorageAdapter (future PR).
	storageRootDir := os.Getenv("STORAGE_ROOT_DIR")
	if storageRootDir == "" {
		storageRootDir = "./storage"
	}
	storagePublicBase := os.Getenv("STORAGE_PUBLIC_BASE")
	if storagePublicBase == "" {
		storagePublicBase = "http://localhost:8080/files"
	}
	storageAdapter := storage.NewLocalFSAdapter(storageRootDir, storagePublicBase)
	log.Info().
		Str("root_dir", storageRootDir).
		Str("public_base", storagePublicBase).
		Msg("storage adapter: LocalFSAdapter")

	router := adapterhttp.NewRouterWithConfig(adapterhttp.RouterConfig{
		SupabaseURL:     cfg.SupabaseURL,
		AnonKey:         cfg.SupabaseAnonKey,
		ServiceRoleKey:  cfg.SupabaseServiceRoleKey,
		Pool:            pool,
		NotificationSvc: notifSvc,
		StorefrontURL:   storefrontURL,
		Storage:         storageAdapter,
	})

	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("starting daas api server")

	if err := router.Start(addr); err != nil {
		log.Fatal().Err(err).Msg("server stopped unexpectedly")
	}
}

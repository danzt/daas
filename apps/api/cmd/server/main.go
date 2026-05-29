package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/danzt/daas/api/internal/config"
	adapterhttp "github.com/danzt/daas/api/internal/adapter/http"
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

	router := adapterhttp.NewRouter()

	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("starting daas api server")

	if err := router.Start(addr); err != nil {
		log.Fatal().Err(err).Msg("server stopped unexpectedly")
	}
}

/*
Package config handles loading and validating application configuration
from environment variables. All values have sensible defaults so the
application can start with zero environment setup during local development.
*/
package config

import (
	"log/slog" // slog = structured logging library
	"os"       // used .Getenv calls and write logs to stdout.
)

// Config struct holds all configuration values for the application.
// values are read once at startup and passed through the app via dependency injection.
// no global config variable is used. callers receive a *Config explicitly,
// making dependencies visible and the code easier to test.
type Config struct {
	// Port is the TCP port the HTTP server listens on
	Port string

	// the file path to the SQLite database file
	// when switching to Postgres, this field becomes the DSN connection string.
	DBPath string

	// the base directory on disk where extracted deployment
	// files are written. each deployment gets its own subdirectory here,
	// which is bind-mounted into the Nginx container.
	ServeRoot string

	// the base directory where build and deploy log files are written.
	// one log file per deployment, named by slug.
	LogRoot string

	// TraefikNetwork is the Docker network name that Traefik and all
	// per-deployment Nginx containers are connected to.
	TraefikNetwork string

	// LogFormat controls the output format of slog (logging library)
	// accepted values: "json" (default) | "text"
	// set to "text" during local development for readable terminal output
	LogFormat string
}

// Load reads configuration from environment variables and RETURNS a populated Config struct.
// missing environment variables fall back to safe local development defaults
// so the app can run without any setup during early development.
func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DBPath:         getEnv("DB_PATH", "./corvus.db"),
		ServeRoot:      getEnv("SERVE_ROOT", "./data/deployments"),
		LogRoot:        getEnv("LOG_ROOT", "./data/logs"),
		TraefikNetwork: getEnv("TRAEFIK_NETWORK", "paas-network"),
		LogFormat:      getEnv("LOG_FORMAT", "json"),
	}
}

// getEnv retrieves the value of an environment variable by key.
// if the variable is not set or is empty, the provided fallback value is returned.
// this avoids scattered os.Getenv calls with inline fallback logic throughout the codebase.
func getEnv(key, fallbackValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallbackValue
}

// NewLogger constructs a *slog.Logger based on the LogFormat field of the config.
// "text" produces human-readable output for local development
// any other value (including "json") produces structured JSON output for production
// and Docker log shipping.
func (config *Config) NewLogger() *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		// AddSource adds the file name and line number to each log record.
		// useful during development to trace log origins.
		AddSource: true,
		Level:     slog.LevelDebug,
	}

	if config.LogFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

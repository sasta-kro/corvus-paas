/*
Package config handles loading and validating application configuration
from environment variables. All values have sensible defaults so the
application can start with zero environment setup during local development.
*/
package config

import (
	"log/slog"      // slog = structured log. used for json logging in this app
	"os"            // used .Getenv calls and write logs to stdout.
	"path/filepath" // used to extract file base name form absolute path in logging.
)

// AppConfig struct holds all configuration values for the application.
// values are read once at startup and passed through the app via dependency injection.
// no global config variable is used. callers receive a *AppConfig explicitly,
// making dependencies visible and the code easier to test.
type AppConfig struct {
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

// NewLogger constructs a *slog.Logger based on the LogFormat field of the config.
// "text" produces human-readable output for local development
// any other value (including "json") produces structured JSON output for production
// and Docker log shipping.
// *AppConfig is a pointer receiver rather than a value receiver cuz copying AppConfig struct unnecessary
// returning a pointer *slog.Logger rather than value is standard for complex objects
// like loggers, database connections, or servers. It forces things to use the same logger instance.
func (config *AppConfig) NewLogger() *slog.Logger {
	var handler slog.Handler // declaration of slog.Handler interface variable to hold the chosen log handler

	// Syntax confusion - `slog.` is the package name, `HandlerOptions` is a struct type defined in slog package.
	// &slog.HandlerOptions{} creates a new instance of HandlerOptions struct and returns its pointer rather than value
	// {} is to initialize the struct's fields
	options := &slog.HandlerOptions{
		// AddSource adds the file name and line number to each log record
		// useful during development to trace log origins.
		AddSource: true, // this returns the absolute file path which is too long and eyesore
		Level:     slog.LevelDebug,

		/* ReplaceAttr is a build-in field (key) that accepts a function, that runs on every log call.

		When the logger processes a log record, the logger checks each attribute (key-value pair)
		like looping through them and runs the ReplaceAttr function on EACH attribute.
		If the function returns a modified attribute, the logger uses that instead of the original.
		`groups []string` is the list of strings if there are nested logs.
		`attribute slog.Attr` is the current attribute being processed.
		`slog.Attr` after the args is the return type
		*/
		ReplaceAttr: func(groups []string, attribute slog.Attr) slog.Attr {
			// Check if the current attribute is the "source" (file path/line info)
			if attribute.Key == slog.SourceKey {
				/*
					attribute.Value.Any(): The slog value is wrapped in a special type-safe container.
					This "unwraps" it to see what's inside.
					(*slog.Source) is like type casting in other languages.
				*/
				source := attribute.Value.Any().(*slog.Source)
				// This takes the file's absolute path and just returns the filename
				source.File = filepath.Base(source.File)
			}
			return attribute
		},
	}

	if config.LogFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, options) // text for local dev
	} else {
		handler = slog.NewJSONHandler(os.Stdout, options) // json for prod
	}

	// returns new logger with chosen handler
	return slog.New(handler)
}

// LoadAppConfig reads configuration from environment variables and RETURNS a populated AppConfig struct.
// missing environment variables fall back to safe local development defaults
// so the app can run without any setup during early development.
// TODO: to move on from hard coded and actually make a external file to load the config data
func LoadAppConfig() *AppConfig {
	// create a new AppConfig struct with values loaded from environment variables or defaults
	// returns pointer to AppConfig struct created
	return &AppConfig{
		Port:           getEnv("PORT", "8080"),
		DBPath:         getEnv("DB_PATH", "./corvus.db"),
		ServeRoot:      getEnv("SERVE_ROOT", "./data/deployments"),
		LogRoot:        getEnv("LOG_ROOT", "./data/logs"),
		TraefikNetwork: getEnv("TRAEFIK_NETWORK", "corvus-paas-network"),
		LogFormat:      getEnv("LOG_FORMAT", "text"),
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

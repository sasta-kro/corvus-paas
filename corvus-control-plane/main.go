package main

import (
	"fmt"
	"log"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/config"
	"github.com/sasta-kro/corvus-paas/corvus-control-plane/db"
)

func main() {
	appConfig := config.LoadConfig() // loads the config and stores pointer
	logger := appConfig.NewLogger()  // return a logger (slog) based on `LogFormat` (text or json)

	/*
		logger.Info() aka `slog.Logger.Info()` is just a glorified print
		The first argument is always the message (the human-readable part).
		Every argument after that must come in pairs: a Key (string) followed by a Value (any type).

		Different log levels in slog.
		Debug:	Extremely detailed info for dev (often hidden in production).
		Info:	General "heartbeat" events (starting up, stopping).
		Warn:	Something is weird, but the app is still running.
		Error:	Something broke (database connection failed, etc.).
	*/
	logger.Info("corvus control plane starting", // this log is level "Info"
		"port", appConfig.Port,
		"db_path", appConfig.DBPath,
		"log_format", appConfig.LogFormat,
	)

	// temp - confirm appConfig values loaded correctly
	fmt.Printf("Config loaded: port=%s db=%s\n", appConfig.Port, appConfig.DBPath)

	// opening the database and run schema migration (init tables)
	// if this fails, the application cannot serve requests, so exit immediately
	database, err := db.OpenDatabase(appConfig.DBPath, logger)
	if err != nil {
		// If the database cannot be opened or migrated, the application
		// cannot function and must "fail fast". the standard library's
		// log.Fatalf is used here because it synchronously writes to standard error
		// before forcing an os.Exit(1), guaranteeing the crash reason is
		// printed to the console. (Using a structured logger followed by os.Exit()
		// risks losing the final log entry if the logger buffers its output, which they usually do).
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.CloseDatabase()

	logger.Info("startup complete, ready to serve", "port", appConfig.Port) // ensures the database connection is closed when main() exits
}

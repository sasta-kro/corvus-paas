package main

import (
	"fmt"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/config"
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

	// temporary: confirm appConfig values loaded correctly
	fmt.Printf("Config loaded: port=%s db=%s\n", appConfig.Port, appConfig.DBPath)
}

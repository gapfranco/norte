//go:build headless

package main

import (
	"log/slog"

	"norte/config"
)

func runDesktop(app *application, cfg config.Config, logger *slog.Logger) {
	runServer(app, cfg, logger)
}

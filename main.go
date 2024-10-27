package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gellen89/mam-update-go/internal/appdir"
	"github.com/gellen89/mam-update-go/internal/mamupdater"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	appdirs, err := appdir.GetAppDirs("mamdynupdate")
	if err != nil {
		logger.Error("Failed to get app directories", "error", err)
		os.Exit(1)
	}

	if err := appdirs.EnsureDirs(); err != nil {
		logger.Error("Failed to ensure directories exist", "error", err)
		os.Exit(1)
	}

	envMamId := os.Getenv("MAM_ID")
	var mamId *string
	if envMamId != "" {
		mamId = &envMamId
	}

	config := mamupdater.Config{
		DataDir:     appdirs.Data,
		CookiePath:  filepath.Join(appdirs.Data, "MAM.cookie"),
		IpPath:      filepath.Join(appdirs.Data, "MAM.ip"),
		LastRunPath: filepath.Join(appdirs.Data, "last_run_time"),
		MamId:       mamId,
		Logger:      logger,
	}

	updater, err := mamupdater.NewMamUpdater(config)
	if err != nil {
		logger.Error("Failed to create updater", "error", err)
		os.Exit(1)
	}

	if err := updater.Run(); err != nil {
		logger.Error("Failed to run updater", "error", err)
		os.Exit(1)
	}
}

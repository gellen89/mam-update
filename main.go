package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gellen89/mam-update-go/internal/appdir"
	"github.com/gellen89/mam-update-go/internal/mamupdater"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	flagCfg := getFlags()

	mamId := getMamId(flagCfg)

	var appdirs *appdir.AppDirs
	if flagCfg.ConfigDir != nil && *flagCfg.ConfigDir != "" {
		appdirs = appdir.New(*flagCfg.ConfigDir)
	} else {
		var err error
		appdirs, err = appdir.NewFromAppName(".mamupdate")
		if err != nil {
			logger.Error("Failed to get app directories", "error", err)
			os.Exit(1)
		}
	}

	if err := appdirs.EnsureDirs(); err != nil {
		logger.Error("Failed to ensure directories exist", "error", err)
		os.Exit(1)
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

type flagConfig struct {
	MamId     *string
	ConfigDir *string
}

func getFlags() *flagConfig {
	var mamId string
	flag.StringVar(&mamId, "mam-id", "", "Provide the mam-id used for the initial request.")
	var mamDir string
	flag.StringVar(&mamDir, "mam-dir", "", "Provide the directory where the config and data will be persisted.")

	flag.Parse()

	return &flagConfig{
		MamId:     &mamId,
		ConfigDir: &mamDir,
	}
}

func getMamId(flagCfg *flagConfig) *string {
	if flagCfg.MamId != nil && *flagCfg.MamId != "" {
		return flagCfg.MamId
	}
	envMamId := os.Getenv("MAM_ID")
	if envMamId != "" {
		return &envMamId
	}
	return nil
}

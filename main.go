package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gellen89/mam-update/internal/appdir"
	"github.com/gellen89/mam-update/internal/mamupdater"
)

func main() {
	flagCfg := getFlags()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(flagCfg),
	}))

	mamId := getMamId(flagCfg)
	mamDir := getMamDir(flagCfg)

	var appdirs *appdir.AppDirs
	if mamDir != nil && *mamDir != "" {
		appdirs = appdir.New(*mamDir)
	} else {
		var err error
		appdirs, err = appdir.NewFromAppName(".mamupdate")
		if err != nil {
			logger.Error("Failed to get app directories", "error", err)
			os.Exit(1)
		}
	}
	logger.Debug(fmt.Sprintf("using mam dir: %s", appdirs.Data))

	if err := appdirs.EnsureDirs(); err != nil {
		logger.Error("Failed to ensure directories exist", "error", err)
		os.Exit(1)
	}

	config := &mamupdater.Config{
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
	ctx := context.Background()
	if err := updater.Run(ctx); err != nil {
		logger.Error("Failed to run updater", "error", err)
		os.Exit(1)
	}
	logger.Debug("run completed")
}

type flagConfig struct {
	MamId     *string
	ConfigDir *string
	Force     bool
	LogLevel  *string
}

func getFlags() *flagConfig {
	var mamId string
	flag.StringVar(&mamId, "mam-id", "", "Provide the mam-id used for the initial request.")
	var mamDir string
	flag.StringVar(&mamDir, "mam-dir", "", "Provide the directory where the config and data will be persisted.")
	var force bool
	flag.BoolVar(&force, "force", false, "Specify force to override the last run time.")
	var loglevel string
	flag.StringVar(&loglevel, "level", "info", "Specify a log level (debug, info, warn, error) default: info.")

	flag.Parse()

	return &flagConfig{
		MamId:     &mamId,
		ConfigDir: &mamDir,
		Force:     force,
		LogLevel:  &loglevel,
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

func getMamDir(flagCfg *flagConfig) *string {
	if flagCfg.ConfigDir != nil && *flagCfg.ConfigDir != "" {
		return flagCfg.ConfigDir
	}
	envMamDir := os.Getenv("MAMUPDATE_DIR")
	if envMamDir != "" {
		return &envMamDir
	}
	return nil
}

func getLogLevel(flagCfg *flagConfig) slog.Level {
	if flagCfg.LogLevel != nil && *flagCfg.LogLevel != "" {
		return toSlogLevel(*flagCfg.LogLevel)
	}
	envlevel := os.Getenv("LOG_LEVEL")
	if envlevel != "" {
		return toSlogLevel(envlevel)
	}
	return slog.LevelInfo
}

func toSlogLevel(input string) slog.Level {
	switch strings.ToLower(input) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

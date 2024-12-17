package app

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

type App struct {
	config  *mamupdater.Config
	appdirs *appdir.AppDirs
}

func New(args []string) (*App, error) {
	// Parse flags
	flagCfg, err := getFlags(args)
	if err != nil {
		return nil, fmt.Errorf("unable to parse flags: %w", err)
	}

	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(flagCfg),
	}))

	// Determine data directory
	appDirs, err := getAppDirs(flagCfg)
	if err != nil {
		return nil, err
	}

	// Build the configuration
	config := &mamupdater.Config{
		DataDir:     appDirs.Data,
		CookiePath:  filepath.Join(appDirs.Data, "MAM.cookie"),
		IpPath:      filepath.Join(appDirs.Data, "MAM.ip"),
		LastRunPath: filepath.Join(appDirs.Data, "last_run_time"),
		MamId:       getMamId(flagCfg),
		Force:       flagCfg.Force,
		IpUrl:       getIpUrl(),
		Logger:      logger,
	}

	return &App{config: config, appdirs: appDirs}, nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.appdirs.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	updater, err := mamupdater.NewMamUpdater(a.config)
	if err != nil {
		return fmt.Errorf("failed to create mam updater: %w", err)
	}

	if err := updater.Run(ctx); err != nil {
		return fmt.Errorf("mam update failed: %w", err)
	}
	return nil
}

func getAppDirs(flagCfg *flagConfig) (*appdir.AppDirs, error) {
	if flagCfg.ConfigDir != nil && *flagCfg.ConfigDir != "" {
		return appdir.New(*flagCfg.ConfigDir), nil
	}
	appDirs, err := appdir.NewFromAppName(".mamupdate")
	if err != nil {
		return nil, fmt.Errorf("failed to build app directories: %w", err)
	}
	return appDirs, nil
}

type flagConfig struct {
	MamId     *string
	ConfigDir *string
	Force     bool
	LogLevel  *string
}

func getFlags(args []string) (*flagConfig, error) {
	var mamId string
	flag.StringVar(&mamId, "mam-id", "", "Provide the mam-id used for the initial request.")
	var mamDir string
	flag.StringVar(&mamDir, "mam-dir", "", "Provide the directory where the config and data will be persisted.")
	var force bool
	flag.BoolVar(&force, "force", false, "Specify force to override the last run time.")
	var loglevel string
	flag.StringVar(&loglevel, "level", "", "Specify a log level (debug, info, warn, error) default: info.")

	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError) // Clear default flag set
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("unable to parse flags: %w", err)
	}

	return &flagConfig{
		MamId:     &mamId,
		ConfigDir: &mamDir,
		Force:     force,
		LogLevel:  &loglevel,
	}, nil
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

const (
	defaultIpUrl = "https://api.ipify.org"
)

func getIpUrl() string {
	envUrl := os.Getenv("IP_URL")
	if envUrl == "" {
		return defaultIpUrl
	}
	return envUrl
}

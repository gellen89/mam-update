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
	appdirs *appdir.AppDirs
	updater *mamupdater.MamUpdater
}

func Run(ctx context.Context, args []string) error {
	application, err := New(args)
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	if err := application.Run(ctx); err != nil {
		return fmt.Errorf("failed to run application: %w", err)
	}
	return nil
}

func New(args []string) (*App, error) {
	flagCfg, err := getFlags(args)
	if err != nil {
		return nil, fmt.Errorf("unable to parse flags: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(flagCfg),
	}))

	mamDir := getMamDir(flagCfg)

	appDirs, err := getAppDirs(mamDir)
	if err != nil {
		return nil, err
	}

	config := &mamupdater.Config{
		DataDir:        appDirs.Data,
		CookiePath:     filepath.Join(appDirs.Data, "MAM.cookie"),
		IpPath:         filepath.Join(appDirs.Data, "MAM.ip"),
		LastUpdatePath: filepath.Join(appDirs.Data, "last_update_time"),
		MamId:          getMamId(flagCfg),
		Force:          flagCfg.Force,
		IpUrl:          getIpUrl(),
		SeedboxUrl:     getDynSeedboxUrl(),
		Logger:         logger,
	}

	updater, err := mamupdater.NewMamUpdater(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create mam updater: %w", err)
	}

	return &App{appdirs: appDirs, updater: updater}, nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.appdirs.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	if err := a.updater.Run(ctx); err != nil {
		return fmt.Errorf("mam update failed: %w", err)
	}
	return nil
}

func getAppDirs(mamDir *string) (*appdir.AppDirs, error) {
	if mamDir != nil && *mamDir != "" {
		return appdir.New(*mamDir), nil
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
	flagSet := flag.NewFlagSet("mam-update", flag.ContinueOnError)

	var mamId string
	flagSet.StringVar(&mamId, "mam-id", "", "Provide the mam-id used for the initial request.")
	var mamDir string
	flagSet.StringVar(&mamDir, "mam-dir", "", "Provide the directory where the config and data will be persisted.")
	var force bool
	flagSet.BoolVar(&force, "force", false, "Specify force to override the last run time.")
	var loglevel string
	flagSet.StringVar(&loglevel, "level", "", "Specify a log level (debug, info, warn, error) default: info.")

	err := flagSet.Parse(args)
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
	envMamDir := os.Getenv("MAM_UPDATE_DIR")
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

const (
	defaultDynSeedboxUrl = "https://t.myanonamouse.net/json/dynamicSeedbox.php"
)

func getDynSeedboxUrl() string {
	envUrl := os.Getenv("MAM_SEEDBOX_URL")
	if envUrl == "" {
		return defaultDynSeedboxUrl
	}
	return envUrl
}

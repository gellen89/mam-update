package appdir

import (
	"os"
	"path/filepath"
)

// AppDirs holds paths for different types of application directories
type AppDirs struct {
	Config string
	Data   string
	Cache  string
}

// GetAppDirs returns platform-specific directory paths for the application
func GetAppDirs(appName string) (AppDirs, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return AppDirs{}, err
	}

	appDir := homeDir

	mamDir := os.Getenv("MAMUPDATE_DIR")
	if mamDir != "" {
		appDir = mamDir
	}

	return AppDirs{
		Config: filepath.Join(appDir, "."+appName),
		Data:   filepath.Join(appDir, "."+appName),
		Cache:  filepath.Join(appDir, "."+appName, "cache"),
	}, nil
}

// EnsureDirs creates all directories in the AppDirs struct if they don't exist
func (dirs AppDirs) EnsureDirs() error {
	for _, dir := range []string{dirs.Config, dirs.Data, dirs.Cache} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

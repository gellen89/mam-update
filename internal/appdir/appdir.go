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

	// switch runtime.GOOS {
	// case "darwin":
	// 	return AppDirs{
	// 		Config: filepath.Join(homeDir, "Library", "Preferences", appName),
	// 		Data:   filepath.Join(homeDir, "Library", "Application Support", appName),
	// 		Cache:  filepath.Join(homeDir, "Library", "Caches", appName),
	// 	}, nil

	// case "linux":
	// 	// Check XDG environment variables with fallbacks
	// 	configHome := os.Getenv("XDG_CONFIG_HOME")
	// 	if configHome == "" {
	// 		configHome = filepath.Join(homeDir, ".config")
	// 	}

	// 	dataHome := os.Getenv("XDG_DATA_HOME")
	// 	if dataHome == "" {
	// 		dataHome = filepath.Join(homeDir, ".local", "share")
	// 	}

	// 	cacheHome := os.Getenv("XDG_CACHE_HOME")
	// 	if cacheHome == "" {
	// 		cacheHome = filepath.Join(homeDir, ".cache")
	// 	}

	// 	return AppDirs{
	// 		Config: filepath.Join(configHome, appName),
	// 		Data:   filepath.Join(dataHome, appName),
	// 		Cache:  filepath.Join(cacheHome, appName),
	// 	}, nil

	// case "windows":
	// 	appData := os.Getenv("APPDATA")
	// 	if appData == "" {
	// 		appData = filepath.Join(homeDir, "AppData", "Roaming")
	// 	}
	// 	return AppDirs{
	// 		Config: filepath.Join(appData, appName),
	// 		Data:   filepath.Join(appData, appName),
	// 		Cache:  filepath.Join(appData, appName, "cache"),
	// 	}, nil

	// default:
	// 	// Fallback to dotfile pattern
	// 	return AppDirs{
	// 		Config: filepath.Join(homeDir, "."+appName),
	// 		Data:   filepath.Join(homeDir, "."+appName),
	// 		Cache:  filepath.Join(homeDir, "."+appName, "cache"),
	// 	}, nil
	// }

	return AppDirs{
		Config: filepath.Join(homeDir, "."+appName),
		Data:   filepath.Join(homeDir, "."+appName),
		Cache:  filepath.Join(homeDir, "."+appName, "cache"),
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

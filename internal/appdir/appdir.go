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

func GetAppDirs(appName string) (AppDirs, error) {
	userProvidedDir := os.Getenv("MAMUPDATE_DIR")
	if userProvidedDir != "" {
		return AppDirs{
			Config: userProvidedDir,
			Data:   userProvidedDir,
			Cache:  filepath.Join(userProvidedDir, ".cache"),
		}, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return AppDirs{}, err
	}

	appDir := filepath.Join(homeDir, appName)

	return AppDirs{
		Config: appDir,
		Data:   appDir,
		Cache:  filepath.Join(appDir, ".cache"),
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

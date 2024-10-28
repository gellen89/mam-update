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

// Returns app dirs with the given app dir
func New(appDir string) *AppDirs {
	return &AppDirs{
		Config: appDir,
		Data:   appDir,
		Cache:  filepath.Join(appDir, ".cache"),
	}
}

// Returns app dirs at the user's homedir + the appName as a subdirectory
func NewFromAppName(appName string) (*AppDirs, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(homeDir, appName)
	return New(appDir), nil
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

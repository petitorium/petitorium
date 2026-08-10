package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

// AppDir holds the resolved absolute path to the petitorium data directory.
// This is where workspaces, plugins, environments, and other data files live.
// It is set once during cobra.OnInitialize (before LoadConfig).
var AppDir string

// ConfigFile holds the resolved absolute path to the config file.
// Defaults to <AppDir>/config.yaml. It is set once during cobra.OnInitialize.
var ConfigFile string

// ResolveAppDir resolves the data directory from the --data-dir flag
// (or the OS-appropriate default when flagPath is empty).
//
// If flagPath is non-empty the path is expanded (~), made absolute, and
// validated: if it points to an existing file (not a directory) an error is
// returned. The directory itself is not created here — that is the caller's
// responsibility (see EnsureAppDir).
func ResolveAppDir(flagPath string) (string, error) {
	if flagPath != "" {
		expanded, err := homedir.Expand(flagPath)
		if err != nil {
			return "", fmt.Errorf("invalid data directory %q: %w", flagPath, err)
		}
		abs, err := filepath.Abs(expanded)
		if err != nil {
			return "", err
		}
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			return "", fmt.Errorf("data directory path must be a directory, not a file: %s", abs)
		}
		return abs, nil
	}
	// default: <os config dir>/petitorium
	// Linux: ~/.config/petitorium  macOS: ~/Library/Application Support/petitorium  Windows: %AppData%/petitorium
	configBase, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configBase, "petitorium"), nil
}

// ResolveConfigFile resolves the config file path from the --config flag.
// If flagPath is empty, it defaults to <AppDir>/config.yaml.
// AppDir must be set before calling this function.
//
// If flagPath is non-empty the path is expanded (~), made absolute, and
// validated: if it points to an existing directory (not a file) an error is
// returned. If the parent directory does not exist, an error is returned.
func ResolveConfigFile(flagPath string) (string, error) {
	if flagPath != "" {
		expanded, err := homedir.Expand(flagPath)
		if err != nil {
			return "", fmt.Errorf("invalid config file path %q: %w", flagPath, err)
		}
		abs, err := filepath.Abs(expanded)
		if err != nil {
			return "", err
		}
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return "", fmt.Errorf("config path must be a file, not a directory: %s", abs)
		}
		parent := filepath.Dir(abs)
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			return "", fmt.Errorf("config file directory does not exist: %s", parent)
		}
		return abs, nil
	}
	if AppDir == "" {
		return "", fmt.Errorf("AppDir is not set; call ResolveAppDir first")
	}
	return filepath.Join(AppDir, "config.yaml"), nil
}

// EnsureAppDir creates the AppDir directory if it does not already exist.
// It should be called after ResolveAppDir has set AppDir.
func EnsureAppDir() error {
	if AppDir == "" {
		return fmt.Errorf("AppDir is not set; call ResolveAppDir first")
	}
	if info, err := os.Stat(AppDir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("data directory path must be a directory, not a file: %s", AppDir)
		}
		return nil
	}
	return os.MkdirAll(AppDir, 0o755)
}

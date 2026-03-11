package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/petitorium/petitorium-plugin-sdk/types"
)

// RegistryClient handles communication with the plugin registry
type RegistryClient struct {
	BaseURL string
}

// NewRegistryClient creates a new RegistryClient
func NewRegistryClient(baseURL string) *RegistryClient {
	if baseURL == "" {
		// Default to mock registry during development if config is empty
		baseURL = "http://localhost:8080/api/v1"
	}
	return &RegistryClient{BaseURL: baseURL}
}

// ListPlugins fetches the list of available plugins from the registry
func (rc *RegistryClient) ListPlugins() ([]types.RegistryPlugin, error) {
	resp, err := http.Get(fmt.Sprintf("%s/plugins", rc.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %s", resp.Status)
	}

	var plugins []types.RegistryPlugin
	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, fmt.Errorf("failed to decode plugins: %w", err)
	}

	return plugins, nil
}

// InstallPlugin downloads and installs a plugin
func (pm *PluginManager) InstallPlugin(p types.RegistryPlugin) error {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	downloadURL := ""
	expectedChecksum := ""

	for _, r := range p.Releases {
		if r.Platform == platform {
			downloadURL = r.URL
			expectedChecksum = r.Checksum
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("plugin %s not available for platform %s", p.Name, platform)
	}

	if expectedChecksum == "" {
		return fmt.Errorf("checksum not available for plugin %s on platform %s", p.Name, platform)
	}

	// Create download path
	pluginFile := p.Name + ".so"
	destPath := filepath.Join(pm.pluginDir, pluginFile)

	// Download file
	if err := downloadFile(downloadURL, destPath, expectedChecksum); err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}

	// Update configuration
	if pm.config.Installed == nil {
		pm.config.Installed = make(map[string]InstalledInfo)
	}
	pm.config.Installed[p.Name] = InstalledInfo{
		Version:  p.Version,
		Checksum: expectedChecksum,
		Path:     destPath,
	}

	return nil
}

// UninstallPlugin removes an installed plugin
func (pm *PluginManager) UninstallPlugin(name string) error {
	info, ok := pm.config.Installed[name]
	if !ok {
		return fmt.Errorf("plugin %s is not installed", name)
	}

	// Remove file
	if err := os.Remove(info.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plugin file: %w", err)
	}

	// Update config
	delete(pm.config.Installed, name)
	pm.DisablePlugin(name)

	return nil
}

// downloadFile downloads a file and verifies its checksum
func downloadFile(url string, destPath string, expectedChecksum string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Verify checksum
	return verifyChecksum(destPath, expectedChecksum)
}

// verifyChecksum checks if the file at path matches the expected SHA256 checksum
func verifyChecksum(path string, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

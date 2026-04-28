// Package plugins provides the plugin management functionality.
package plugins

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/petitorium/petitorium-plugin-sdk/shared"
	"github.com/petitorium/petitorium-plugin-sdk/types"
)

// PluginManager manages the loading and execution of plugins
type PluginManager struct {
	plugins   map[string]Plugin
	hooks     map[HookType][]Plugin
	config    *PluginConfig
	pluginDir string
	clients   []*plugin.Client
	baseURL   string
}

// NewPluginManager creates a new PluginManager instance
func NewPluginManager(config *PluginConfig, pluginDir string) *PluginManager {
	baseURL := config.RegistryURL
	if baseURL == "" {
		baseURL = "https://hub.petitorium.dev/api/v1"
	}
	return &PluginManager{
		plugins:   make(map[string]Plugin),
		hooks:     make(map[HookType][]Plugin),
		config:    config,
		pluginDir: pluginDir,
		clients:   make([]*plugin.Client, 0),
		baseURL:   baseURL,
	}
}

// Close kills all running plugin processes
func (pm *PluginManager) Close() {
	for _, client := range pm.clients {
		client.Kill()
	}
}

// UnloadPlugin unloads a specific plugin by killing its client and removing it from the registry
func (pm *PluginManager) UnloadPlugin(name string) {
	// Remove from plugins map
	delete(pm.plugins, name)

	// Remove from hooks map
	for hookType, hookList := range pm.hooks {
		newList := []Plugin{}
		for _, p := range hookList {
			if p.Name() != name {
				newList = append(newList, p)
			}
		}
		pm.hooks[hookType] = newList
	}

	// Kill all clients - this is the safest approach since we can't identify
	// which client belongs to which plugin without additional tracking.
	// After killing, the plugin will be reloaded by LoadPlugins if still enabled.
	for _, client := range pm.clients {
		client.Kill()
	}
	pm.clients = []*plugin.Client{}
}

// RegisterPlugin registers a plugin with the manager
func (pm *PluginManager) RegisterPlugin(p Plugin) error {
	name := p.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}
	pm.plugins[name] = p

	// Register hooks
	for _, hookType := range p.Hooks() {
		if pm.hooks[hookType] == nil {
			pm.hooks[hookType] = []Plugin{}
		}
		pm.hooks[hookType] = append(pm.hooks[hookType], p)
	}

	return nil
}

// ExecuteHooks executes all hooks of the given type with timeout and error isolation
func (pm *PluginManager) ExecuteHooks(hookType HookType, ctx *HookContext) error {
	plugins, exists := pm.hooks[hookType]
	if !exists {
		return nil // No hooks for this type
	}

	for _, p := range plugins {
		done := make(chan error, 1)
		var updatedCtx *types.HookContext
		go func(plg Plugin) {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Errorf("hook panicked: %v", r)
				}
			}()
			var err error
			updatedCtx, err = plg.ExecuteHook(hookType, ctx)
			done <- err
		}(p)

		select {
		case err := <-done:
			if err == nil && updatedCtx != nil {
				// Update context for next plugin in chain
				*ctx = *updatedCtx
			}
		case <-time.After(5 * time.Second):
			// Timeout, continue
		}
	}
	return nil
}

// LoadPlugin loads a single plugin by name
func (pm *PluginManager) LoadPlugin(name string) error {
	pluginPath := filepath.Join(pm.pluginDir, name)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s not found in %s", name, pm.pluginDir)
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			name: &shared.PetitoriumPlugin{},
		},
		Cmd:              exec.Command(pluginPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger: hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Output: io.Discard,
			Level:  hclog.Error,
		}),
		SyncStdout: io.Discard,
		SyncStderr: io.Discard,
	})
	pm.clients = append(pm.clients, client)

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to connect to plugin %s: %w", name, err)
	}

	raw, err := rpcClient.Dispense(name)
	if err != nil {
		return fmt.Errorf("failed to dispense plugin %s: %w", name, err)
	}

	plg, ok := raw.(Plugin)
	if !ok {
		return fmt.Errorf("plugin %s does not implement Plugin interface", name)
	}

	return pm.RegisterPlugin(plg)
}

// LoadPlugins loads all enabled plugins from the plugin directory
func (pm *PluginManager) LoadPlugins() error {
	for _, name := range pm.config.Enabled {
		if err := pm.LoadPlugin(name); err != nil {
			// Log error and continue
			fmt.Printf("Warning: failed to load plugin %s: %v\n", name, err)
		}
	}
	return nil
}

// EnablePlugin enables a plugin by name
func (pm *PluginManager) EnablePlugin(name string) error {
	pluginPath := filepath.Join(pm.pluginDir, name)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s not found", name)
	}
	// Add to enabled list if not already
	for _, enabled := range pm.config.Enabled {
		if enabled == name {
			return nil // Already enabled
		}
	}
	pm.config.Enabled = append(pm.config.Enabled, name)
	return nil
}

// IsPluginInstalled checks if a plugin is installed
func (pm *PluginManager) IsPluginInstalled(name string) bool {
	if pm.config.Installed == nil {
		return false
	}
	_, installed := pm.config.Installed[name]
	return installed
}

// GetInstalledInfo returns the installation info for a plugin
func (pm *PluginManager) GetInstalledInfo(name string) (InstalledInfo, bool) {
	if pm.config.Installed == nil {
		return InstalledInfo{}, false
	}
	info, installed := pm.config.Installed[name]
	return info, installed
}

// DisablePlugin disables a plugin by name
func (pm *PluginManager) DisablePlugin(name string) error {
	for i, enabled := range pm.config.Enabled {
		if enabled == name {
			pm.config.Enabled = append(pm.config.Enabled[:i], pm.config.Enabled[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("plugin %s not enabled", name)
}

// GetEnabledPlugins returns the list of enabled plugin names
func (pm *PluginManager) GetEnabledPlugins() []string {
	return pm.config.Enabled
}

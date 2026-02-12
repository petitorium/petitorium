// Package plugins provides the plugin management functionality.
package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"time"
)

// PluginManager manages the loading and execution of plugins
type PluginManager struct {
	plugins   map[string]Plugin
	hooks     map[HookType][]PluginHook
	config    *PluginConfig
	pluginDir string
}

// NewPluginManager creates a new PluginManager instance
func NewPluginManager(config *PluginConfig, pluginDir string) *PluginManager {
	return &PluginManager{
		plugins:   make(map[string]Plugin),
		hooks:     make(map[HookType][]PluginHook),
		config:    config,
		pluginDir: pluginDir,
	}
}

// RegisterPlugin registers a plugin with the manager
func (pm *PluginManager) RegisterPlugin(p Plugin) error {
	name := p.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}
	pm.plugins[name] = p

	// Register hooks
	hookFuncs := p.HookFuncs()
	for _, hookType := range p.Hooks() {
		if pm.hooks[hookType] == nil {
			pm.hooks[hookType] = []PluginHook{}
		}
		if hookFunc, exists := hookFuncs[hookType]; exists {
			pm.hooks[hookType] = append(pm.hooks[hookType], hookFunc)
		}
	}

	return nil
}

// ExecuteHooks executes all hooks of the given type with timeout and error isolation
func (pm *PluginManager) ExecuteHooks(hookType HookType, ctx *HookContext) error {
	hooks, exists := pm.hooks[hookType]
	if !exists {
		return nil // No hooks for this type
	}

	for _, hook := range hooks {
		done := make(chan error, 1)
		go func(h PluginHook) {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Errorf("hook panicked: %v", r)
				}
			}()
			done <- h(ctx)
		}(hook)

		select {
		case err := <-done:
			if err != nil {
				// Log error but continue for graceful degradation
				// For now, just continue
			}
		case <-time.After(5 * time.Second):
			// Timeout, continue
		}
	}
	return nil
}

// LoadPlugin loads a single plugin by name
func (pm *PluginManager) LoadPlugin(name string) error {
	pluginPath := filepath.Join(pm.pluginDir, name+".so")
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", name, err)
	}

	sym, err := p.Lookup("Plugin")
	if err != nil {
		qualifiedName := fmt.Sprintf("github.com/petitorium/petitorium/plugins/examples/%s.Plugin", name)
		sym, err = p.Lookup(qualifiedName)
		if err != nil {
			return fmt.Errorf("failed to lookup Plugin symbol in %s: %w", name, err)
		}
	}

	plg, ok := sym.(Plugin)
	if !ok {
		if ptr, ok := sym.(*Plugin); ok {
			plg = *ptr
		} else {
			return fmt.Errorf("plugin %s does not implement Plugin interface", name)
		}
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
	// Check if plugin file exists
	pluginPath := filepath.Join(pm.pluginDir, name+".so")
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

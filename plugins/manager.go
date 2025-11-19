// Package plugins provides the plugin management functionality.
package plugins

import (
	"fmt"
)

// PluginManager manages the loading and execution of plugins
type PluginManager struct {
	plugins map[string]Plugin
	hooks   map[HookType][]PluginHook
	config  *PluginConfig
}

// NewPluginManager creates a new PluginManager instance
func NewPluginManager(config *PluginConfig) *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		hooks:   make(map[HookType][]PluginHook),
		config:  config,
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
	for _, hookType := range p.Hooks() {
		if pm.hooks[hookType] == nil {
			pm.hooks[hookType] = []PluginHook{}
		}
		// Note: In a real implementation, the plugin would provide the hook functions
		// For now, this is a placeholder
	}

	return nil
}

// ExecuteHooks executes all hooks of the given type
func (pm *PluginManager) ExecuteHooks(hookType HookType, ctx *HookContext) error {
	hooks, exists := pm.hooks[hookType]
	if !exists {
		return nil // No hooks for this type
	}

	for _, hook := range hooks {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("hook execution failed: %w", err)
		}
	}
	return nil
}

// LoadPlugins loads plugins from the configured directory
func (pm *PluginManager) LoadPlugins() error {
	// Placeholder implementation
	// In a real implementation, this would scan the plugin directory
	// and load .so files using Go's plugin package
	for _, name := range pm.config.Enabled {
		// Simulate loading
		// p, err := plugin.Open(filepath.Join(pluginDir, name+".so"))
		// if err != nil {
		//     return err
		// }
		// sym, err := p.Lookup("Plugin")
		// if err != nil {
		//     return err
		// }
		// plg := sym.(Plugin)
		// pm.RegisterPlugin(plg)
		_ = name // Placeholder
	}
	return nil
}

// EnablePlugin enables a plugin by name
func (pm *PluginManager) EnablePlugin(name string) error {
	if _, exists := pm.plugins[name]; !exists {
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

// Package main provides an example auth injector plugin.
// To build: go build -buildmode=plugin -o auth-injector.so .
package main

import (
	"github.com/hbarral/petitorium/plugins"
)

// AuthInjector is an example plugin that injects auth headers
type AuthInjector struct{}

// Name returns the plugin name
func (ai *AuthInjector) Name() string {
	return "auth-injector"
}

// Version returns the plugin version
func (ai *AuthInjector) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (ai *AuthInjector) Description() string {
	return "Injects authentication headers into requests"
}

// Hooks returns the hook types this plugin implements
func (ai *AuthInjector) Hooks() []plugins.HookType {
	return []plugins.HookType{plugins.PreSend}
}

// HookFuncs returns the hook functions
func (ai *AuthInjector) HookFuncs() map[plugins.HookType]plugins.PluginHook {
	return map[plugins.HookType]plugins.PluginHook{
		plugins.PreSend: ai.injectAuth,
	}
}

// injectAuth adds auth header
func (ai *AuthInjector) injectAuth(ctx *plugins.HookContext) error {
	// Example: add Bearer token
	ctx.Request.Headers["Authorization"] = "Bearer example-token"
	return nil
}

// Plugin is the exported plugin instance
var Plugin plugins.Plugin = &AuthInjector{}

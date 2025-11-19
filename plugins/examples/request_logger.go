// Package main provides an example request logger plugin.
// To build: go build -buildmode=plugin -o request-logger.so .
package main

import (
	"fmt"

	"github.com/hbarral/petitorium/plugins"
)

// RequestLogger is an example plugin that logs requests and responses
type RequestLogger struct{}

// Name returns the plugin name
func (rl *RequestLogger) Name() string {
	return "request-logger"
}

// Version returns the plugin version
func (rl *RequestLogger) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (rl *RequestLogger) Description() string {
	return "Logs outgoing requests and incoming responses"
}

// Hooks returns the hook types this plugin implements
func (rl *RequestLogger) Hooks() []plugins.HookType {
	return []plugins.HookType{plugins.PreRequest, plugins.PostReceive}
}

// HookFuncs returns the hook functions
func (rl *RequestLogger) HookFuncs() map[plugins.HookType]plugins.PluginHook {
	return map[plugins.HookType]plugins.PluginHook{
		plugins.PreRequest:  rl.logRequest,
		plugins.PostReceive: rl.logResponse,
	}
}

// logRequest logs the outgoing request
func (rl *RequestLogger) logRequest(ctx *plugins.HookContext) error {
	req := ctx.Request
	fmt.Printf("[REQUEST] %s %s\n", req.Method, req.URL)
	return nil
}

// logResponse logs the incoming response
func (rl *RequestLogger) logResponse(ctx *plugins.HookContext) error {
	// Since Response is interface{}, we need to type assert
	// Assuming it's *cmd.HTTPResponse, but for plugin, we can't import cmd
	// For example purposes, we'll skip detailed logging
	fmt.Println("[RESPONSE] Received response")
	return nil
}

// Plugin is the exported plugin instance
var Plugin plugins.Plugin = &RequestLogger{}

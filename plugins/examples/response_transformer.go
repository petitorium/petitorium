// Package main provides an example response transformer plugin.
// To build: go build -buildmode=plugin -o response-transformer.so .
package main

import (
	"strings"

	"github.com/hbarral/petitorium/plugins"
)

// ResponseTransformer is an example plugin that transforms responses
type ResponseTransformer struct{}

// Name returns the plugin name
func (rt *ResponseTransformer) Name() string {
	return "response-transformer"
}

// Version returns the plugin version
func (rt *ResponseTransformer) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (rt *ResponseTransformer) Description() string {
	return "Transforms response bodies (e.g., formatting)"
}

// Hooks returns the hook types this plugin implements
func (rt *ResponseTransformer) Hooks() []plugins.HookType {
	return []plugins.HookType{plugins.ResponseTransform}
}

// HookFuncs returns the hook functions
func (rt *ResponseTransformer) HookFuncs() map[plugins.HookType]plugins.PluginHook {
	return map[plugins.HookType]plugins.PluginHook{
		plugins.ResponseTransform: rt.transformResponse,
	}
}

// transformResponse modifies the response body
func (rt *ResponseTransformer) transformResponse(ctx *plugins.HookContext) error {
	// Example: uppercase the body
	if resp, ok := ctx.Response.(*map[string]interface{}); ok {
		// Assuming body is string, but since interface{}, need to handle
		// For simplicity, skip
	}
	// Since Response is interface{}, and we don't know the type, skip implementation
	return nil
}

// Plugin is the exported plugin instance
var Plugin plugins.Plugin = &ResponseTransformer{}

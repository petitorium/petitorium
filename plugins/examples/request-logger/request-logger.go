// Package main provides an example request logger plugin.
// To build: go build -buildmode=plugin -o request-logger.so .
package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/mitchellh/go-homedir"

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
	return []plugins.HookType{plugins.PreRequest, plugins.PostVariableSubstitution, plugins.PostReceive}
}

// HookFuncs returns the hook functions
func (rl *RequestLogger) HookFuncs() map[plugins.HookType]plugins.PluginHook {
	return map[plugins.HookType]plugins.PluginHook{
		plugins.PreRequest:               rl.logRawRequest,
		plugins.PostVariableSubstitution: rl.logExpandedRequest,
		plugins.PostReceive:              rl.logResponse,
	}
}

// logRawRequest logs the outgoing request with raw template variables
func (rl *RequestLogger) logRawRequest(ctx *plugins.HookContext) error {
	req := ctx.Request
	logEntry := fmt.Sprintf("[%s] [REQUEST] [RAW] %s %s\n", time.Now().Format(time.RFC3339), req.Method, req.URL)

	// Add headers if present
	if len(req.Headers) > 0 {
		logEntry += "Headers:\n"
		for key, value := range req.Headers {
			logEntry += fmt.Sprintf("  %s: %s\n", key, value)
		}
	}

	// Add body if present
	if req.Body != "" {
		logEntry += fmt.Sprintf("Body:\n%s\n", req.Body)
	}

	logEntry += "\n"
	return rl.writeToLog(logEntry, ctx)
}

// logExpandedRequest logs the outgoing request with expanded environment variables
func (rl *RequestLogger) logExpandedRequest(ctx *plugins.HookContext) error {
	req := ctx.Request
	logEntry := fmt.Sprintf("[%s] [REQUEST] [EXPANDED] %s %s\n", time.Now().Format(time.RFC3339), req.Method, req.URL)

	// Add headers if present
	if len(req.Headers) > 0 {
		logEntry += "Headers:\n"
		for key, value := range req.Headers {
			logEntry += fmt.Sprintf("  %s: %s\n", key, value)
		}
	}

	// Add body if present
	if req.Body != "" {
		logEntry += fmt.Sprintf("Body:\n%s\n", req.Body)
	}

	logEntry += "\n"
	return rl.writeToLog(logEntry, ctx)
}

// logResponse logs the incoming response
func (rl *RequestLogger) logResponse(ctx *plugins.HookContext) error {
	logEntry := fmt.Sprintf("[%s] [RESPONSE]\n", time.Now().Format(time.RFC3339))

	// Try to get response details if available
	if ctx.Response != nil {
		// Use reflection to safely access response fields
		respVal := reflect.ValueOf(ctx.Response)
		if respVal.Kind() == reflect.Ptr {
			respVal = respVal.Elem()
		}

		if respVal.IsValid() && respVal.Kind() == reflect.Struct {
			// Try to get StatusCode field
			if statusField := respVal.FieldByName("StatusCode"); statusField.IsValid() {
				if statusField.Kind() == reflect.Int {
					logEntry += fmt.Sprintf("Status: %d\n", statusField.Int())
				}
			}

			// Try to get Status field
			if statusField := respVal.FieldByName("Status"); statusField.IsValid() {
				if statusField.Kind() == reflect.String {
					logEntry += fmt.Sprintf("Status Text: %s\n", statusField.String())
				}
			}

			// Try to get Body field
			if bodyField := respVal.FieldByName("Body"); bodyField.IsValid() {
				if bodyField.Kind() == reflect.String {
					body := bodyField.String()
					if body != "" {
						logEntry += fmt.Sprintf("Body:\n%s\n", body)
					}
				}
			}
		}
	} else {
		logEntry += "No response data available\n"
	}

	logEntry += "\n"
	return rl.writeToLog(logEntry, ctx)
}

// writeToLog writes a log entry to the log file
func (rl *RequestLogger) writeToLog(entry string, ctx *plugins.HookContext) error {
	// Get log file path from context config, default to "petitorium.log" if not specified
	logFile := "petitorium.log"
	if ctx.Config != nil {
		if configLogFile, ok := ctx.Config["logFile"].(string); ok && configLogFile != "" {
			logFile = configLogFile
		}
	}

	// Expand home directory if path contains ~
	if expandedPath, err := homedir.Expand(logFile); err == nil {
		logFile = expandedPath
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}

// Plugin is the exported plugin instance
var Plugin plugins.Plugin = &RequestLogger{}

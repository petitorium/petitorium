// Package plugins provides the core interfaces and types for the Petitorium plugin ecosystem.
package plugins

import (
	"net/http"
	"time"
)

// Plugin represents a loaded plugin
type Plugin interface {
	Name() string
	Version() string
	Description() string
	Hooks() []HookType
	HookFuncs() map[HookType]PluginHook
}

// HookType defines when a plugin hook executes
type HookType string

const (
	PreRequest               HookType = "pre_request"
	PostRequest              HookType = "post_request"
	PostReceive              HookType = "post_receive"
	PreSend                  HookType = "pre_send"
	PostSend                 HookType = "post_send"
	RequestValidation        HookType = "request_validation"
	ResponseValidation       HookType = "response_validation"
	PreVariableSubstitution  HookType = "pre_variable_substitution"
	PostVariableSubstitution HookType = "post_variable_substitution"
	PreSave                  HookType = "pre_save"
	PostSave                 HookType = "post_save"
	PreUIUpdate              HookType = "pre_ui_update"
	PostUIUpdate             HookType = "post_ui_update"
	OnUIInit                 HookType = "on_ui_init"
	OnUIClose                HookType = "on_ui_close"
	OnCollectionLoad         HookType = "on_collection_load"
	OnCollectionSave         HookType = "on_collection_save"
	OnEnvironmentLoad        HookType = "on_environment_load"
	OnEnvironmentSave        HookType = "on_environment_save"
	OnConfigLoad             HookType = "on_config_load"
	OnConfigSave             HookType = "on_config_save"
	OnError                  HookType = "on_error"
	OnSuccess                HookType = "on_success"
	RequestRetry             HookType = "request_retry"
	RequestTimeout           HookType = "request_timeout"
	ResponseTransform        HookType = "response_transform"
	ResponseCache            HookType = "response_cache"
)

// HookContext provides context data to plugin hooks
type HookContext struct {
	Request     *RequestData
	Response    *HTTPResponse
	Environment map[string]string
	Config      map[string]interface{}
}

// RequestData represents the request being processed
type RequestData struct {
	Method      string
	URL         string
	Headers     map[string]string
	Body        string
	Collection  string
	RequestName string
}

// HTTPResponse represents the response from an HTTP request
type HTTPResponse struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Cookies    []*http.Cookie
	Body       string
	Duration   time.Duration
	Timestamp  time.Time
	BodySize   int
}

// PluginHook defines the function signature for plugin hooks
type PluginHook func(ctx *HookContext) error

// PluginConfig holds configuration for plugins
type PluginConfig struct {
	Enabled []string               `yaml:"enabled"`
	Config  map[string]interface{} `yaml:"config,omitempty"`
}

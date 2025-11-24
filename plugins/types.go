// Package plugins provides the core interfaces and types for the Petitorium plugin ecosystem.
// This package now uses type aliases to the shared plugin SDK for better compatibility.
package plugins

import (
	"github.com/petitorium/petitorium-plugin-sdk/types"
)

// Type aliases for backward compatibility
type Plugin = types.Plugin

type HookType = types.HookType

type HookContext = types.HookContext

type RequestData = types.RequestData

type PluginHook = types.PluginHook

type PluginConfig = types.PluginConfig

type ResponseData = types.ResponseData

// Re-export all hook type constants
const (
	PreRequest               = types.PreRequest
	PostRequest              = types.PostRequest
	PostReceive              = types.PostReceive
	PreSend                  = types.PreSend
	PostSend                 = types.PostSend
	RequestValidation        = types.RequestValidation
	ResponseValidation       = types.ResponseValidation
	PreVariableSubstitution  = types.PreVariableSubstitution
	PostVariableSubstitution = types.PostVariableSubstitution
	PreSave                  = types.PreSave
	PostSave                 = types.PostSave
	PreUIUpdate              = types.PreUIUpdate
	PostUIUpdate             = types.PostUIUpdate
	OnUIInit                 = types.OnUIInit
	OnUIClose                = types.OnUIClose
	OnCollectionLoad         = types.OnCollectionLoad
	OnCollectionSave         = types.OnCollectionSave
	OnEnvironmentLoad        = types.OnEnvironmentLoad
	OnEnvironmentSave        = types.OnEnvironmentSave
	OnConfigLoad             = types.OnConfigLoad
	OnConfigSave             = types.OnConfigSave
	OnError                  = types.OnError
	OnSuccess                = types.OnSuccess
	RequestRetry             = types.RequestRetry
	RequestTimeout           = types.RequestTimeout
	ResponseTransform        = types.ResponseTransform
	ResponseCache            = types.ResponseCache
)

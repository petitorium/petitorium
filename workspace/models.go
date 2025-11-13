// Package workspace provides models for managing collections of HTTP requests and responses.
package workspace

import (
	"time"
)

// HTTPResponse represents a stored HTTP response
// This is a simplified version for storage purposes
type HTTPResponse struct {
	StatusCode int                 `yaml:"status_code"`
	Status     string              `yaml:"status"`
	Headers    map[string][]string `yaml:"headers,omitempty"`
	Body       string              `yaml:"body,omitempty"`
	Duration   time.Duration       `yaml:"duration"`
	Timestamp  time.Time           `yaml:"timestamp"`
}

type Request struct {
	Name            string            `yaml:"name"`
	Method          string            `yaml:"method"`
	URL             string            `yaml:"url"`
	Headers         map[string]string `yaml:"headers,omitempty"`
	Body            string            `yaml:"body,omitempty"`
	ResponseHistory []HTTPResponse    `yaml:"response_history,omitempty"`
}

type Environment struct {
	Name      string            `yaml:"name"`
	Base      string            `yaml:"base,omitempty"` // Name of base environment to inherit from
	Variables map[string]string `yaml:"variables,omitempty"`
}

type Collection struct {
	Name        string       `yaml:"name"`
	Requests    []Request    `yaml:"requests,omitempty"`
	Collections []Collection `yaml:"collections,omitempty"`
	Expanded    bool         `yaml:"expanded,omitempty"`
}

type Workspace struct {
	Name         string        `yaml:"name"`
	Description  string        `yaml:"description,omitempty"`
	Collections  []Collection  `yaml:"collections,omitempty"`
	Environments []Environment `yaml:"environments,omitempty"`
	CreatedAt    time.Time     `yaml:"created_at"`
	UpdatedAt    time.Time     `yaml:"updated_at"`
}

type WorkspaceMetadata struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description,omitempty"`
	CreatedAt   time.Time `yaml:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at"`
}

type WorkspaceManager struct {
	CurrentWorkspace string              `yaml:"current_workspace"`
	Workspaces       []WorkspaceMetadata `yaml:"workspaces"`
}

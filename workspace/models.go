// Package workspace provides models for managing collections of HTTP requests and responses.
package workspace

import "time"

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

type Collection struct {
	Name        string       `yaml:"name"`
	Requests    []Request    `yaml:"requests,omitempty"`
	Collections []Collection `yaml:"collections,omitempty"`
	Expanded    bool         `yaml:"expanded,omitempty"`
}

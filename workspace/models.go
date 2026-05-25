// Package workspace provides models for managing collections of HTTP requests and responses.
package workspace

import (
	"net/http"
	"time"

	"gopkg.in/yaml.v3"
)

var HTTPMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodHead,
	http.MethodOptions,
}

type Cookie struct {
	Name     string `yaml:"name"`
	Value    string `yaml:"value"`
	Domain   string `yaml:"domain,omitempty"`
	Path     string `yaml:"path,omitempty"`
	Expires  string `yaml:"expires,omitempty"`
	Secure   bool   `yaml:"secure"`
	HttpOnly bool   `yaml:"http_only"`
	SameSite string `yaml:"same_site,omitempty"`
	Enabled  bool   `yaml:"enabled"`
}

type CookieJar struct {
	Cookies []Cookie `yaml:"cookies,omitempty"`
}

// HTTPResponse represents a stored HTTP response
// This is a simplified version for storage purposes
type HTTPResponse struct {
	StatusCode int                 `yaml:"status_code"`
	Status     string              `yaml:"status"`
	Headers    map[string][]string `yaml:"headers,omitempty"`
	Cookies    []Cookie            `yaml:"cookies,omitempty"`
	Body       string              `yaml:"body,omitempty"`
	Duration   time.Duration       `yaml:"duration"`
	Timestamp  time.Time           `yaml:"timestamp"`
}

// Entry represents a key-value pair with enabled state for headers and query params
type Entry struct {
	Value   string `yaml:"value"`
	Enabled bool   `yaml:"enabled"`
}

// UnmarshalYAML implements custom unmarshaling for Entry
// It handles both legacy string format and new Entry format
func (e *Entry) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		// Legacy format: just a string value
		e.Value = node.Value
		e.Enabled = true
		return nil
	}

	// New format: map with value and enabled fields
	var legacy struct {
		Value   string `yaml:"value"`
		Enabled bool   `yaml:"enabled"`
	}
	if err := node.Decode(&legacy); err != nil {
		return err
	}
	e.Value = legacy.Value
	e.Enabled = legacy.Enabled
	return nil
}

type Request struct {
	Name            string           `yaml:"name"`
	Method          string           `yaml:"method"`
	URL             string           `yaml:"url"`
	QueryParams     map[string]Entry `yaml:"query_params,omitempty"`
	Headers         map[string]Entry `yaml:"headers,omitempty"`
	Cookies         []Cookie         `yaml:"cookies,omitempty"`
	ContentType     string           `yaml:"content_type,omitempty"`
	Body            string           `yaml:"body,omitempty"`
	ResponseHistory []HTTPResponse   `yaml:"response_history,omitempty"`
}

type BodyContent struct {
	Raw       string           `yaml:"raw,omitempty"` // For JSON content
	Multipart []MultipartField `yaml:"multipart,omitempty"`
}

type MultipartField struct {
	Name     string `yaml:"name"`               // Field name
	Type     string `yaml:"type"`               // "text", "text_multiline", or "file"
	Value    string `yaml:"value"`              // Text content or file path
	Filename string `yaml:"filename,omitempty"` // Optional filename for file fields
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
	Name                string        `yaml:"name"`
	Description         string        `yaml:"description,omitempty"`
	CookieJar           CookieJar     `yaml:"cookie_jar,omitempty"`
	Collections         []Collection  `yaml:"collections,omitempty"`
	Environments        []Environment `yaml:"environments,omitempty"`
	SelectedEnvironment string        `yaml:"selected_environment,omitempty"`
	SelectedRequest     []string      `yaml:"selected_request,omitempty"`
	CreatedAt           time.Time     `yaml:"created_at"`
	UpdatedAt           time.Time     `yaml:"updated_at"`
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

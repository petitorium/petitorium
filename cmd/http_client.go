package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
)

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

// SendRequest sends an HTTP request with the given parameters
func SendRequest(method, url, body string, headers map[string]string) (*HTTPResponse, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request body
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}

	// Create HTTP request
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default Content-Type for methods that typically have a body
	if body != "" && req.Header.Get("Content-Type") == "" {
		switch method {
		case "POST", "PUT", "PATCH":
			// Try to detect content type
			if strings.TrimSpace(body) != "" {
				if strings.HasPrefix(strings.TrimSpace(body), "{") && strings.HasSuffix(strings.TrimSpace(body), "}") {
					req.Header.Set("Content-Type", "application/json")
				} else if strings.HasPrefix(strings.TrimSpace(body), "<") && strings.Contains(body, ">") {
					req.Header.Set("Content-Type", "application/xml")
				} else {
					req.Header.Set("Content-Type", "text/plain")
				}
			}
		}
	}

	// Set User-Agent if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Petitorium/1.0")
	}

	// Execute request and measure duration
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Convert headers to map for easier display
	headersMap := make(map[string][]string)
	for key, values := range resp.Header {
		headersMap[key] = values
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    headersMap,
		Cookies:    resp.Cookies(),
		Body:       string(respBody),
		Duration:   duration,
		Timestamp:  time.Now(),
		BodySize:   len(respBody),
	}, nil
}

// FormatResponse formats the HTTP response for display in the UI
func FormatResponse(response *HTTPResponse) string {
	var result strings.Builder

	// Status line - hidden for now
	// result.WriteString(fmt.Sprintf("Status: %d %s\n", response.StatusCode, response.Status))
	// result.WriteString(fmt.Sprintf("Time: %v\n\n", response.Duration.Round(time.Millisecond)))

	// Headers - hidden for now
	// if len(response.Headers) > 0 {
	// 	result.WriteString("Headers:\n")
	// 	for key, values := range response.Headers {
	// 		for _, value := range values {
	// 			result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
	// 		}
	// 	}
	// 	result.WriteString("\n")
	// }

	// Body
	if response.Body != "" {
		// Pretty-print JSON if it's valid JSON
		bodyToFormat := response.Body
		if strings.TrimSpace(response.Body)[0] == '{' || strings.TrimSpace(response.Body)[0] == '[' {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(response.Body), &jsonData); err == nil {
				// It's valid JSON, pretty-print it
				prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
				if err == nil {
					bodyToFormat = string(prettyJSON)
				}
			}
		}

		formattedBody := formatBodyContent(bodyToFormat)
		result.WriteString(formattedBody)
	} else {
		result.WriteString("(empty response)")
	}

	return result.String()
}

// GetCurrentRequestData extracts the current request data from UI state
func GetCurrentRequestData(methodDropdown *tview.DropDown, urlInput *VariableURLInput, currentRequest *workspace.Request) (string, string, string, map[string]string) {
	// Get method from dropdown
	_, method := methodDropdown.GetCurrentOption()

	// Get URL from input field
	url := urlInput.GetText()

	// Get body from current request
	body := ""
	if currentRequest != nil {
		body = currentRequest.Body
	}

	// Get headers from UI
	headers := getHeadersFromUI()

	return method, url, body, headers
}

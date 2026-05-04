package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

// HTTPResponse represents the response from an HTTP request
type HTTPResponse struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Cookies    []*http.Cookie
	Body       string
	BodyBytes  []byte
	Duration   time.Duration
	Timestamp  time.Time
	BodySize   int
}

// SendRequest sends an HTTP request with the given parameters
func SendRequest(method, urlStr, body string, contentType string, headers map[string]string, queryParams map[string]string) (*HTTPResponse, error) {
	// Append query parameters to URL
	if len(queryParams) > 0 {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %v", err)
		}
		q := u.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
		urlStr = u.String()
	}

	// Create HTTP client with configurable timeout
	timeoutVal := config.C.RequestTimeout
	if timeoutVal <= 0 {
		timeoutVal = 60 // Default to 60 seconds if not configured or 0
	}
	timeout := time.Duration(timeoutVal) * time.Second
	client := &http.Client{
		Timeout: timeout,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create request body
	var bodyReader io.Reader
	var multipartWriter *multipart.Writer
	if body != "" {
		if contentType == "Multipart" {
			// Parse multipart fields from body string (temporary format: name1=value1&name2=file:/path/to/file)
			var b bytes.Buffer
			multipartWriter = multipart.NewWriter(&b)
			err := createMultipartBodyWithWriter(body, multipartWriter)
			if err != nil {
				return nil, fmt.Errorf("failed to create multipart body: %v", err)
			}
			// Close the writer to write the final boundary
			multipartWriter.Close()
			bodyReader = &b
		} else {
			bodyReader = bytes.NewBufferString(body)
		}
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
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
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			// Use explicit content type if provided, otherwise auto-detect
			if contentType != "" {
				switch contentType {
				case "JSON":
					req.Header.Set("Content-Type", "application/json")
				case "Multipart":
					if multipartWriter != nil {
						req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
					} else {
						req.Header.Set("Content-Type", "multipart/form-data")
					}
				}
			} else {
				// Fallback to auto-detection for backward compatibility
				if strings.TrimSpace(body) != "" {
					if strings.HasPrefix(strings.TrimSpace(body), "{") && strings.HasSuffix(strings.TrimSpace(body), "}") {
						req.Header.Set("Content-Type", "application/json")
					}
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

	result := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    headersMap,
		Cookies:    resp.Cookies(),
		Body:       string(respBody),
		BodyBytes:  respBody,
		Duration:   duration,
		Timestamp:  time.Now(),
		BodySize:   len(respBody),
	}

	return result, nil
}

// createMultipartBodyWithWriter creates a multipart form body from a string definition
// Format: name1=value1&name2=file:/path/to/file&name3=value3
func createMultipartBodyWithWriter(bodyDef string, writer *multipart.Writer) error {
	// Validate input
	bodyDef = strings.TrimSpace(bodyDef)
	if bodyDef == "" {
		return fmt.Errorf("multipart body cannot be empty")
	}

	// Parse field definitions
	fields := strings.Split(bodyDef, "&")

	fieldNames := make(map[string]bool) // Track field names for uniqueness
	validFields := 0

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid field format: %s (expected name=value)", field)
		}

		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Validate field name
		if name == "" {
			return fmt.Errorf("field name cannot be empty")
		}
		if fieldNames[name] {
			return fmt.Errorf("duplicate field name: %s", name)
		}
		fieldNames[name] = true

		if strings.HasPrefix(value, "file:") {
			// File field
			filePath := strings.TrimPrefix(value, "file:")

			// Validate file path is absolute
			if !filepath.IsAbs(filePath) {
				return fmt.Errorf("file path must be absolute: %s", filePath)
			}

			// Validate file exists and is readable
			fileInfo, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", filePath)
			}
			if err != nil {
				return fmt.Errorf("cannot access file %s: %v", filePath, err)
			}

			// Check file size (warn for large files, but don't block)
			const maxRecommendedSize = 10 * 1024 * 1024 // 10MB
			if fileInfo.Size() > maxRecommendedSize {
				// Log warning but continue (avoiding fmt.Printf in TUI)
				// fmt.Printf("Warning: Large file detected (%d MB): %s\n", fileInfo.Size()/(1024*1024), filePath)
			}

			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", filePath, err)
			}
			defer file.Close()

			// Get filename from path
			parts := strings.Split(filePath, "/")
			filename := parts[len(parts)-1]

			fw, err := writer.CreateFormFile(name, filename)
			if err != nil {
				return fmt.Errorf("failed to create form file: %v", err)
			}

			_, err = io.Copy(fw, file)
			if err != nil {
				return fmt.Errorf("failed to copy file content: %v", err)
			}
		} else {
			// Text field
			err := writer.WriteField(name, value)
			if err != nil {
				return fmt.Errorf("failed to write field: %v", err)
			}
		}
		validFields++
	}

	if validFields == 0 {
		return fmt.Errorf("no fields found in multipart body")
	}

	return nil
}

// FormatResponse formats the HTTP response for display in the UI
func FormatResponse(response *HTTPResponse) string {
	var result strings.Builder

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
func GetCurrentRequestData(methodDropdown *tview.DropDown, urlInput *URLVariableInput, currentRequest *workspace.Request) (string, string, string, map[string]string, map[string]string) {
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

	// Get query params from UI
	queryParams := getQueryParamsFromUI()

	return method, url, body, headers, queryParams
}

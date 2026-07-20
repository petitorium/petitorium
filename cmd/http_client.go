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
	"github.com/petitorium/petitorium/version"
	"github.com/petitorium/petitorium/workspace"
)

// HTTPResponse represents the response from an HTTP request
type HTTPResponse struct {
	StatusCode    int
	Status        string
	Headers       map[string][]string
	Cookies       []*http.Cookie
	Body          string
	BodyBytes     []byte
	Duration      time.Duration
	TotalDuration time.Duration
	Timestamp     time.Time
	BodySize      int
}

func convertCookies(jarCookies []workspace.Cookie) []*http.Cookie {
	result := make([]*http.Cookie, 0, len(jarCookies))
	for _, c := range jarCookies {
		result = append(result, c.ToHttpCookie())
	}
	return result
}

func httpCookieToWorkspaceCookie(cookies []*http.Cookie) []workspace.Cookie {
	result := make([]workspace.Cookie, 0, len(cookies))
	for _, c := range cookies {
		expires := ""
		if !c.Expires.IsZero() {
			expires = c.Expires.Format(time.RFC3339)
		}
		sameSite := ""
		switch c.SameSite {
		case http.SameSiteNoneMode:
			sameSite = "none"
		case http.SameSiteLaxMode:
			sameSite = "lax"
		case http.SameSiteStrictMode:
			sameSite = "strict"
		}
		result = append(result, workspace.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  expires,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
			SameSite: sameSite,
		})
	}
	return result
}

// SendRequest sends an HTTP request with the given parameters
func SendRequest(method, urlStr, body string, contentType string, headers map[string]string, queryParams map[string]string, jar *workspace.CookieJar) (*HTTPResponse, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	if len(queryParams) > 0 {
		q := parsedURL.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		parsedURL.RawQuery = q.Encode()
		urlStr = parsedURL.String()
	}

	timeoutVal := config.C.RequestTimeout
	if timeoutVal <= 0 {
		timeoutVal = 60
	}
	timeout := time.Duration(timeoutVal) * time.Second

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var bodyReader io.Reader
	var multipartWriter *multipart.Writer
	if body != "" {
		if contentType == "Multipart" {
			var b bytes.Buffer
			multipartWriter = multipart.NewWriter(&b)
			err := createMultipartBodyWithWriter(body, multipartWriter)
			if err != nil {
				return nil, fmt.Errorf("failed to create multipart body: %v", err)
			}
			multipartWriter.Close()
			bodyReader = &b
		} else {
			bodyReader = bytes.NewBufferString(body)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if body != "" && req.Header.Get("Content-Type") == "" {
		switch method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
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
				if strings.TrimSpace(body) != "" {
					if strings.HasPrefix(strings.TrimSpace(body), "{") && strings.HasSuffix(strings.TrimSpace(body), "}") {
						req.Header.Set("Content-Type", "application/json")
					}
				}
			}
		}
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Petitorium/"+version.Version)
	}

	if jar != nil {
		cookies := jar.GetCookies(parsedURL)
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	startTime := time.Now()
	finalURL := parsedURL
	var lastResp *http.Response
	capturedCookies := []*http.Cookie{}

	for redirectCount := 0; redirectCount < 10; redirectCount++ {
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %v", err)
		}
		defer resp.Body.Close()

		lastResp = resp

		if jar != nil {
			cookies := resp.Cookies()
			capturedCookies = append(capturedCookies, cookies...)
			if len(cookies) > 0 {
				jar.SetCookies(finalURL, cookies)
			}
		}

		if resp.StatusCode < 300 || resp.StatusCode > 399 {
			break
		}

		locHeader := resp.Header.Get("Location")
		if locHeader == "" {
			break
		}

		locURL, err := finalURL.Parse(locHeader)
		if err != nil {
			break
		}
		finalURL = locURL

		bodyReader = nil
		req, err = http.NewRequestWithContext(ctx, method, finalURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create redirect request: %v", err)
		}

		req.Header.Set("User-Agent", "Petitorium/"+version.Version)

		if jar != nil {
			cookies := jar.GetCookies(finalURL)
			for _, c := range cookies {
				req.AddCookie(c)
			}
		}
	}

	duration := time.Since(startTime)

	respBody, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	headersMap := make(map[string][]string)
	for key, values := range lastResp.Header {
		headersMap[key] = values
	}

	result := &HTTPResponse{
		StatusCode:    lastResp.StatusCode,
		Status:        lastResp.Status,
		Headers:       headersMap,
		Cookies:       capturedCookies,
		Body:          string(respBody),
		BodyBytes:     respBody,
		Duration:      duration,
		TotalDuration: duration,
		Timestamp:     time.Now(),
		BodySize:      len(respBody),
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
		// Disabled fields are persisted with a leading '!' marker; skip them on send
		// (collectMultipartFieldsForSend already excludes them).
		if strings.HasPrefix(name, "!") {
			continue
		}
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

func GetCurrentRequestData(methodDropdown *tview.DropDown, urlInput *URLVariableInput, currentRequest *workspace.Request) (string, string, string, map[string]workspace.Entry, map[string]workspace.Entry) {
	_, method := methodDropdown.GetCurrentOption()

	url := urlInput.GetText()

	body := ""
	if currentRequest != nil {
		body = currentRequest.Body
	}

	headers := getHeadersFromUI()

	queryParams := getQueryParamsFromUI()

	return method, url, body, headers, queryParams
}

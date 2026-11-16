package cmd

import (
	"strings"
	"testing"
)

func TestGenerateCurlCommandAddsJSONContentType(t *testing.T) {
	body := `{"name":"example"}`
	got := generateCurlCommand("POST", "https://example.com/api", map[string]string{}, body, "JSON", map[string]string{})
	if !strings.Contains(got, "-H 'Content-Type: application/json'") {
		t.Fatalf("expected JSON Content-Type header, got: %s", got)
	}
}

func TestGenerateCurlCommandDoesNotDuplicateContentType(t *testing.T) {
	body := `{"name":"example"}`
	got := generateCurlCommand("POST", "https://example.com/api", map[string]string{"Content-Type": "application/json"}, body, "JSON", map[string]string{})
	if strings.Count(got, "Content-Type: application/json") != 1 {
		t.Fatalf("expected a single Content-Type header, got: %s", got)
	}
}

func TestGenerateCurlCommandSkipsContentTypeForMultipart(t *testing.T) {
	got := generateCurlCommand("POST", "https://example.com/api", map[string]string{}, "field=value", "Multipart", map[string]string{})
	if strings.Contains(got, "Content-Type: application/json") {
		t.Fatalf("multipart export must not add a JSON Content-Type header, got: %s", got)
	}
}

func TestGenerateCurlCommandAddsXMLContentType(t *testing.T) {
	body := `<envelope><body/></envelope>`
	got := generateCurlCommand("POST", "https://example.com/api", map[string]string{}, body, "XML", map[string]string{})
	if !strings.Contains(got, "-H 'Content-Type: text/xml; charset=utf-8'") {
		t.Fatalf("expected XML Content-Type header, got: %s", got)
	}
	if !strings.Contains(got, "--data-binary") {
		t.Fatalf("expected --data-binary for XML body, got: %s", got)
	}
}

func TestGenerateCurlCommandAddsNoContentTypeForRaw(t *testing.T) {
	body := `plain text body`
	got := generateCurlCommand("POST", "https://example.com/api", map[string]string{}, body, "Raw", map[string]string{})
	if strings.Contains(got, "Content-Type:") {
		t.Fatalf("raw export must not infer a Content-Type header, got: %s", got)
	}
}

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

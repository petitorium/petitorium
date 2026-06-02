package cmd

import (
	"testing"

	"github.com/petitorium/petitorium/workspace"
)

func TestProcessCommandRunnerTags_Simple(t *testing.T) {
	tag := `{{command-runner:run command="echo hello" type="string"}}`
	got := processCommandRunnerTags(tag)
	if got != "hello" {
		t.Errorf("processCommandRunnerTags(%q) = %q, want %q", tag, got, "hello")
	}
}

func TestProcessCommandRunnerTags_JSONPath(t *testing.T) {
	// Use octal escapes (\042) so the command itself contains no double quotes.
	tag := `{{command-runner:run command="printf '{\042name\042:\042alice\042}'" type="json" jsonPath="name"}}`
	got := processCommandRunnerTags(tag)
	if got != "alice" {
		t.Errorf("processCommandRunnerTags(%q) = %q, want %q", tag, got, "alice")
	}
}

func TestProcessCommandRunnerTags_Passthrough(t *testing.T) {
	text := "no tags here"
	got := processCommandRunnerTags(text)
	if got != text {
		t.Errorf("processCommandRunnerTags(%q) = %q, want %q", text, got, text)
	}
}

func TestProcessCommandRunnerTags_CommandError(t *testing.T) {
	tag := `{{command-runner:run command="exit 1" type="string"}}`
	got := processCommandRunnerTags(tag)
	if got != tag {
		t.Errorf("processCommandRunnerTags(%q) = %q, want original tag preserved", tag, got)
	}
}

func TestProcessCommandRunnerTags_Multiple(t *testing.T) {
	text := `before {{command-runner:run command="echo first" type="string"}} middle {{command-runner:run command="echo second" type="string"}} after`
	got := processCommandRunnerTags(text)
	want := "before first middle second after"
	if got != want {
		t.Errorf("processCommandRunnerTags(%q) = %q, want %q", text, got, want)
	}
}

func TestResolveEnvVarsFromIndex_Base(t *testing.T) {
	envs := []workspace.Environment{
		{Name: "Base", Variables: map[string]string{"a": "1"}},
		{Name: "Dev", Variables: map[string]string{"b": "2"}},
	}
	got := resolveEnvVarsFromIndex(0, envs, nil, "", nil)
	if got == nil {
		t.Fatal("resolveEnvVarsFromIndex(0) returned nil")
	}
	if got["a"] != "1" {
		t.Errorf("a = %q, want %q", got["a"], "1")
	}
}

func TestResolveEnvVarsFromIndex_Specific(t *testing.T) {
	envs := []workspace.Environment{
		{Name: "Base", Variables: map[string]string{"a": "1"}},
		{Name: "Dev", Variables: map[string]string{"b": "2"}},
	}
	// Dropdown index 0 = synthetic "Base Environment", index 1 = envs[0], index 2 = envs[1]
	got := resolveEnvVarsFromIndex(2, envs, nil, "", nil)
	if got == nil {
		t.Fatal("resolveEnvVarsFromIndex(2) returned nil")
	}
	if got["b"] != "2" {
		t.Errorf("b = %q, want %q", got["b"], "2")
	}
}

func TestResolveEnvVarsFromIndex_NotFound(t *testing.T) {
	envs := []workspace.Environment{
		{Name: "Dev", Variables: map[string]string{"b": "2"}},
	}
	got := resolveEnvVarsFromIndex(0, envs, nil, "", nil)
	if got != nil {
		t.Errorf("resolveEnvVarsFromIndex(0) = %v, want nil", got)
	}
}

func TestParseExistingCommandRunnerTag_JSONEscapedQuotes(t *testing.T) {
	// Tag as it appears inside a JSON string (quotes escaped)
	text := `{"pass": "{{command-runner:run command=\"echo secret\" type=\"string\"}}"}`
	cmd, outType, jsonPath := parseExistingCommandRunnerTag(text)
	if cmd != "echo secret" {
		t.Errorf("command = %q, want %q", cmd, "echo secret")
	}
	if outType != "string" {
		t.Errorf("type = %q, want %q", outType, "string")
	}
	if jsonPath != "" {
		t.Errorf("jsonPath = %q, want empty", jsonPath)
	}
}

func TestSanitizeCommandRunnerTagsInJSON(t *testing.T) {
	text := `{"pass": "{{command-runner:run command="echo secret" type="string"}}"}`
	got := sanitizeCommandRunnerTagsInJSON(text)
	want := `{"pass": "{{command-runner:run command=\"echo secret\" type=\"string\"}}"}`
	if got != want {
		t.Errorf("sanitizeCommandRunnerTagsInJSON(%q) = %q, want %q", text, got, want)
	}
}

func TestSanitizeCommandRunnerTagsInJSON_AlreadyEscaped(t *testing.T) {
	text := `{"pass": "{{command-runner:run command=\"echo secret\" type=\"string\"}}"}`
	got := sanitizeCommandRunnerTagsInJSON(text)
	if got != text {
		t.Errorf("sanitizeCommandRunnerTagsInJSON should not double-escape: got %q", got)
	}
}

func TestSanitizeCommandRunnerTagsInJSON_NoTag(t *testing.T) {
	text := `{"pass": "simple value"}`
	got := sanitizeCommandRunnerTagsInJSON(text)
	if got != text {
		t.Errorf("sanitizeCommandRunnerTagsInJSON should passthrough: got %q", got)
	}
}

func TestCursorByteOffset(t *testing.T) {
	text := "line1\nline2\nline3"
	if cursorByteOffset(text, 0, 0) != 0 {
		t.Errorf("cursorByteOffset(row=0, col=0) = %d, want 0", cursorByteOffset(text, 0, 0))
	}
	if cursorByteOffset(text, 1, 0) != 6 {
		t.Errorf("cursorByteOffset(row=1, col=0) = %d, want 6", cursorByteOffset(text, 1, 0))
	}
	if cursorByteOffset(text, 1, 3) != 9 {
		t.Errorf("cursorByteOffset(row=1, col=3) = %d, want 9", cursorByteOffset(text, 1, 3))
	}
	if cursorByteOffset(text, 2, 5) != 17 {
		t.Errorf("cursorByteOffset(row=2, col=5) = %d, want 17", cursorByteOffset(text, 2, 5))
	}
}

func TestGetResolvedEnvironmentVariables_CommandRunnerAndCrossRef(t *testing.T) {
	envs := []workspace.Environment{
		{
			Name: "Base",
			Variables: map[string]string{
				"host":      "localhost",
				"timestamp": `{{command-runner:run command="echo 12345" type="string"}}`,
				"url":       "{{host}}/api?t={{timestamp}}",
			},
		},
	}

	env := &envs[0]
	got := getResolvedEnvironmentVariables(env, envs, nil, "", nil)

	if got["timestamp"] != "12345" {
		t.Errorf("timestamp = %q, want %q", got["timestamp"], "12345")
	}
	if got["url"] != "localhost/api?t=12345" {
		t.Errorf("url = %q, want %q", got["url"], "localhost/api?t=12345")
	}
}

func TestSavePluginEnvironmentChanges_PreservesTags(t *testing.T) {
	envs := []workspace.Environment{
		{Name: "Base", Variables: map[string]string{
			"timestamp": `{{command-runner:run command="echo 12345" type="string"}}`,
			"static":    "old",
		}},
	}

	// Simulate resolved env being passed back
	resolvedEnv := map[string]string{
		"timestamp": "12345", // resolved — should NOT be saved
		"static":    "new",   // genuinely changed — should be saved
	}

	savePluginEnvironmentChanges(resolvedEnv, 0, &envs)

	if envs[0].Variables["timestamp"] != `{{command-runner:run command="echo 12345" type="string"}}` {
		t.Errorf("timestamp was overwritten: %q", envs[0].Variables["timestamp"])
	}
	if envs[0].Variables["static"] != "new" {
		t.Errorf("static was not updated: %q", envs[0].Variables["static"])
	}
}

func TestSavePluginEnvironmentChanges_SavesPluginAdditions(t *testing.T) {
	envs := []workspace.Environment{
		{Name: "Base", Variables: map[string]string{
			"existing": "val",
		}},
	}

	resolvedEnv := map[string]string{
		"existing":  "val",
		"pluginKey": "pluginValue",
	}

	savePluginEnvironmentChanges(resolvedEnv, 0, &envs)

	if envs[0].Variables["pluginKey"] != "pluginValue" {
		t.Errorf("pluginKey was not added: %q", envs[0].Variables["pluginKey"])
	}
}

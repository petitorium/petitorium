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

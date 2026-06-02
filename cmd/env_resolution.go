package cmd

import (
	"strings"

	"github.com/tidwall/gjson"

	"github.com/petitorium/petitorium/plugins"
	"github.com/petitorium/petitorium/workspace"
)

// getResolvedEnvironmentVariables builds the fully resolved environment variable
// map for the given environment. The pipeline is:
//  1. Merge base + own variables
//  2. Run PreEnvironmentResolution plugin hooks
//  3. Execute {{command-runner:run ...}} tags in values
//  4. Resolve {{key}} cross-references
//  5. Run PostEnvironmentResolution plugin hooks
func getResolvedEnvironmentVariables(
	env *workspace.Environment,
	allEnvs []workspace.Environment,
	pluginManager *plugins.PluginManager,
	workspaceName string,
	pluginConfig map[string]any,
) map[string]string {
	effective := env.GetEffectiveVariables(allEnvs)

	// Pre-resolution plugin hooks
	if pluginManager != nil {
		ctx := &plugins.HookContext{
			Environment: effective,
			Config:      pluginConfig,
			Workspace:   workspaceName,
		}
		pluginManager.ExecuteHooks(plugins.PreEnvironmentResolution, ctx)
		effective = ctx.Environment
	}

	// Resolve {{key}} cross-references so command-runner commands see
	// fully resolved values when they reference other env vars.
	workspace.ResolveVariableReferences(effective)

	// Execute command-runner tags in every value
	for k, v := range effective {
		effective[k] = processCommandRunnerTags(v, effective)
	}

	// Resolve any new {{key}} cross-references introduced by command outputs.
	workspace.ResolveVariableReferences(effective)

	// Post-resolution plugin hooks
	if pluginManager != nil {
		ctx := &plugins.HookContext{
			Environment: effective,
			Config:      pluginConfig,
			Workspace:   workspaceName,
		}
		pluginManager.ExecuteHooks(plugins.PostEnvironmentResolution, ctx)
		effective = ctx.Environment
	}

	return effective
}

// resolveEnvVarsFromIndex finds the environment by dropdown index and returns
// a fully resolved variable map. Returns nil when no environment matches.
func resolveEnvVarsFromIndex(
	currentEnvIndex int,
	environmentsData []workspace.Environment,
	pluginManager *plugins.PluginManager,
	workspaceName string,
	pluginConfig map[string]any,
) map[string]string {
	var env *workspace.Environment
	if currentEnvIndex == 0 {
		for i := range environmentsData {
			if environmentsData[i].Name == "Base" {
				env = &environmentsData[i]
				break
			}
		}
	} else if currentEnvIndex > 0 && currentEnvIndex <= len(environmentsData) {
		env = &environmentsData[currentEnvIndex-1]
	}
	if env == nil {
		return nil
	}
	return getResolvedEnvironmentVariables(env, environmentsData, pluginManager, workspaceName, pluginConfig)
}

// processCommandRunnerTags scans text for {{command-runner:run ...}} tags,
// executes each command, and replaces the tag with the trimmed output.
// Before execution, any {{key}} placeholders inside the command attribute are
// resolved using the provided variables map, so command-runner tags can
// reference other environment variables.
// On error the original tag is left untouched.  Tags are processed right-to-left
// so byte offsets remain valid.
func processCommandRunnerTags(text string, vars map[string]string) string {
	tags := scanTags(text)
	if len(tags) == 0 {
		return text
	}

	result := text
	for i := len(tags) - 1; i >= 0; i-- {
		dt := tags[i]
		if dt.Plugin != "command-runner" {
			continue
		}

		matches := paramRegex.FindAllStringSubmatch(dt.Inner, -1)
		params := make(map[string]string)
		for _, m := range matches {
			if len(m) == 3 {
				params[m[1]] = m[2]
			}
		}

		command := params["command"]
		outputType := params["type"]
		jsonPath := params["jsonPath"]
		if command == "" {
			continue
		}

		// Resolve {{key}} placeholders inside the command string using other env vars.
		command = substituteVariables(command, vars)

		output, err := RunShellCommand(command)
		if err != nil {
			continue // Leave tag as-is on error
		}

		if outputType == "json" && jsonPath != "" {
			result := gjson.Get(output, jsonPath)
			output = result.String()
		}

		result = result[:dt.Start] + output + result[dt.End:]
	}

	return result
}

// sanitizeCommandRunnerTagsInJSON scans text for {{command-runner:run ...}} tags
// and escapes any unescaped double quotes inside them so the surrounding JSON
// remains valid. It is used as a safety net in the env modal save handler.
func sanitizeCommandRunnerTagsInJSON(text string) string {
	prefix := "{{command-runner:run"
	result := text
	offset := 0
	for {
		idx := strings.Index(result[offset:], prefix)
		if idx == -1 {
			break
		}
		start := offset + idx
		end := findCommandRunnerTagEnd(result, start)
		if end == -1 {
			break
		}
		tag := result[start:end]
		sanitizedTag := escapeUnescapedQuotes(tag)
		result = result[:start] + sanitizedTag + result[end:]
		offset = start + len(sanitizedTag)
	}
	return result
}

// findCommandRunnerTagEnd scans forward from start (which points at the
// opening "{{command-runner:run") and returns the byte offset just after the
// matching "}}". It tracks whether we're inside a quoted string so that "}}"
// inside a command does not terminate the tag prematurely.
func findCommandRunnerTagEnd(text string, start int) int {
	i := start + len("{{command-runner:run")
	inQuotes := false
	for i < len(text)-1 {
		if text[i] == '\\' && i+1 < len(text) && text[i+1] == '"' {
			i += 2
			continue
		}
		if text[i] == '"' {
			inQuotes = !inQuotes
			i++
			continue
		}
		if !inQuotes && text[i] == '}' && text[i+1] == '}' {
			return i + 2
		}
		i++
	}
	return -1
}

// escapeUnescapedQuotes returns a copy of s where every unescaped double quote
// is replaced by an escaped one (\"). Quotes that are already escaped are left
// untouched.
func escapeUnescapedQuotes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			if i > 0 && s[i-1] == '\\' {
				b.WriteByte('"')
			} else {
				b.WriteString("\\\"")
			}
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

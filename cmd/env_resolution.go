package cmd

import (
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

	// Execute command-runner tags in every value
	for k, v := range effective {
		effective[k] = processCommandRunnerTags(v)
	}

	// Resolve {{key}} cross-references
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
// On error the original tag is left untouched.  Tags are processed right-to-left
// so byte offsets remain valid.
func processCommandRunnerTags(text string) string {
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

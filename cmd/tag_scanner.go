package cmd

import (
	"regexp"
)

// tagRegex matches the new namespace format: {{plugin:action key="value" ...}}
// Groups: 1 = plugin name, 2 = action name, 3 = raw inner params.
var tagRegex = regexp.MustCompile(`\{\{([\w-]+):(\w+)((?:\s+\w+="[^"]*")*)\}\}`)

// DetectedTag represents a single {{...}} tag found in text.
type DetectedTag struct {
	Raw    string // full "{{plugin:action ...}}"
	Start  int    // byte offset in parent text
	End    int    // byte offset (exclusive)
	Plugin string // e.g. "command-runner"
	Action string // e.g. "run"
	Inner  string // raw inner params: ` command="..." type="..."`
}

// scanTags finds all namespace-format tags in the given text and returns them
// in the order they appear.
func scanTags(text string) []DetectedTag {
	matches := tagRegex.FindAllStringSubmatchIndex(text, -1)
	if matches == nil {
		return nil
	}

	results := make([]DetectedTag, 0, len(matches))
	for _, m := range matches {
		// m indices: [fullStart, fullEnd, pluginStart, pluginEnd, actionStart, actionEnd, innerStart, innerEnd]
		if len(m) < 8 {
			continue
		}
		results = append(results, DetectedTag{
			Raw:    text[m[0]:m[1]],
			Start:  m[0],
			End:    m[1],
			Plugin: text[m[2]:m[3]],
			Action: text[m[4]:m[5]],
			Inner:  text[m[6]:m[7]],
		})
	}
	return results
}

// hasEditableTag reports whether text contains at least one tag that looks
// like it belongs to a known plugin.  This is a quick check used before
// opening the tag editor.
func hasEditableTag(text string) bool {
	return tagRegex.MatchString(text)
}

package cmd

import "testing"

func TestStripJSONComments(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no comments",
			in:   `{"a": 1}`,
			want: `{"a": 1}`,
		},
		{
			name: "line comment",
			in:   "{\n\t// a comment\n\t\"a\": 1\n}",
			want: "{\n\t\n\t\"a\": 1\n}",
		},
		{
			name: "block comment",
			in:   `{"a": /* inline */ 1}`,
			want: `{"a":  1}`,
		},
		{
			name: "multi-line block comment preserves newlines",
			in:   "{\"a\": 1,/* multi\nline */\"b\": 2}",
			want: "{\"a\": 1,\n\"b\": 2}",
		},
		{
			name: "line comment at end of file",
			in:   `{"a": 1}// trailing`,
			want: `{"a": 1}`,
		},
		{
			name: "comment markers inside strings are preserved",
			in:   `{"url": "http://example.com", "a": "x//y", "b": "/*not a comment*/"}`,
			want: `{"url": "http://example.com", "a": "x//y", "b": "/*not a comment*/"}`,
		},
		{
			name: "escaped quotes inside strings",
			in:   `{"s": "say \"hi\" // still string"}`,
			want: `{"s": "say \"hi\" // still string"}`,
		},
		{
			name: "mixed comments",
			in:   "{\n// line\n\"a\": 1,\n/* block\n   comment */\n\"b\": 2\n}",
			want: "{\n\n\"a\": 1,\n\n\n\"b\": 2\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripJSONComments(tt.in)
			if got != tt.want {
				t.Errorf("stripJSONComments() mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeRequestBodyForSend(t *testing.T) {
	if got := sanitizeRequestBodyForSend(`{"a": 1} // c`, "JSON"); got != `{"a": 1} ` {
		t.Errorf("JSON body not sanitized, got %q", got)
	}
	if got := sanitizeRequestBodyForSend("// not json", "Multipart"); got != "// not json" {
		t.Errorf("non-JSON body must be unchanged, got %q", got)
	}
	if got := sanitizeRequestBodyForSend("// not json", ""); got != "// not json" {
		t.Errorf("empty content type must be unchanged, got %q", got)
	}
}

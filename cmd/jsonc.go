package cmd

import "strings"

// stripJSONComments removes // line comments and /* */ block comments from a
// JSON (JSONC) document while leaving comment markers inside double-quoted
// string literals untouched.
//
// Newlines inside comments are preserved so that line numbering in any
// downstream error messages stays stable.
func stripJSONComments(body string) string {
	var out strings.Builder
	out.Grow(len(body))

	inString := false
	inLineComment := false
	inBlockComment := false

	for i := 0; i < len(body); i++ {
		ch := body[i]

		switch {
		case inLineComment:
			if ch == '\n' {
				inLineComment = false
				out.WriteByte('\n')
			}

		case inBlockComment:
			if ch == '\n' {
				out.WriteByte('\n')
			} else if ch == '*' && i+1 < len(body) && body[i+1] == '/' {
				inBlockComment = false
				i++
			}

		case inString:
			out.WriteByte(ch)
			if ch == '\\' && i+1 < len(body) {
				out.WriteByte(body[i+1])
				i++
			} else if ch == '"' {
				inString = false
			}

		default:
			switch ch {
			case '"':
				inString = true
				out.WriteByte(ch)
			case '/':
				if i+1 < len(body) {
					switch body[i+1] {
					case '/':
						inLineComment = true
						i++
					case '*':
						inBlockComment = true
						i++
					default:
						out.WriteByte(ch)
					}
				} else {
					out.WriteByte(ch)
				}
			default:
				out.WriteByte(ch)
			}
		}
	}

	return out.String()
}

// sanitizeRequestBodyForSend strips JSONC comments from JSON request bodies so
// comment-only characters are never transmitted to the server or resolved as
// environment variables / command-runner tags. Non-JSON bodies are returned
// unchanged.
func sanitizeRequestBodyForSend(body, contentType string) string {
	if contentType == "JSON" {
		return stripJSONComments(body)
	}
	return body
}

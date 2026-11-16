package cmd

// Content type identifiers persisted in workspace.Request.ContentType.
// These values are also the labels shown in the content-type dropdowns.
const (
	ContentTypeJSON      = "JSON"
	ContentTypeXML       = "XML"
	ContentTypeRaw       = "Raw"
	ContentTypeMultipart = "Multipart"
	ContentTypeNoBody    = "No Body"
)

// requestContentTypes is the ordered list of options shown in every
// content-type dropdown. Keep this as the single source of truth so the
// editor, request forms, and the send path stay in sync.
var requestContentTypes = []string{
	ContentTypeJSON,
	ContentTypeXML,
	ContentTypeRaw,
	ContentTypeMultipart,
	ContentTypeNoBody,
}

// isTextBodyContentType reports whether the content type uses the free-form
// body editor (JSON, XML, or Raw) as opposed to the multipart fields UI or
// the no-body placeholder.
func isTextBodyContentType(ct string) bool {
	switch ct {
	case ContentTypeJSON, ContentTypeXML, ContentTypeRaw:
		return true
	default:
		return false
	}
}

// bodyEditorExtension returns the file extension used for the external editor
// temp file based on the request content type, so the user's editor picks the
// right syntax highlighting (XML for SOAP payloads, txt for raw bodies, etc.).
func bodyEditorExtension(ct string) string {
	switch ct {
	case ContentTypeXML:
		return "xml"
	case ContentTypeRaw:
		return "txt"
	default:
		return "jsonc"
	}
}

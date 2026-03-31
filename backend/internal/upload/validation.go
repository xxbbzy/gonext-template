package upload

import (
	"fmt"
	"sort"
	"strings"
)

var compatibleMIMETypesByExtension = map[string][]string{
	".doc": {"application/msword"},
	".docx": {
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/zip",
	},
	".gif":  {"image/gif"},
	".jpeg": {"image/jpeg"},
	".jpg":  {"image/jpeg"},
	".pdf":  {"application/pdf"},
	".png":  {"image/png"},
	".txt":  {"text/plain"},
}

// CompatibleMIMETypes returns the compatible MIME types for an extension.
func CompatibleMIMETypes(extension string) []string {
	normalized := strings.ToLower(strings.TrimSpace(extension))
	mimeTypes, ok := compatibleMIMETypesByExtension[normalized]
	if !ok {
		return nil
	}

	out := make([]string, len(mimeTypes))
	copy(out, mimeTypes)
	return out
}

// ValidateSupportedExtensions ensures every configured extension has MIME rules.
func ValidateSupportedExtensions(extensions []string) error {
	unsupported := make([]string, 0)
	for _, ext := range extensions {
		if len(CompatibleMIMETypes(ext)) == 0 {
			unsupported = append(unsupported, ext)
		}
	}
	if len(unsupported) == 0 {
		return nil
	}

	sort.Strings(unsupported)
	return fmt.Errorf(
		"contains unsupported extension(s) for MIME compatibility: %s",
		strings.Join(unsupported, ", "),
	)
}

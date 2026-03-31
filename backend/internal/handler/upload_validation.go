package handler

import (
	"path"
	"strings"
	"unicode"

	"github.com/gabriel-vasile/mimetype"

	uploadvalidation "github.com/xxbbzy/gonext-template/backend/internal/upload"
)

func sanitizeUploadFilename(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	basename := path.Base(strings.ReplaceAll(trimmed, "\\", "/"))
	if basename == "." || basename == ".." {
		return ""
	}

	safeRunes := make([]rune, 0, len(basename))
	for _, r := range basename {
		if unicode.IsControl(r) {
			continue
		}
		if r == '/' || r == '\\' {
			continue
		}
		safeRunes = append(safeRunes, r)
	}

	return strings.TrimSpace(string(safeRunes))
}

func hasAllowedUploadExtension(extension string, allowedExtensions []string) bool {
	normalizedExtension := strings.TrimSpace(strings.ToLower(extension))
	for _, allowed := range allowedExtensions {
		normalizedAllowed := strings.TrimSpace(strings.ToLower(allowed))
		if normalizedAllowed == normalizedExtension {
			return true
		}
	}
	return false
}

func isCompatibleUploadMIME(extension string, detected *mimetype.MIME) bool {
	if detected == nil {
		return false
	}

	for _, allowedMIME := range uploadvalidation.CompatibleMIMETypes(extension) {
		if detected.Is(allowedMIME) {
			return true
		}
	}

	return false
}

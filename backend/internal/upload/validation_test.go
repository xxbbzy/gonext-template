package upload

import "testing"

func TestCompatibleMIMETypesIncludesZipForDocxOnly(t *testing.T) {
	docxMIMEs := CompatibleMIMETypes(".docx")
	if !containsMIME(docxMIMEs, "application/zip") {
		t.Fatalf("docx MIME set missing application/zip: %v", docxMIMEs)
	}

	docMIMEs := CompatibleMIMETypes(".doc")
	if containsMIME(docMIMEs, "application/zip") {
		t.Fatalf("doc MIME set should not include application/zip: %v", docMIMEs)
	}
}

func containsMIME(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

package handler

import "testing"

func TestHasAllowedUploadExtensionNormalizesValues(t *testing.T) {
	allowedExtensions := []string{" .png ", ".JPG"}

	if !hasAllowedUploadExtension(" .PnG ", allowedExtensions) {
		t.Fatalf("expected extension match after normalization")
	}
	if hasAllowedUploadExtension(".gif", allowedExtensions) {
		t.Fatalf("unexpected extension match")
	}
}

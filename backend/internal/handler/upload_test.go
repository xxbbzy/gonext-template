package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xxbbzy/gonext-template/backend/internal/config"
)

func TestUploadRejectsDisallowedFileType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	uploadHandler := NewUploadHandler(NewLocalStorage(tempDir, "http://localhost:8080"), &config.Config{
		Upload: config.UploadConfig{
			MaxSize:      1024,
			Dir:          tempDir,
			AllowedTypes: ".png,.jpg",
		},
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base("payload.exe"))
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write([]byte("malware")); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(resp)
	ctx.Request = req

	uploadHandler.Upload(ctx)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusBadRequest)
	}
}

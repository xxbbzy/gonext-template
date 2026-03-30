package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
)

const testUploadBaseURL = "http://localhost:8080"

type uploadResponseEnvelope struct {
	Code    int                `json:"code"`
	Data    uploadResponseData `json:"data"`
	Message string             `json:"message"`
}

type uploadResponseData struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

type failingFileStorageRepository struct {
	err error
}

func (r failingFileStorageRepository) SaveFile(context.Context, string, io.Reader) error {
	return r.err
}

func (r failingFileStorageRepository) DeleteFile(context.Context, string) error {
	return nil
}

func (r failingFileStorageRepository) GetFileURL(string) string {
	return testUploadBaseURL + "/uploads/failing.txt"
}

func TestUploadPersistsFullContentAndOriginalFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	uploadHandler := newTestUploadHandler(t, tempDir, ".png,.jpg,.txt", 1024*1024)
	content := bytes.Repeat([]byte("GoNext upload test payload."), 64)

	resp := performUploadRequest(t, uploadHandler, "report.txt", content)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	payload := decodeUploadResponse(t, resp)
	if payload.Code != 0 {
		t.Fatalf("code = %d, want 0", payload.Code)
	}
	if payload.Message != "success" {
		t.Fatalf("message = %q, want %q", payload.Message, "success")
	}
	if payload.Data.Filename != "report.txt" {
		t.Fatalf("filename = %q, want %q", payload.Data.Filename, "report.txt")
	}
	if payload.Data.Size != int64(len(content)) {
		t.Fatalf("size = %d, want %d", payload.Data.Size, len(content))
	}

	storedFilename := storedFilenameFromURL(t, payload.Data.URL)
	if storedFilename == "report.txt" {
		t.Fatalf("stored filename reused original name: %q", storedFilename)
	}
	if filepath.Ext(storedFilename) != ".txt" {
		t.Fatalf("stored extension = %q, want %q", filepath.Ext(storedFilename), ".txt")
	}
	if _, err := uuid.Parse(strings.TrimSuffix(storedFilename, ".txt")); err != nil {
		t.Fatalf("stored filename does not contain UUID basename: %v", err)
	}

	storedContent, err := os.ReadFile(filepath.Join(tempDir, storedFilename))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(storedContent, content) {
		t.Fatalf("stored content mismatch")
	}
}

func TestUploadSameNameTwiceCreatesDistinctFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	uploadHandler := newTestUploadHandler(t, tempDir, ".png,.jpg,.txt", 1024*1024)

	firstContent := []byte("alpha")
	secondContent := []byte("omega")

	firstResp := performUploadRequest(t, uploadHandler, "duplicate.txt", firstContent)
	secondResp := performUploadRequest(t, uploadHandler, "duplicate.txt", secondContent)

	if firstResp.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", firstResp.Code, http.StatusOK)
	}
	if secondResp.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d", secondResp.Code, http.StatusOK)
	}

	firstPayload := decodeUploadResponse(t, firstResp)
	secondPayload := decodeUploadResponse(t, secondResp)

	if firstPayload.Data.URL == secondPayload.Data.URL {
		t.Fatalf("URLs should differ, got %q", firstPayload.Data.URL)
	}
	if firstPayload.Data.Filename != "duplicate.txt" || secondPayload.Data.Filename != "duplicate.txt" {
		t.Fatalf("response filenames should preserve original name")
	}

	firstStoredFilename := storedFilenameFromURL(t, firstPayload.Data.URL)
	secondStoredFilename := storedFilenameFromURL(t, secondPayload.Data.URL)

	if firstStoredFilename == secondStoredFilename {
		t.Fatalf("stored filenames should differ, got %q", firstStoredFilename)
	}

	firstStoredContent, err := os.ReadFile(filepath.Join(tempDir, firstStoredFilename))
	if err != nil {
		t.Fatalf("ReadFile(first) error = %v", err)
	}
	secondStoredContent, err := os.ReadFile(filepath.Join(tempDir, secondStoredFilename))
	if err != nil {
		t.Fatalf("ReadFile(second) error = %v", err)
	}
	if !bytes.Equal(firstStoredContent, firstContent) {
		t.Fatalf("first stored content mismatch")
	}
	if !bytes.Equal(secondStoredContent, secondContent) {
		t.Fatalf("second stored content mismatch")
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("stored file count = %d, want 2", len(entries))
	}
}

func TestUploadRejectsDisallowedFileType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	uploadHandler := newTestUploadHandler(t, tempDir, ".png,.jpg", 1024)

	resp := performUploadRequest(t, uploadHandler, "payload.exe", []byte("malware"))

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusBadRequest)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("stored file count = %d, want 0", len(entries))
	}
}

func TestUploadReturnsInternalServerErrorWhenStorageFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	uploadService := service.NewUploadService(failingFileStorageRepository{err: errors.New("storage down")}, zap.NewNop())
	uploadHandler := newTestUploadHandlerWithService(uploadService, tempDir, ".txt", 1024)

	resp := performUploadRequest(t, uploadHandler, "report.txt", []byte("payload"))

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusInternalServerError)
	}
}

func TestUploadRejectsOversizedRequestBodyBeforeMultipartParsing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	uploadHandler := newTestUploadHandler(t, tempDir, ".txt", 1)
	largeContent := bytes.Repeat([]byte("A"), int(2*(1<<20)))

	resp := performUploadRequest(t, uploadHandler, "oversize.txt", largeContent)
	if resp.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusRequestEntityTooLarge)
	}
}

func newTestUploadHandler(t *testing.T, tempDir, allowedTypes string, maxSize int64) *UploadHandler {
	t.Helper()

	fileStorage, err := repository.NewLocalFileStorageRepository(tempDir, testUploadBaseURL)
	if err != nil {
		t.Fatalf("NewLocalFileStorageRepository() error = %v", err)
	}
	uploadService := service.NewUploadService(fileStorage, zap.NewNop())
	return newTestUploadHandlerWithService(uploadService, tempDir, allowedTypes, maxSize)
}

func newTestUploadHandlerWithService(uploadService *service.UploadService, tempDir, allowedTypes string, maxSize int64) *UploadHandler {
	return NewUploadHandler(uploadService, &config.Config{
		Upload: config.UploadConfig{
			MaxSize:      maxSize,
			Dir:          tempDir,
			AllowedTypes: allowedTypes,
		},
	})
}

func performUploadRequest(t *testing.T, uploadHandler *UploadHandler, filename string, content []byte) *httptest.ResponseRecorder {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(content); err != nil {
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

	return resp
}

func decodeUploadResponse(t *testing.T, resp *httptest.ResponseRecorder) uploadResponseEnvelope {
	t.Helper()

	var payload uploadResponseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return payload
}

func storedFilenameFromURL(t *testing.T, rawURL string) string {
	t.Helper()

	prefix := testUploadBaseURL + "/uploads/"
	if !strings.HasPrefix(rawURL, prefix) {
		t.Fatalf("url = %q, want prefix %q", rawURL, prefix)
	}

	return strings.TrimPrefix(rawURL, prefix)
}

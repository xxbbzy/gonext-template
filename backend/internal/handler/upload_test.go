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
	"github.com/xxbbzy/gonext-template/backend/internal/middleware"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

const testUploadBaseURL = "http://localhost:8080"
const testUploadRequestID = "req-upload-test"

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

	storedFilename := storedFilenameFromURL(t, payload.Data.URL, testUploadBaseURL)
	if storedFilename == "report.txt" {
		t.Fatalf("stored filename reused original name: %q", storedFilename)
	}
	if ext := filepath.Ext(storedFilename); ext != ".txt" {
		t.Fatalf("stored extension = %q, want %q", ext, ".txt")
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

	firstStoredFilename := storedFilenameFromURL(t, firstPayload.Data.URL, testUploadBaseURL)
	secondStoredFilename := storedFilenameFromURL(t, secondPayload.Data.URL, testUploadBaseURL)

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
	errorPayload := decodeUploadErrorResponse(t, resp)
	if errorPayload.Code != http.StatusBadRequest {
		t.Fatalf("error code = %d, want %d", errorPayload.Code, http.StatusBadRequest)
	}
	if errorPayload.Message != "file type not allowed" {
		t.Fatalf("error message = %q, want %q", errorPayload.Message, "file type not allowed")
	}
	if errorPayload.RequestID != testUploadRequestID {
		t.Fatalf("request_id = %q, want %q", errorPayload.RequestID, testUploadRequestID)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("stored file count = %d, want 0", len(entries))
	}
}

func TestUploadSanitizesPathLikeFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	uploadHandler := newTestUploadHandler(t, tempDir, ".png", 1024*1024)

	resp := performUploadRequest(t, uploadHandler, `C:\fakepath\avatar.png`, samplePNGContent())
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	payload := decodeUploadResponse(t, resp)
	if payload.Data.Filename != "avatar.png" {
		t.Fatalf("filename = %q, want %q", payload.Data.Filename, "avatar.png")
	}

	storedFilename := storedFilenameFromURL(t, payload.Data.URL, testUploadBaseURL)
	if ext := filepath.Ext(storedFilename); ext != ".png" {
		t.Fatalf("stored extension = %q, want %q", ext, ".png")
	}
}

func TestUploadRejectsMIMEMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	uploadHandler := newTestUploadHandler(t, tempDir, ".png", 1024)

	resp := performUploadRequest(t, uploadHandler, "avatar.png", []byte("plain text payload"))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusBadRequest)
	}

	errorPayload := decodeUploadErrorResponse(t, resp)
	if errorPayload.Message != "file type not allowed" {
		t.Fatalf("error message = %q, want %q", errorPayload.Message, "file type not allowed")
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("stored file count = %d, want 0", len(entries))
	}
}

func TestUploadUsesConfiguredPublicBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	const customUploadBaseURL = "https://assets.example.com"
	uploadHandler := newTestUploadHandlerWithBaseURL(t, tempDir, customUploadBaseURL, ".txt", 1024*1024)

	resp := performUploadRequest(t, uploadHandler, "report.txt", []byte("upload payload"))
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	payload := decodeUploadResponse(t, resp)
	storedFilename := storedFilenameFromURL(t, payload.Data.URL, customUploadBaseURL)
	if ext := filepath.Ext(storedFilename); ext != ".txt" {
		t.Fatalf("stored extension = %q, want %q", ext, ".txt")
	}
}

func TestUploadReturnsInternalServerErrorWhenStorageFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	uploadService := service.NewUploadService(failingFileStorageRepository{err: errors.New("storage down")}, zap.NewNop())
	uploadHandler := newTestUploadHandlerWithService(uploadService, tempDir, testUploadBaseURL, ".txt", 1024)

	resp := performUploadRequest(t, uploadHandler, "report.txt", []byte("payload"))

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusInternalServerError)
	}
	errorPayload := decodeUploadErrorResponse(t, resp)
	if errorPayload.Code != http.StatusInternalServerError {
		t.Fatalf("error code = %d, want %d", errorPayload.Code, http.StatusInternalServerError)
	}
	if errorPayload.Message != "internal server error" {
		t.Fatalf("error message = %q, want %q", errorPayload.Message, "internal server error")
	}
	if errorPayload.RequestID != testUploadRequestID {
		t.Fatalf("request_id = %q, want %q", errorPayload.RequestID, testUploadRequestID)
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
	errorPayload := decodeUploadErrorResponse(t, resp)
	if errorPayload.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("error code = %d, want %d", errorPayload.Code, http.StatusRequestEntityTooLarge)
	}
	if errorPayload.Message != "file too large" {
		t.Fatalf("error message = %q, want %q", errorPayload.Message, "file too large")
	}
	if errorPayload.RequestID != testUploadRequestID {
		t.Fatalf("request_id = %q, want %q", errorPayload.RequestID, testUploadRequestID)
	}
}

func newTestUploadHandler(t *testing.T, tempDir, allowedTypes string, maxSize int64) *UploadHandler {
	return newTestUploadHandlerWithBaseURL(t, tempDir, testUploadBaseURL, allowedTypes, maxSize)
}

func newTestUploadHandlerWithBaseURL(t *testing.T, tempDir, baseURL, allowedTypes string, maxSize int64) *UploadHandler {
	t.Helper()

	fileStorage, err := repository.NewLocalFileStorageRepository(tempDir, baseURL)
	if err != nil {
		t.Fatalf("NewLocalFileStorageRepository() error = %v", err)
	}
	uploadService := service.NewUploadService(fileStorage, zap.NewNop())
	return newTestUploadHandlerWithService(uploadService, tempDir, baseURL, allowedTypes, maxSize)
}

func newTestUploadHandlerWithService(uploadService *service.UploadService, tempDir, baseURL, allowedTypes string, maxSize int64) *UploadHandler {
	return NewUploadHandler(uploadService, &config.Config{
		Upload: config.UploadConfig{
			MaxSize:       maxSize,
			Dir:           tempDir,
			AllowedTypes:  allowedTypes,
			PublicBaseURL: baseURL,
		},
	})
}

func performUploadRequest(t *testing.T, uploadHandler *UploadHandler, filename string, content []byte) *httptest.ResponseRecorder {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
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
	ctx.Set(middleware.RequestIDKey, testUploadRequestID)
	ctx.Writer.Header().Set(middleware.RequestIDHeader, testUploadRequestID)

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

func decodeUploadErrorResponse(t *testing.T, resp *httptest.ResponseRecorder) response.ErrorResponse {
	t.Helper()

	var payload response.ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return payload
}

func storedFilenameFromURL(t *testing.T, rawURL, uploadBaseURL string) string {
	t.Helper()

	prefix := strings.TrimRight(uploadBaseURL, "/") + "/uploads/"
	if !strings.HasPrefix(rawURL, prefix) {
		t.Fatalf("url = %q, want prefix %q", rawURL, prefix)
	}

	return strings.TrimPrefix(rawURL, prefix)
}

func samplePNGContent() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47,
		0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00,
		0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00,
		0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63,
		0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d,
		0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49,
		0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
}

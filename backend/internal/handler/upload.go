package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

// Storage defines the file storage interface.
type Storage interface {
	Upload(filename string, data []byte) (string, error)
	Delete(filename string) error
	GetURL(filename string) string
}

// LocalStorage implements Storage using the local filesystem.
type LocalStorage struct {
	uploadDir string
	baseURL   string
}

// NewLocalStorage creates a new LocalStorage.
func NewLocalStorage(uploadDir, baseURL string) *LocalStorage {
	// Ensure upload directory exists
	os.MkdirAll(uploadDir, 0755)
	return &LocalStorage{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// Upload saves a file to the local filesystem.
func (s *LocalStorage) Upload(filename string, data []byte) (string, error) {
	path := filepath.Join(s.uploadDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return s.GetURL(filename), nil
}

// Delete removes a file from the local filesystem.
func (s *LocalStorage) Delete(filename string) error {
	path := filepath.Join(s.uploadDir, filename)
	return os.Remove(path)
}

// GetURL returns the URL for a file.
func (s *LocalStorage) GetURL(filename string) string {
	return s.baseURL + "/uploads/" + filename
}

// UploadHandler handles file upload requests.
type UploadHandler struct {
	storage Storage
	cfg     *config.Config
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(storage Storage, cfg *config.Config) *UploadHandler {
	return &UploadHandler{
		storage: storage,
		cfg:     cfg,
	}
}

// Upload handles file upload.
// @Summary Upload a file
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File to upload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 413 {object} response.Response
// @Router /api/v1/upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "no file provided")
		return
	}

	// Check file size
	if file.Size > h.cfg.Upload.MaxSize {
		response.Error(c, http.StatusRequestEntityTooLarge, 413, "file too large")
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := false
	for _, t := range h.cfg.GetAllowedFileTypes() {
		if strings.TrimSpace(t) == ext {
			allowed = true
			break
		}
	}
	if !allowed {
		response.BadRequest(c, "file type not allowed")
		return
	}

	// Read file
	f, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "failed to read file")
		return
	}
	defer f.Close()

	data := make([]byte, file.Size)
	if _, err := f.Read(data); err != nil {
		response.InternalServerError(c, "failed to read file")
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", file.Size, file.Filename)

	url, err := h.storage.Upload(filename, data)
	if err != nil {
		response.InternalServerError(c, "failed to upload file")
		return
	}

	response.Success(c, gin.H{
		"url":      url,
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// RegisterRoutes registers upload routes.
func (h *UploadHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, protectedMiddlewares ...gin.HandlerFunc) {
	handlers := append([]gin.HandlerFunc{authMiddleware}, protectedMiddlewares...)
	handlers = append(handlers, h.Upload)
	r.POST("/upload", handlers...)
}

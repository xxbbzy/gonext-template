package handler

import (
	"errors"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"

	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

const multipartOverheadAllowance = int64(1 << 20)

// UploadHandler handles file upload requests.
type UploadHandler struct {
	uploadService *service.UploadService
	cfg           *config.Config
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(uploadService *service.UploadService, cfg *config.Config) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
		cfg:           cfg,
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadRequestBytes())

	file, err := c.FormFile("file")
	if err != nil {
		if isMultipartBodyTooLarge(err) {
			response.Error(c, http.StatusRequestEntityTooLarge, 413, "file too large")
			return
		}
		response.BadRequest(c, "no file provided")
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}

	// Check file size
	if file.Size > h.cfg.Upload.MaxSize {
		response.Error(c, http.StatusRequestEntityTooLarge, 413, "file too large")
		return
	}

	// Check file type
	sanitizedFilename := sanitizeUploadFilename(file.Filename)
	ext := strings.ToLower(filepath.Ext(sanitizedFilename))
	if !hasAllowedUploadExtension(ext, h.cfg.GetAllowedFileTypes()) {
		response.BadRequest(c, "file type not allowed")
		return
	}

	// Read file
	f, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "failed to read file")
		return
	}
	defer func() { _ = f.Close() }()

	detectedMIME, err := mimetype.DetectReader(f)
	if err != nil {
		response.InternalServerError(c, "failed to read file")
		return
	}
	if !isCompatibleUploadMIME(ext, detectedMIME) {
		response.BadRequest(c, "file type not allowed")
		return
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		response.InternalServerError(c, "failed to read file")
		return
	}

	url, err := h.uploadService.UploadFile(c.Request.Context(), sanitizedFilename, f)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "failed to upload file")
		return
	}

	response.Success(c, gin.H{
		"url":      url,
		"filename": sanitizedFilename,
		"size":     file.Size,
	})
}

func (h *UploadHandler) maxUploadRequestBytes() int64 {
	if h.cfg.Upload.MaxSize > math.MaxInt64-multipartOverheadAllowance {
		return math.MaxInt64
	}
	return h.cfg.Upload.MaxSize + multipartOverheadAllowance
}

func isMultipartBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

// RegisterRoutes registers upload routes.
func (h *UploadHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, protectedMiddlewares ...gin.HandlerFunc) {
	handlers := append([]gin.HandlerFunc{authMiddleware}, protectedMiddlewares...)
	handlers = append(handlers, h.Upload)
	r.POST("/upload", handlers...)
}

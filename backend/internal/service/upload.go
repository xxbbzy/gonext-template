package service

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
)

// UploadService handles file upload business rules.
type UploadService struct {
	fileStorage repository.FileStorageRepository
	logger      *zap.Logger
}

// NewUploadService creates a new UploadService.
func NewUploadService(fileStorage repository.FileStorageRepository, logger *zap.Logger) *UploadService {
	return &UploadService{
		fileStorage: fileStorage,
		logger:      logger,
	}
}

// UploadFile stores file content and returns its accessible URL.
func (s *UploadService) UploadFile(ctx context.Context, originalName string, src io.Reader) (string, error) {
	storedName := generateStoredFilename(originalName)
	if err := s.fileStorage.SaveFile(ctx, storedName, src); err != nil {
		if s.logger != nil {
			s.logger.Error(
				"failed to store uploaded file",
				zap.String("stored_name", storedName),
				zap.String("original_name", originalName),
				zap.Error(err),
			)
		}
		return "", errcode.ErrInternalServer
	}
	return s.fileStorage.GetFileURL(storedName), nil
}

// RemoveFile removes a stored file by its stored name.
func (s *UploadService) RemoveFile(ctx context.Context, storedName string) error {
	if err := s.fileStorage.DeleteFile(ctx, storedName); err != nil {
		return errcode.ErrInternalServer
	}
	return nil
}

// GetFileURL returns the public URL for a stored file.
func (s *UploadService) GetFileURL(storedName string) string {
	return s.fileStorage.GetFileURL(storedName)
}

func generateStoredFilename(originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	return uuid.NewString() + ext
}

package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileStorageRepository defines persistence operations for uploaded files.
type FileStorageRepository interface {
	SaveFile(ctx context.Context, storedName string, src io.Reader) error
	DeleteFile(ctx context.Context, storedName string) error
	GetFileURL(storedName string) string
}

// LocalFileStorageRepository persists files on the local filesystem.
type LocalFileStorageRepository struct {
	uploadDir string
	baseURL   string
}

// NewLocalFileStorageRepository creates a local file storage repository.
func NewLocalFileStorageRepository(uploadDir, baseURL string) (*LocalFileStorageRepository, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	return &LocalFileStorageRepository{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}, nil
}

// SaveFile persists file content using a precomputed stored name.
func (r *LocalFileStorageRepository) SaveFile(_ context.Context, storedName string, src io.Reader) error {
	path := filepath.Join(r.uploadDir, storedName)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	if _, err := io.Copy(file, src); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// DeleteFile removes a file by stored name.
func (r *LocalFileStorageRepository) DeleteFile(_ context.Context, storedName string) error {
	safeStoredName, err := sanitizeStoredName(storedName)
	if err != nil {
		return err
	}

	uploadDirAbs, err := filepath.Abs(r.uploadDir)
	if err != nil {
		return fmt.Errorf("failed to resolve upload directory: %w", err)
	}

	path := filepath.Join(uploadDirAbs, safeStoredName)
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve target file path: %w", err)
	}

	separator := string(os.PathSeparator)
	if pathAbs != uploadDirAbs && !strings.HasPrefix(pathAbs, uploadDirAbs+separator) {
		return fmt.Errorf("invalid stored filename")
	}

	return os.Remove(pathAbs)
}

// GetFileURL returns the public URL of a stored file.
func (r *LocalFileStorageRepository) GetFileURL(storedName string) string {
	return r.baseURL + "/uploads/" + storedName
}

func sanitizeStoredName(storedName string) (string, error) {
	cleaned := filepath.Clean(storedName)
	if cleaned == "." || cleaned == ".." || cleaned == "" {
		return "", fmt.Errorf("invalid stored filename")
	}
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("invalid stored filename")
	}
	if cleaned != filepath.Base(cleaned) {
		return "", fmt.Errorf("invalid stored filename")
	}
	if strings.Contains(cleaned, "/") || strings.Contains(cleaned, "\\") {
		return "", fmt.Errorf("invalid stored filename")
	}
	return cleaned, nil
}

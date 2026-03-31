package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
)

type stubFileStorageRepository struct {
	saveErr error
	url     string
	urlErr  error
}

func (s stubFileStorageRepository) SaveFile(context.Context, string, io.Reader) error {
	return s.saveErr
}

func (s stubFileStorageRepository) DeleteFile(context.Context, string) error {
	return nil
}

func (s stubFileStorageRepository) GetFileURL(string) (string, error) {
	return s.url, s.urlErr
}

func TestUploadServiceUploadFileReturnsInternalServerWhenURLGenerationFails(t *testing.T) {
	svc := NewUploadService(stubFileStorageRepository{
		urlErr: errors.New("bad stored name"),
	}, zap.NewNop())

	got, err := svc.UploadFile(context.Background(), "avatar.png", strings.NewReader("payload"))
	if got != "" {
		t.Fatalf("UploadFile() url = %q, want empty string", got)
	}
	assertAppError(t, err, errcode.ErrInternal, http.StatusInternalServerError)
}

func TestUploadServiceGetFileURLReturnsInternalServerWhenStorageReturnsEmptyURL(t *testing.T) {
	svc := NewUploadService(stubFileStorageRepository{}, zap.NewNop())

	got, err := svc.GetFileURL("avatar.png")
	if got != "" {
		t.Fatalf("GetFileURL() url = %q, want empty string", got)
	}
	assertAppError(t, err, errcode.ErrInternal, http.StatusInternalServerError)
}

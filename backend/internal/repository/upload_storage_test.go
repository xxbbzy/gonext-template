package repository

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLocalFileStorageRepositoryRejectsExistingFilename(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewLocalFileStorageRepository(tempDir, "http://localhost:8080")
	if err != nil {
		t.Fatalf("NewLocalFileStorageRepository() error = %v", err)
	}

	if err := storage.SaveFile(context.Background(), "existing.txt", strings.NewReader("first")); err != nil {
		t.Fatalf("first SaveFile() error = %v", err)
	}
	if err := storage.SaveFile(context.Background(), "existing.txt", strings.NewReader("second")); err == nil {
		t.Fatal("second SaveFile() error = nil, want error")
	}

	content, err := os.ReadFile(filepath.Join(tempDir, "existing.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "first" {
		t.Fatalf("content = %q, want %q", string(content), "first")
	}
}

func TestLocalFileStorageRepositoryRemovesPartialFileOnReaderError(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewLocalFileStorageRepository(tempDir, "http://localhost:8080")
	if err != nil {
		t.Fatalf("NewLocalFileStorageRepository() error = %v", err)
	}

	storedName := "partial.txt"
	err = storage.SaveFile(context.Background(), storedName, &failingReader{
		firstChunk: []byte("partial data"),
		err:        errors.New("injected read failure"),
	})
	if err == nil {
		t.Fatal("SaveFile() error = nil, want error")
	}

	_, statErr := os.Stat(filepath.Join(tempDir, storedName))
	if !os.IsNotExist(statErr) {
		t.Fatalf("os.Stat() error = %v, want not-exist error", statErr)
	}
}

func TestNewLocalFileStorageRepositoryReturnsErrorWhenDirCreationFails(t *testing.T) {
	tempDir := t.TempDir()
	blockingPath := filepath.Join(tempDir, "not-a-directory")
	if err := os.WriteFile(blockingPath, []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := NewLocalFileStorageRepository(filepath.Join(blockingPath, "uploads"), "http://localhost:8080")
	if err == nil {
		t.Fatal("NewLocalFileStorageRepository() error = nil, want error")
	}
}

func TestLocalFileStorageRepositoryDeleteFileRejectsPathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewLocalFileStorageRepository(tempDir, "http://localhost:8080")
	if err != nil {
		t.Fatalf("NewLocalFileStorageRepository() error = %v", err)
	}

	if err := storage.DeleteFile(context.Background(), "../escape.txt"); err == nil {
		t.Fatal("DeleteFile() error = nil, want error")
	}
}

func TestLocalFileStorageRepositorySaveFileRejectsPathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewLocalFileStorageRepository(tempDir, "http://localhost:8080")
	if err != nil {
		t.Fatalf("NewLocalFileStorageRepository() error = %v", err)
	}

	if err := storage.SaveFile(context.Background(), "../escape.txt", strings.NewReader("payload")); err == nil {
		t.Fatal("SaveFile() error = nil, want error")
	}
}

func TestLocalFileStorageRepositoryGetFileURLEscapesPathSegment(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewLocalFileStorageRepository(tempDir, "http://localhost:8080")
	if err != nil {
		t.Fatalf("NewLocalFileStorageRepository() error = %v", err)
	}

	rawName := "name with space#?.txt"
	got, err := storage.GetFileURL(rawName)
	if err != nil {
		t.Fatalf("GetFileURL() error = %v", err)
	}
	want := "http://localhost:8080/uploads/" + url.PathEscape(rawName)
	if got != want {
		t.Fatalf("GetFileURL() = %q, want %q", got, want)
	}
}

func TestS3FileStorageRepositorySaveFileUsesPrefix(t *testing.T) {
	fakeClient := &fakeS3Client{}
	repo := newS3FileStorageRepositoryWithClient(fakeClient, S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Prefix:          "uploads/images",
		UseSSL:          true,
	})

	if err := repo.SaveFile(context.Background(), "avatar.png", strings.NewReader("payload")); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	if fakeClient.putBucket != "media-bucket" {
		t.Fatalf("put bucket = %q, want %q", fakeClient.putBucket, "media-bucket")
	}
	if fakeClient.putKey != "uploads/images/avatar.png" {
		t.Fatalf("put key = %q, want %q", fakeClient.putKey, "uploads/images/avatar.png")
	}
	if fakeClient.putBody != "payload" {
		t.Fatalf("put body = %q, want %q", fakeClient.putBody, "payload")
	}
}

func TestS3FileStorageRepositorySaveFileRejectsPathTraversal(t *testing.T) {
	fakeClient := &fakeS3Client{}
	repo := newS3FileStorageRepositoryWithClient(fakeClient, S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		UseSSL:          true,
	})

	if err := repo.SaveFile(context.Background(), "../escape.png", strings.NewReader("payload")); err == nil {
		t.Fatal("SaveFile() error = nil, want error")
	}
}

func TestS3FileStorageRepositoryDeleteFileUsesPrefix(t *testing.T) {
	fakeClient := &fakeS3Client{}
	repo := newS3FileStorageRepositoryWithClient(fakeClient, S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Prefix:          "uploads",
		UseSSL:          true,
	})

	if err := repo.DeleteFile(context.Background(), "avatar.png"); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}

	if fakeClient.deleteBucket != "media-bucket" {
		t.Fatalf("delete bucket = %q, want %q", fakeClient.deleteBucket, "media-bucket")
	}
	if fakeClient.deleteKey != "uploads/avatar.png" {
		t.Fatalf("delete key = %q, want %q", fakeClient.deleteKey, "uploads/avatar.png")
	}
}

func TestS3FileStorageRepositoryGetFileURLUsesPublicBaseURLOverride(t *testing.T) {
	repo := newS3FileStorageRepositoryWithClient(&fakeS3Client{}, S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Prefix:          "assets",
		PublicBaseURL:   "https://cdn.example.com",
		UseSSL:          true,
	})

	got, err := repo.GetFileURL("avatar with space.png")
	if err != nil {
		t.Fatalf("GetFileURL() error = %v", err)
	}
	want := "https://cdn.example.com/assets/avatar%20with%20space.png"
	if got != want {
		t.Fatalf("GetFileURL() = %q, want %q", got, want)
	}
}

func TestS3FileStorageRepositoryGetFileURLUsesPathStyleEndpoint(t *testing.T) {
	repo := newS3FileStorageRepositoryWithClient(&fakeS3Client{}, S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Endpoint:        "http://minio.local:9000",
		Prefix:          "assets",
		ForcePathStyle:  true,
		UseSSL:          false,
	})

	got, err := repo.GetFileURL("avatar.png")
	if err != nil {
		t.Fatalf("GetFileURL() error = %v", err)
	}
	want := "http://minio.local:9000/media-bucket/assets/avatar.png"
	if got != want {
		t.Fatalf("GetFileURL() = %q, want %q", got, want)
	}
}

func TestS3FileStorageRepositoryGetFileURLUsesVirtualHostedEndpoint(t *testing.T) {
	repo := newS3FileStorageRepositoryWithClient(&fakeS3Client{}, S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Endpoint:        "https://s3.example.com",
		Prefix:          "assets",
		ForcePathStyle:  false,
		UseSSL:          true,
	})

	got, err := repo.GetFileURL("avatar.png")
	if err != nil {
		t.Fatalf("GetFileURL() error = %v", err)
	}
	want := "https://media-bucket.s3.example.com/assets/avatar.png"
	if got != want {
		t.Fatalf("GetFileURL() = %q, want %q", got, want)
	}
}

func TestS3FileStorageRepositoryGetFileURLLogsInvalidStoredName(t *testing.T) {
	core, observed := observer.New(zap.ErrorLevel)
	repo := newS3FileStorageRepositoryWithClient(&fakeS3Client{}, S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		UseSSL:          true,
		Logger:          zap.New(core),
	})

	got, err := repo.GetFileURL("../avatar.png")
	if err == nil {
		t.Fatal("GetFileURL() error = nil, want error")
	}
	if got != "" {
		t.Fatalf("GetFileURL() = %q, want empty string", got)
	}
	if observed.Len() != 1 {
		t.Fatalf("observed logs = %d, want 1", observed.Len())
	}
	entry := observed.All()[0]
	if entry.Message != "failed to sanitize stored name for S3 file URL" {
		t.Fatalf("log message = %q, want %q", entry.Message, "failed to sanitize stored name for S3 file URL")
	}
}

func TestS3FileStorageRepositoryGetFileURLPreservesEndpointPathAndEscapesOnce(t *testing.T) {
	repo := newS3FileStorageRepositoryWithClient(&fakeS3Client{}, S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Endpoint:        "https://objects.example.com/storage",
		Prefix:          "assets",
		ForcePathStyle:  false,
		UseSSL:          true,
	})

	got, err := repo.GetFileURL("avatar with space.png")
	if err != nil {
		t.Fatalf("GetFileURL() error = %v", err)
	}
	want := "https://media-bucket.objects.example.com/storage/assets/avatar%20with%20space.png"
	if got != want {
		t.Fatalf("GetFileURL() = %q, want %q", got, want)
	}
}

func TestS3FileStorageRepositoryGetFileURLFallsBackToAWSStyle(t *testing.T) {
	repo := newS3FileStorageRepositoryWithClient(&fakeS3Client{}, S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "ap-southeast-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Prefix:          "assets",
		ForcePathStyle:  false,
		UseSSL:          true,
	})

	got, err := repo.GetFileURL("avatar.png")
	if err != nil {
		t.Fatalf("GetFileURL() error = %v", err)
	}
	want := "https://media-bucket.s3.ap-southeast-1.amazonaws.com/assets/avatar.png"
	if got != want {
		t.Fatalf("GetFileURL() = %q, want %q", got, want)
	}
}

func TestNormalizeS3FileStorageConfigRejectsMissingFields(t *testing.T) {
	_, err := normalizeS3FileStorageConfig(S3FileStorageConfig{})
	if err == nil {
		t.Fatal("normalizeS3FileStorageConfig() error = nil, want error")
	}
}

func TestNewS3FileStorageRepositoryIgnoresAmbientAWSProfile(t *testing.T) {
	t.Setenv("AWS_PROFILE", "does-not-exist")

	repo, err := NewS3FileStorageRepository(context.Background(), S3FileStorageConfig{
		Bucket:          "media-bucket",
		Region:          "us-east-1",
		Endpoint:        "http://minio.local:9000",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		Prefix:          "uploads",
		UseSSL:          false,
		ForcePathStyle:  true,
	})
	if err != nil {
		t.Fatalf("NewS3FileStorageRepository() error = %v", err)
	}
	if repo == nil {
		t.Fatal("NewS3FileStorageRepository() repo = nil, want non-nil")
	}
}

type failingReader struct {
	firstChunk []byte
	err        error
	readOnce   bool
}

func (r *failingReader) Read(p []byte) (int, error) {
	if !r.readOnce {
		r.readOnce = true
		n := copy(p, r.firstChunk)
		return n, nil
	}
	return 0, r.err
}

var _ io.Reader = (*failingReader)(nil)

type fakeS3Client struct {
	putBucket    string
	putKey       string
	putBody      string
	putErr       error
	deleteBucket string
	deleteKey    string
	deleteErr    error
}

func (f *fakeS3Client) PutObject(
	_ context.Context,
	params *s3.PutObjectInput,
	_ ...func(*s3.Options),
) (*s3.PutObjectOutput, error) {
	if f.putErr != nil {
		return nil, f.putErr
	}

	if params.Bucket != nil {
		f.putBucket = *params.Bucket
	}
	if params.Key != nil {
		f.putKey = *params.Key
	}
	if params.Body != nil {
		body, err := io.ReadAll(params.Body)
		if err != nil {
			return nil, err
		}
		f.putBody = string(body)
	}

	return &s3.PutObjectOutput{}, nil
}

func (f *fakeS3Client) DeleteObject(
	_ context.Context,
	params *s3.DeleteObjectInput,
	_ ...func(*s3.Options),
) (*s3.DeleteObjectOutput, error) {
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}

	if params.Bucket != nil {
		f.deleteBucket = *params.Bucket
	}
	if params.Key != nil {
		f.deleteKey = *params.Key
	}

	return &s3.DeleteObjectOutput{}, nil
}

var _ s3ObjectClient = (*fakeS3Client)(nil)

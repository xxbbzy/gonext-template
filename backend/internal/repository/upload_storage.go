package repository

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
		baseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
	}, nil
}

// SaveFile persists file content using a precomputed stored name.
func (r *LocalFileStorageRepository) SaveFile(_ context.Context, storedName string, src io.Reader) error {
	safeStoredName, err := sanitizeStoredName(storedName)
	if err != nil {
		return err
	}

	path := filepath.Join(r.uploadDir, safeStoredName)

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
	return r.baseURL + "/uploads/" + url.PathEscape(storedName)
}

type s3ObjectClient interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// S3FileStorageConfig contains settings for S3-compatible object storage.
type S3FileStorageConfig struct {
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Prefix          string
	UseSSL          bool
	ForcePathStyle  bool
	PublicBaseURL   string
}

// S3FileStorageRepository persists files in S3-compatible object storage.
type S3FileStorageRepository struct {
	client         s3ObjectClient
	bucket         string
	region         string
	endpoint       string
	prefix         string
	useSSL         bool
	forcePathStyle bool
	publicBaseURL  string
}

// NewS3FileStorageRepository creates an S3-compatible file storage repository.
func NewS3FileStorageRepository(_ context.Context, cfg S3FileStorageConfig) (*S3FileStorageRepository, error) {
	normalizedCfg, err := normalizeS3FileStorageConfig(cfg)
	if err != nil {
		return nil, err
	}

	staticCredentials := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
		normalizedCfg.AccessKeyID,
		normalizedCfg.SecretAccessKey,
		"",
	))

	awsCfg := aws.Config{
		Region:      normalizedCfg.Region,
		Credentials: staticCredentials,
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = normalizedCfg.ForcePathStyle
		if normalizedCfg.Endpoint != "" {
			o.BaseEndpoint = &normalizedCfg.Endpoint
		}
	})

	return newS3FileStorageRepositoryWithClient(client, normalizedCfg), nil
}

func newS3FileStorageRepositoryWithClient(client s3ObjectClient, cfg S3FileStorageConfig) *S3FileStorageRepository {
	return &S3FileStorageRepository{
		client:         client,
		bucket:         cfg.Bucket,
		region:         cfg.Region,
		endpoint:       cfg.Endpoint,
		prefix:         cfg.Prefix,
		useSSL:         cfg.UseSSL,
		forcePathStyle: cfg.ForcePathStyle,
		publicBaseURL:  cfg.PublicBaseURL,
	}
}

// SaveFile stores a file in an S3-compatible bucket.
func (r *S3FileStorageRepository) SaveFile(ctx context.Context, storedName string, src io.Reader) error {
	safeStoredName, err := sanitizeStoredName(storedName)
	if err != nil {
		return err
	}

	key := buildObjectKey(r.prefix, safeStoredName)
	if _, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &r.bucket,
		Key:    &key,
		Body:   src,
	}); err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

// DeleteFile removes a file from an S3-compatible bucket.
func (r *S3FileStorageRepository) DeleteFile(ctx context.Context, storedName string) error {
	safeStoredName, err := sanitizeStoredName(storedName)
	if err != nil {
		return err
	}

	key := buildObjectKey(r.prefix, safeStoredName)
	if _, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &r.bucket,
		Key:    &key,
	}); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// GetFileURL returns the public URL of an uploaded object.
func (r *S3FileStorageRepository) GetFileURL(storedName string) string {
	safeStoredName, err := sanitizeStoredName(storedName)
	if err != nil {
		log.Printf("repository: S3FileStorageRepository.GetFileURL sanitizeStoredName failed: %v", err)
		return ""
	}

	objectKey := buildObjectKey(r.prefix, safeStoredName)
	escapedKey := escapeObjectKey(objectKey)
	if r.publicBaseURL != "" {
		return r.publicBaseURL + "/" + escapedKey
	}

	if r.endpoint != "" {
		if r.forcePathStyle {
			return fmt.Sprintf("%s/%s/%s", r.endpoint, r.bucket, escapedKey)
		}
		endpointURL, err := url.Parse(r.endpoint)
		if err == nil && endpointURL.Host != "" {
			endpointURL.Host = r.bucket + "." + endpointURL.Host
			endpointURL.Path = joinURLPath(endpointURL.Path, objectKey)
			endpointURL.RawPath = joinURLPath(endpointURL.EscapedPath(), escapedKey)
			return endpointURL.String()
		}
	}

	scheme := "https"
	if !r.useSSL {
		scheme = "http"
	}
	if r.forcePathStyle {
		return fmt.Sprintf("%s://s3.%s.amazonaws.com/%s/%s", scheme, r.region, r.bucket, escapedKey)
	}
	return fmt.Sprintf("%s://%s.s3.%s.amazonaws.com/%s", scheme, r.bucket, r.region, escapedKey)
}

func normalizeS3FileStorageConfig(cfg S3FileStorageConfig) (S3FileStorageConfig, error) {
	normalized := cfg
	normalized.Bucket = strings.TrimSpace(cfg.Bucket)
	normalized.Region = strings.TrimSpace(cfg.Region)
	normalized.AccessKeyID = strings.TrimSpace(cfg.AccessKeyID)
	normalized.SecretAccessKey = strings.TrimSpace(cfg.SecretAccessKey)
	normalized.Endpoint = strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	normalized.Prefix = strings.Trim(cfg.Prefix, "/ ")
	normalized.PublicBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")

	if normalized.Bucket == "" {
		return S3FileStorageConfig{}, fmt.Errorf("s3 bucket must be non-empty")
	}
	if normalized.Region == "" {
		return S3FileStorageConfig{}, fmt.Errorf("s3 region must be non-empty")
	}
	if normalized.AccessKeyID == "" {
		return S3FileStorageConfig{}, fmt.Errorf("s3 access key id must be non-empty")
	}
	if normalized.SecretAccessKey == "" {
		return S3FileStorageConfig{}, fmt.Errorf("s3 secret access key must be non-empty")
	}

	return normalized, nil
}

func joinURLPath(basePath, suffix string) string {
	if basePath == "" || basePath == "/" {
		return "/" + strings.TrimLeft(suffix, "/")
	}

	return strings.TrimRight(basePath, "/") + "/" + strings.TrimLeft(suffix, "/")
}

func buildObjectKey(prefix, storedName string) string {
	if prefix == "" {
		return storedName
	}
	return prefix + "/" + storedName
}

func escapeObjectKey(objectKey string) string {
	if objectKey == "" {
		return ""
	}

	parts := strings.Split(objectKey, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
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

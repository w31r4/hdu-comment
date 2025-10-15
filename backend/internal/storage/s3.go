package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config holds configuration for S3 compatible storage.
type S3Config struct {
	Endpoint  string
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	BaseURL   string
}

// S3 implements FileStorage backed by an S3-compatible service.
type S3 struct {
	client  *minio.Client
	bucket  string
	baseURL string
}

// NewS3 creates an S3 storage provider based on configuration.
func NewS3(cfg S3Config) (*S3, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 storage requires endpoint and bucket")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}

	return &S3{client: client, bucket: cfg.Bucket, baseURL: cfg.BaseURL}, nil
}

// Save uploads content to the configured bucket.
func (s *S3) Save(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (FileInfo, error) {
	opts := minio.PutObjectOptions{ContentType: contentType}
	if contentType == "" {
		opts.ContentType = "application/octet-stream"
	}
	if size < 0 {
		size = -1
	}

	if _, err := s.client.PutObject(ctx, s.bucket, key, reader, size, opts); err != nil {
		return FileInfo{}, err
	}

	url, err := s.objectURL(key)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{Key: key, URL: url}, nil
}

// Delete removes an object from the configured bucket.
func (s *S3) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		if resp := minio.ToErrorResponse(err); resp.Code == "NoSuchKey" {
			return nil
		}
		return err
	}
	return nil
}

func (s *S3) objectURL(key string) (string, error) {
	if s.baseURL != "" {
		return resolveURL(s.baseURL, key)
	}

	endpoint := s.client.EndpointURL()
	if endpoint == nil {
		return "", fmt.Errorf("s3 endpoint not configured")
	}

	if endpoint.Scheme == "" {
		endpoint.Scheme = "https"
	}

	endpoint.Path = path.Join(endpoint.Path, s.bucket, key)
	if !strings.HasPrefix(endpoint.Path, "/") {
		endpoint.Path = "/" + endpoint.Path
	}

	return endpoint.String(), nil
}

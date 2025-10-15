package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/hdu-dp/backend/internal/config"
)

// FileInfo describes an uploaded file result.
type FileInfo struct {
	Key string
	URL string
}

// UploadFile represents metadata required to persist a file.
type UploadFile struct {
	Reader      io.ReadCloser
	Size        int64
	Filename    string
	ContentType string
}

// FileStorage describes a generic storage backend.
type FileStorage interface {
	Save(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (FileInfo, error)
	Delete(ctx context.Context, key string) error
}

// New creates a storage implementation based on configuration.
func New(cfg *config.Config) (FileStorage, error) {
	switch strings.ToLower(cfg.Storage.Provider) {
	case "local", "":
		publicBase := cfg.Storage.PublicBaseURL
		if publicBase == "" {
			publicBase = "/api/v1/uploads"
		}
		return NewLocal(cfg.Storage.UploadDir, publicBase)
	case "s3":
		return NewS3(S3Config{
			Endpoint:  cfg.Storage.S3.Endpoint,
			Bucket:    cfg.Storage.S3.Bucket,
			Region:    cfg.Storage.S3.Region,
			AccessKey: cfg.Storage.S3.AccessKey,
			SecretKey: cfg.Storage.S3.SecretKey,
			UseSSL:    cfg.Storage.S3.UseSSL,
			BaseURL:   cfg.Storage.S3.BaseURL,
		})
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Storage.Provider)
	}
}

func resolveURL(base, key string) (string, error) {
	if base == "" {
		// treat key as absolute
		if strings.HasPrefix(key, "http") {
			return key, nil
		}
		return "/" + strings.TrimLeft(key, "/"), nil
	}

	if strings.HasSuffix(base, "/") {
		base = strings.TrimSuffix(base, "/")
	}

	if strings.HasPrefix(base, "http") {
		u, err := url.Parse(base)
		if err != nil {
			return "", err
		}
		u.Path = path.Join(u.Path, key)
		return u.String(), nil
	}

	return path.Join(base, filepath.ToSlash(key)), nil
}

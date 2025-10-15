package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Local implements FileStorage by writing to the local filesystem.
type Local struct {
	baseDir    string
	publicBase string
}

// NewLocal creates a new Local storage provider.
func NewLocal(dir, publicBase string) (*Local, error) {
	if dir == "" {
		dir = "uploads"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Local{baseDir: dir, publicBase: publicBase}, nil
}

// Save persists a file to local storage.
func (l *Local) Save(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (FileInfo, error) {
	dstPath := filepath.Join(l.baseDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return FileInfo{}, err
	}

	f, err := os.Create(dstPath)
	if err != nil {
		return FileInfo{}, err
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return FileInfo{}, err
	}

	url, err := resolveURL(l.publicBase, key)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{Key: key, URL: url}, nil
}

// Delete removes a stored object if present.
func (l *Local) Delete(ctx context.Context, key string) error {
	path := filepath.Join(l.baseDir, filepath.FromSlash(key))
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}

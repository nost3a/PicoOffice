package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage local disk; key is relative to root
type LocalStorage struct {
	root string
}

// NewLocal builds local storage
func NewLocal(root string) *LocalStorage {
	os.MkdirAll(root, 0o755)
	return &LocalStorage{root: root}
}

func (l *LocalStorage) path(key string) string {
	return filepath.Join(l.root, key)
}

// Put writes file, returns accessible key
func (l *LocalStorage) Put(ctx context.Context, key string, r io.Reader, size int64) (string, error) {
	p := l.path(key)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return key, nil
}

// Get opens file for reading
func (l *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open(l.path(key))
}

// Delete removes file
func (l *LocalStorage) Delete(ctx context.Context, key string) error {
	return os.Remove(l.path(key))
}

// Stat returns file size
func (l *LocalStorage) Stat(ctx context.Context, key string) (int64, error) {
	st, err := os.Stat(l.path(key))
	if err != nil {
		return 0, err
	}
	if st.IsDir() {
		return 0, errors.New("is a directory")
	}
	return st.Size(), nil
}

// Kind backend type
func (l *LocalStorage) Kind() string { return "local" }

// Package storage: object storage abstraction; local/OSS/S3 pluggable.
package storage

import (
	"context"
	"errors"
	"io"
)

// Storage unified object access interface
type Storage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64) (string, error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Stat(ctx context.Context, key string) (int64, error)
	Kind() string
}

// Init builds Storage by kind; opts are backend-specific
func Init(kind string, opts map[string]string) (Storage, error) {
	switch kind {
	case "local":
		root := opts["root"]
		if root == "" {
			return nil, errors.New("local storage requires root")
		}
		return NewLocal(root), nil
	case "s3":
		return NewS3(opts)
	default:
		return nil, errors.New("unknown storage kind: " + kind)
	}
}

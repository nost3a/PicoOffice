package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Storage S3/OSS/MinIO backend via minio-go/v7.
type S3Storage struct {
	client *minio.Client
	bucket string
}

// NewS3 builds S3 backend from opts.
// supported keys: endpoint / access_key / secret_key / bucket / region / ssl("true"|"1").
// constructor is offline; connectivity errors surface at Put/Get.
func NewS3(opts map[string]string) (Storage, error) {
	endpoint := opts["endpoint"]
	ak := opts["access_key"]
	sk := opts["secret_key"]
	bucket := opts["bucket"]
	region := opts["region"]

	if endpoint == "" || ak == "" || sk == "" || bucket == "" {
		return nil, errors.New("s3 requires endpoint/access_key/secret_key/bucket")
	}

	useSSL := opts["ssl"] == "true" || opts["ssl"] == "1"
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(ak, sk, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("s3 new client: %w", err)
	}
	return &S3Storage{client: cli, bucket: bucket}, nil
}

// Put uploads object, returns key
func (s *S3Storage) Put(ctx context.Context, key string, r io.Reader, size int64) (string, error) {
	if _, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{}); err != nil {
		return "", err
	}
	return key, nil
}

// Get downloads object, returns lazy-error ReadCloser (minio.Object)
func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
}

// Delete removes object
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

// Stat returns object size
func (s *S3Storage) Stat(ctx context.Context, key string) (int64, error) {
	info, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

// Kind backend type
func (s *S3Storage) Kind() string { return "s3" }

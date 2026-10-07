package storage

import (
	"bytes"
	"context"
	"io"
	"testing"
)

// fakeStorage in-memory fake to verify Storage contract and callers.
type fakeStorage struct {
	data map[string][]byte
}

var _ Storage = (*fakeStorage)(nil) // compile-time interface assertion

func newFake() *fakeStorage { return &fakeStorage{data: map[string][]byte{}} }

func (f *fakeStorage) Put(_ context.Context, key string, r io.Reader, size int64) (string, error) {
	b, _ := io.ReadAll(r)
	f.data[key] = b
	return key, nil
}
func (f *fakeStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.data[key])), nil
}
func (f *fakeStorage) Delete(_ context.Context, key string) error {
	delete(f.data, key)
	return nil
}
func (f *fakeStorage) Stat(_ context.Context, key string) (int64, error) {
	return int64(len(f.data[key])), nil
}
func (f *fakeStorage) Kind() string { return "fake" }

func TestFakeStorageRoundTrip(t *testing.T) {
	s := newFake()
	ctx := context.Background()
	body := []byte("hello pico")

	if k, err := s.Put(ctx, "a/b.txt", bytes.NewReader(body), int64(len(body))); err != nil || k != "a/b.txt" {
		t.Fatalf("put: %v %s", err, k)
	}
	if n, _ := s.Stat(ctx, "a/b.txt"); n != int64(len(body)) {
		t.Fatalf("stat = %d, want %d", n, len(body))
	}
	rc, err := s.Get(ctx, "a/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(rc)
	rc.Close()
	if string(got) != "hello pico" {
		t.Fatalf("get = %q", got)
	}
	if err := s.Delete(ctx, "a/b.txt"); err != nil {
		t.Fatal(err)
	}
	if n, _ := s.Stat(ctx, "a/b.txt"); n != 0 {
		t.Fatalf("after delete stat = %d, want 0", n)
	}
}

// S3 param validation: missing args must error; valid args build client (offline).
func TestNewS3Validation(t *testing.T) {
	if _, err := NewS3(map[string]string{}); err == nil {
		t.Fatal("expected error when s3 opts empty")
	}
	s, err := NewS3(map[string]string{
		"endpoint": "localhost:9000",
		"access_key": "ak",
		"secret_key": "sk",
		"bucket": "pico",
		"region": "us-east-1",
	})
	if err != nil {
		t.Fatalf("valid s3 opts should construct: %v", err)
	}
	if s.Kind() != "s3" {
		t.Fatalf("kind = %s", s.Kind())
	}
}

// Init routes to s3/local
func TestInitKindDispatch(t *testing.T) {
	if _, err := Init("local", map[string]string{"root": t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	if _, err := Init("nope", nil); err == nil {
		t.Fatal("unknown kind should error")
	}
}

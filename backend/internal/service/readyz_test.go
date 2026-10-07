package service

import "testing"

// free disk check: temp dir volume must have enough space.
func TestDiskFreeAndHealthy(t *testing.T) {
	dir := t.TempDir()
	avail, err := DiskFreeBytes(dir)
	if err != nil {
		t.Fatalf("DiskFreeBytes: %v", err)
	}
	if avail <= 0 {
		t.Fatalf("avail bytes = %d, want > 0", avail)
	}
	ok, err := DiskHealthy(dir)
	if err != nil {
		t.Fatalf("DiskHealthy: %v", err)
	}
	if !ok {
		t.Fatalf("temp dir disk should be healthy (avail=%d, need>=%d)", avail, MinFreeDiskBytes)
	}
}

// missing path must error (readyz uses this to mark disk fail)
func TestDiskCheckMissingPath(t *testing.T) {
	if _, err := DiskFreeBytes("/nonexistent/picooffice/path/xyz"); err == nil {
		t.Fatal("expected error for missing path")
	}
}

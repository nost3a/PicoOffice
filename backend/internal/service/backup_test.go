package service

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBackupPackaging(t *testing.T) {
	dir := t.TempDir()

	// build a sqlite with tables and rows
	dbPath := filepath.Join(dir, "test.db")
	gdb, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	type Row struct {
		ID   int
		Name string
	}
	if err := gdb.AutoMigrate(&Row{}); err != nil {
		t.Fatal(err)
	}
	gdb.Create(&Row{Name: "hello"})
	if sqlDB, err := gdb.DB(); err == nil {
		sqlDB.Close()
	}

	// storage dir has a file
	storageDir := filepath.Join(dir, "storage")
	_ = os.MkdirAll(storageDir, 0o755)
	_ = os.WriteFile(filepath.Join(storageDir, "a.txt"), []byte("data"), 0o644)

	out := filepath.Join(dir, "backup.tar.gz")
	if err := Backup(dbPath, storageDir, out); err != nil {
		t.Fatalf("backup: %v", err)
	}

	// read tar.gz to verify entries
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gw, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gw)

	names := map[string]bool{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		names[hdr.Name] = true
		if hdr.Name == "picooffice.db" {
			if hdr.Size == 0 {
				t.Fatal("snapshot db entry is empty")
			}
			buf := make([]byte, 16)
			if _, err := io.ReadFull(tr, buf); err != nil {
				t.Fatal(err)
			}
			if string(buf[:15]) != "SQLite format 3" {
				t.Fatalf("snapshot is not a sqlite db: %q", buf)
			}
		}
	}
	if !names["picooffice.db"] {
		t.Fatal("tar missing picooffice.db")
	}
	if !names["storage/a.txt"] {
		t.Fatal("tar missing storage/a.txt")
	}
}

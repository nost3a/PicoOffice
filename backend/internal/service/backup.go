package service

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Backup snapshots SQLite (VACUUM INTO) + storage dir into outPath tar.gz.
// online backup: VACUUM INTO consistent snapshot (no lock on source); bundle with assets.
func Backup(dbPath, storageDir, outPath string) error {
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("db not found: %w", err)
	}

	// 1. open source db, busy_timeout avoids write contention
	gdb, err := gorm.Open(sqlite.Open(dbPath+"?_pragma=busy_timeout(5000)"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	// 2. VACUUM INTO a temp snapshot (target must not exist)
	tmpDir, err := os.MkdirTemp("", "pico-backup-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	snapDB := filepath.Join(tmpDir, "picooffice-snapshot.db")
	quoted := strings.ReplaceAll(snapDB, "'", "''")
	if _, err := sqlDB.Exec("VACUUM INTO '" + quoted + "'"); err != nil {
		return fmt.Errorf("vacuum into: %w", err)
	}

	// 3. tar.gz: write snapshot db first, then storage dir
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	closeAll := func() error {
		err1 := tw.Close()
		err2 := gw.Close()
		err3 := f.Close()
		if err1 != nil {
			return err1
		}
		if err2 != nil {
			return err2
		}
		return err3
	}

	if err := addFileToTar(tw, snapDB, "picooffice.db"); err != nil {
		closeAll()
		return err
	}

	if storageDir != "" {
		walkErr := filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, rerr := filepath.Rel(storageDir, path)
			if rerr != nil {
				return rerr
			}
			return addFileToTar(tw, path, filepath.Join("storage", rel))
		})
		if walkErr != nil {
			closeAll()
			return walkErr
		}
	}

	return closeAll()
}

// addFileToTar writes src as name into tar stream
func addFileToTar(tw *tar.Writer, src, name string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	hdr := &tar.Header{
		Name:     name,
		Mode:     0o644,
		Size:     st.Size(),
		ModTime:  st.ModTime(),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

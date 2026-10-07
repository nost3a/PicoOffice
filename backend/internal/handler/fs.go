package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/model"
)

// allowed upload ext whitelist
var allowedExt = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".txt": true, ".md": true, ".png": true,
	".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".csv": true,
	".zip": true, ".mp4": true,
}

const maxFileSize = 104857600 // max 100MB per file

// checkFileType checks ext against whitelist
func checkFileType(filename string) bool {
	return allowedExt[strings.ToLower(filepath.Ext(filename))]
}

// touchQuota adds used bytes after successful upload
func touchQuota(uid uint, delta int64) {
	db.Model(&model.User{}).Where("id = ?", uid).
		UpdateColumn("used_bytes", gorm.Expr("used_bytes + ?", delta))
}

// uniqueFileKey builds object key files/<uid>/<name>; short random suffix on collision.
// objStore nil (pure local fallback) returns candidate key.
// note: S3 dead endpoint makes Stat fail; treat as "missing", keep candidate key; Put error surfaces upward.
func uniqueFileKey(ctx context.Context, uid uint, filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	key := fmt.Sprintf("files/%d/%s%s", uid, name, ext)
	if objStore == nil {
		return key
	}
	if _, err := objStore.Stat(ctx, key); err != nil {
		// missing/undecidable (s3 dead endpoint) falls back to candidate key
		return key
	}
	// exists, append 4 random bytes hex
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("files/%d/%s-%s%s", uid, name, hex.EncodeToString(b), ext)
}

func ListFiles(c *gin.Context) {
	uid := c.GetUint("uid")
	var list []model.File
	db.Where("uploader_id = ?", uid).Order("id DESC").Find(&list)
	c.JSON(200, gin.H{"items": list})
}

// UploadFile direct multipart upload to storage/uploads/
func UploadFile(c *gin.Context) {
	uid := c.GetUint("uid")
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "multipart field 'file' required"})
		return
	}
	// check type and size first
	if !checkFileType(fh.Filename) {
		c.JSON(400, gin.H{"error": "file type not allowed"})
		return
	}
	if fh.Size > maxFileSize {
		c.JSON(413, gin.H{"error": "file too large"})
		return
	}
	// then check quota
	var u model.User
	if err := db.First(&u, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	if u.UsedBytes+fh.Size > u.QuotaBytes {
		c.JSON(413, gin.H{"error": "quota exceeded"})
		return
	}

	// validation done, start write. Open multipart stream; object store and local fallback share it.
	in, err := fh.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "open upload"})
		return
	}

	var storedPath string
	if objStore != nil {
		// go through object store (local: relative key under storageDir; s3: object store)
		key := uniqueFileKey(c.Request.Context(), uid, fh.Filename)
		if _, err := objStore.Put(c.Request.Context(), key, in, fh.Size); err != nil {
			in.Close()
			c.JSON(500, gin.H{"error": "upload to storage: " + err.Error()})
			return
		}
		in.Close()
		storedPath = key
	} else {
		// pure local fallback (objStore nil): legacy write to storageDir/uploads
		dir := filepath.Join(storageDir, "uploads")
		os.MkdirAll(dir, 0o755)
		stored := fmt.Sprintf("%d-%s", uid, filepath.Base(fh.Filename))
		dst := filepath.Join(dir, stored)
		out, err := os.Create(dst)
		if err != nil {
			in.Close()
			c.JSON(500, gin.H{"error": "create file"})
			return
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			in.Close()
			c.JSON(500, gin.H{"error": "save file"})
			return
		}
		out.Close()
		in.Close()
		storedPath = dst
	}

	f := model.File{
		UploaderID:   uid,
		OriginalName: fh.Filename,
		StoredPath:   storedPath,
		Size:         fh.Size,
		Mime:         fh.Header.Get("Content-Type"),
	}
	db.Create(&f)
	touchQuota(uid, f.Size)
	c.JSON(200, f)
}

// DeleteFile removes own upload and restores quota
func DeleteFile(c *gin.Context) {
	uid := c.GetUint("uid")
	id, _ := strconv.Atoi(c.Param("id"))
	var f model.File
	if err := db.First(&f, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "file not found"})
		return
	}
	if f.UploaderID != uid {
		c.JSON(403, gin.H{"error": "not your file"})
		return
	}
	// delete object first (failure does not block DB cleanup; s3 dead endpoint)
	if objStore != nil {
		if err := objStore.Delete(c.Request.Context(), f.StoredPath); err != nil {
			// log only, continue DB/quota cleanup
			fmt.Printf("[warn] objStore delete %s failed: %v\n", f.StoredPath, err)
		}
	} else {
		os.Remove(f.StoredPath)
	}
	db.Delete(&f)
	touchQuota(uid, -f.Size)
	c.JSON(200, gin.H{"ok": true})
}

func DownloadFile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var f model.File
	if err := db.First(&f, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "file not found"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", f.OriginalName))
	if objStore != nil {
		// stream from object store (local: relative key under storageDir; s3: object store)
		rc, err := objStore.Get(c.Request.Context(), f.StoredPath)
		if err != nil {
			c.JSON(404, gin.H{"error": "file not found in storage"})
			return
		}
		defer rc.Close()
		c.DataFromReader(200, f.Size, guessMime(f.Mime), rc, nil)
		return
	}
	c.File(f.StoredPath)
}

// guessMime fallback Content-Type; unknown -> application/octet-stream
func guessMime(mime string) string {
	if mime != "" {
		return mime
	}
	return "application/octet-stream"
}

// ---- chunked upload ----

// uploadMeta per-upload_id metadata in tmp dir
type uploadMeta struct {
	UploadID   string `json:"upload_id"`
	UserID     uint   `json:"user_id"`
	Filename   string `json:"filename"`
	TotalSize  int64  `json:"total_size"`
	TotalChunks int   `json:"total_chunks"`
	Mime       string `json:"mime"`
}

func tmpDirOf(uploadID string) string {
	return filepath.Join(storageDir, "tmp", uploadID)
}

func readMeta(uploadID string) (*uploadMeta, error) {
	data, err := os.ReadFile(filepath.Join(tmpDirOf(uploadID), "meta.json"))
	if err != nil {
		return nil, err
	}
	var m uploadMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// listChunks scans tmp dir, returns received chunk indices
func listChunks(m *uploadMeta) []int {
	dir := tmpDirOf(m.UploadID)
	out := []int{}
	for i := 0; i < m.TotalChunks; i++ {
		if _, err := os.Stat(filepath.Join(dir, strconv.Itoa(i))); err == nil {
			out = append(out, i)
		}
	}
	return out
}

// newUploadID 16 random bytes hex
func newUploadID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// UploadInit opens an upload session
func UploadInit(c *gin.Context) {
	uid := c.GetUint("uid")
	var req struct {
		Filename    string `json:"filename"`
		TotalSize   int64  `json:"total_size"`
		TotalChunks int    `json:"total_chunks"`
		Mime        string `json:"mime"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	if req.Filename == "" || req.TotalSize <= 0 || req.TotalChunks <= 0 {
		c.JSON(400, gin.H{"error": "bad params"})
		return
	}
	if !checkFileType(req.Filename) {
		c.JSON(400, gin.H{"error": "file type not allowed"})
		return
	}
	if req.TotalSize > maxFileSize {
		c.JSON(413, gin.H{"error": "file too large"})
		return
	}
	// reject quota overflow at init, avoid wasted upload
	var u model.User
	if err := db.First(&u, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	if u.UsedBytes+req.TotalSize > u.QuotaBytes {
		c.JSON(413, gin.H{"error": "quota exceeded"})
		return
	}

	uploadID := newUploadID()
	dir := tmpDirOf(uploadID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(500, gin.H{"error": "make tmp dir"})
		return
	}
	m := uploadMeta{
		UploadID:    uploadID,
		UserID:      uid,
		Filename:    filepath.Base(req.Filename),
		TotalSize:   req.TotalSize,
		TotalChunks: req.TotalChunks,
		Mime:        req.Mime,
	}
	data, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), data, 0o644); err != nil {
		c.JSON(500, gin.H{"error": "write meta"})
		return
	}
	c.JSON(200, gin.H{"upload_id": uploadID})
}

// UploadChunk receives one chunk; query has upload_id and index
func UploadChunk(c *gin.Context) {
	uid := c.GetUint("uid")
	uploadID := c.Query("upload_id")
	index, _ := strconv.Atoi(c.Query("index"))
	if uploadID == "" || index < 0 {
		c.JSON(400, gin.H{"error": "bad params"})
		return
	}
	m, err := readMeta(uploadID)
	if err != nil {
		c.JSON(400, gin.H{"error": "upload not found"})
		return
	}
	if m.UserID != uid {
		c.JSON(403, gin.H{"error": "not your upload"})
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "multipart field 'file' required"})
		return
	}
	// write directly as {index}
	src, err := fh.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "open chunk"})
		return
	}
	defer src.Close()
	dst, err := os.Create(filepath.Join(tmpDirOf(uploadID), strconv.Itoa(index)))
	if err != nil {
		c.JSON(500, gin.H{"error": "save chunk"})
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(500, gin.H{"error": "save chunk"})
		return
	}
	c.JSON(200, gin.H{"received": listChunks(m)})
}

// UploadStatus reports missing chunks for resumable upload
func UploadStatus(c *gin.Context) {
	uid := c.GetUint("uid")
	uploadID := c.Query("upload_id")
	m, err := readMeta(uploadID)
	if err != nil {
		c.JSON(400, gin.H{"error": "upload not found"})
		return
	}
	if m.UserID != uid {
		c.JSON(403, gin.H{"error": "not your upload"})
		return
	}
	c.JSON(200, gin.H{"received_chunks": listChunks(m)})
}

// UploadComplete assembles chunks into final file
func UploadComplete(c *gin.Context) {
	uid := c.GetUint("uid")
	var req struct {
		UploadID string `json:"upload_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UploadID == "" {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	m, err := readMeta(req.UploadID)
	if err != nil {
		c.JSON(400, gin.H{"error": "upload not found"})
		return
	}
	if m.UserID != uid {
		c.JSON(403, gin.H{"error": "not your upload"})
		return
	}
	// ensure all chunks present first
	if len(listChunks(m)) != m.TotalChunks {
		c.JSON(400, gin.H{"error": "chunks incomplete"})
		return
	}

	// assemble chunks by index into staging file (local tmp, then object store takes over)
	staging := filepath.Join(tmpDirOf(m.UploadID), "final.bin")
	out, err := os.Create(staging)
	if err != nil {
		c.JSON(500, gin.H{"error": "create staging"})
		return
	}
	// assemble by index order
	for i := 0; i < m.TotalChunks; i++ {
		chunk, err := os.Open(filepath.Join(tmpDirOf(m.UploadID), strconv.Itoa(i)))
		if err != nil {
			out.Close()
			os.Remove(staging)
			c.JSON(500, gin.H{"error": "read chunk"})
			return
		}
		if _, err := io.Copy(out, chunk); err != nil {
			chunk.Close()
			out.Close()
			os.Remove(staging)
			c.JSON(500, gin.H{"error": "concat"})
			return
		}
		chunk.Close()
	}
	out.Close()

	// drop on size mismatch
	fi, _ := os.Stat(staging)
	if fi.Size() != m.TotalSize {
		os.Remove(staging)
		c.JSON(400, gin.H{"error": "size mismatch"})
		return
	}

	// persist: object store first, else legacy storageDir/uploads/<uid>/<name>
	var storedPath string
	if objStore != nil {
		key := uniqueFileKey(c.Request.Context(), uid, m.Filename)
		fin, err := os.Open(staging)
		if err != nil {
			os.Remove(staging)
			c.JSON(500, gin.H{"error": "open staging"})
			return
		}
		if _, err := objStore.Put(c.Request.Context(), key, fin, fi.Size()); err != nil {
			fin.Close()
			os.Remove(staging)
			c.JSON(500, gin.H{"error": "upload to storage: " + err.Error()})
			return
		}
		fin.Close()
		os.Remove(staging)
		storedPath = key
	} else {
		userDir := filepath.Join(storageDir, "uploads", strconv.Itoa(int(uid)))
		os.MkdirAll(userDir, 0o755)
		dst := filepath.Join(userDir, m.Filename)
		if err := os.Rename(staging, dst); err != nil {
			c.JSON(500, gin.H{"error": "move staging"})
			return
		}
		storedPath = dst
	}

	f := model.File{
		UploaderID:   uid,
		OriginalName:  m.Filename,
		StoredPath:    storedPath,
		Size:          fi.Size(),
		Mime:          m.Mime,
	}
	if err := db.Create(&f).Error; err != nil {
		os.Remove(storedPath) // local mode uses absolute path; object store mode is best-effort
		if objStore != nil {
			objStore.Delete(c.Request.Context(), storedPath)
		}
		c.JSON(500, gin.H{"error": "save record"})
		return
	}
	touchQuota(uid, f.Size)
	os.RemoveAll(tmpDirOf(m.UploadID))
	c.JSON(200, gin.H{"file": f})
}

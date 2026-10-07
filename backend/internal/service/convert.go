package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/model"
)

// sofficeProfile uses a fresh user profile per conversion to avoid lock conflicts
func sofficeProfile() string {
	return fmt.Sprintf("file:///tmp/pico-lo-profile-%d", time.Now().UnixNano())
}

// ConvertFile runs soffice --headless --convert-to <filter> --outdir <dir> <src>,
// expect output <src-without-ext>.<outExt>; return absolute path.
func ConvertFile(dir, src, filter, outExt string) (string, error) {
	args := []string{
		"--headless", "--norestore", "--nologo", "--nofirststartwizard",
		"-env:UserInstallation=" + sofficeProfile(),
		"--convert-to", filter,
		"--outdir", dir,
		src,
	}
	cmd := exec.Command("soffice", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("soffice %s: %v: %s", filter, err, strings.TrimSpace(string(out)))
	}
	base := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
	outPath := filepath.Join(dir, base+"."+outExt)
	if _, err := os.Stat(outPath); err != nil {
		// list dir for debugging
		entries, _ := os.ReadDir(dir)
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		return "", fmt.Errorf("输出文件 %s 不存在，目录内容: %v", outPath, names)
	}
	return outPath, nil
}

// ImportToHTML converts uploaded office file to html, returns html path.
// docx/xlsx -> html directly; pptx tries direct html, falls back via odp.
func ImportToHTML(dir, src string) (string, error) {
	if p, err := ConvertFile(dir, src, "html", "html"); err == nil {
		return p, nil
	}
	// pptx fallback: html -> odp -> html
	odp, err := ConvertFile(dir, src, "odp", "odp")
	if err != nil {
		return "", err
	}
	return ConvertFile(dir, odp, "html:impress_web_publishing", "html")
}

// RecalcUsedBytes recomputes used bytes for a user and persists:
//  len of content for non-deleted docs
//  + len of doc_versions.content for these docs
//  + sum of uploaded files.size
//
// called after save/delete/purge; recompute to avoid drift.
func RecalcUsedBytes(db *gorm.DB, userID uint) {
	var docContent int64
	db.Model(&model.Doc{}).
		Where("owner_id = ?", userID).
		Where("deleted_at IS NULL").
		Select("COALESCE(SUM(LENGTH(content)),0)").
		Scan(&docContent)

	var verContent int64
	db.Table("doc_versions").
		Joins("JOIN docs ON docs.id = doc_versions.doc_id").
		Where("docs.owner_id = ? AND docs.deleted_at IS NULL", userID).
		Select("COALESCE(SUM(LENGTH(doc_versions.content)),0)").
		Scan(&verContent)

	var fileSize int64
	db.Model(&model.File{}).
		Where("uploader_id = ?", userID).
		Select("COALESCE(SUM(size),0)").
		Scan(&fileSize)

	total := docContent + verContent + fileSize
	db.Model(&model.User{}).Where("id = ?", userID).Update("used_bytes", total)
}

// CleanExpiredTrash purges soft-deleted docs older than 30 days (incl. version snapshots),
// recompute quota for each affected user. main.go is fixed, so only expose the function,
// called by external cron or future scheduler. Returns purged doc count.
func CleanExpiredTrash(db *gorm.DB) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -30)
	var docs []model.Doc
	if err := db.Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
		Find(&docs).Error; err != nil {
		return 0, err
	}
	var n int64
	affected := map[uint]struct{}{}
	for i := range docs {
		d := docs[i]
		db.Unscoped().Where("doc_id = ?", d.ID).Delete(&model.DocVersion{})
		if err := db.Unscoped().Delete(&d).Error; err == nil {
			n++
			affected[d.OwnerID] = struct{}{}
		}
	}
	for uid := range affected {
		RecalcUsedBytes(db, uid)
	}
	return n, nil
}

package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ExportJob external view of one async export (no sensitive/large fields)
type ExportJob struct {
	ID     string `json:"id"`
	DocID  uint   `json:"doc_id"`
	Title  string `json:"title"`
	Type   string `json:"type"` // pdf|docx|xlsx
	Status string `json:"status"` // pending|running|done|error
	Result string `json:"result,omitempty"`
	Err    string `json:"error,omitempty"`

	// queue-internal, not serialized
	content string
	fileExt string
}

// ExportWork input for converter; OutDir is per-job output dir
type ExportWork struct {
	DocID   uint
	Title   string
	Content string
	FileExt string
	Type    string
	OutDir  string
}

// Converter turns a doc into target file, returns absolute output path.
// extracted for tests to inject fake converter; no real soffice needed.
type Converter func(ctx context.Context, w ExportWork) (string, error)

// ExportQueue worker pool: buffered job channel + N workers,
// at most N concurrent exports (soffice is heavy, must throttle).
type ExportQueue struct {
	mu      sync.RWMutex
	jobs    map[string]*ExportJob
	ch      chan *ExportJob
	workers int
	outRoot string
	timeout time.Duration

	convert Converter

	// runtime metrics; tests assert concurrency cap
	active    atomic.Int64
	maxActive atomic.Int64
}

// default params
const (
	defaultWorkers = 2
	defaultTimeout = 120 * time.Second
	queueBuf       = 256
)

// NewExportQueue builds queue; workers<=0 falls back to 2
func NewExportQueue(workers int, outRoot string) *ExportQueue {
	if workers <= 0 {
		workers = defaultWorkers
	}
	_ = os.MkdirAll(outRoot, 0o755)
	q := &ExportQueue{
		jobs:    make(map[string]*ExportJob),
		ch:      make(chan *ExportJob, queueBuf),
		workers: workers,
		outRoot: outRoot,
		timeout: defaultTimeout,
	}
	q.convert = q.sofficeConvert
	return q
}

// NewExportQueueFromEnv reads PICO_EXPORT_WORKERS, default 2
func NewExportQueueFromEnv(outRoot string) *ExportQueue {
	n := defaultWorkers
	if v := os.Getenv("PICO_EXPORT_WORKERS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}
	return NewExportQueue(n, outRoot)
}

// Start launches N workers; they exit when ctx cancels
func (q *ExportQueue) Start(ctx context.Context) {
	for i := 0; i < q.workers; i++ {
		go q.worker(ctx)
	}
}

// Submit enqueues an export job, returns jobID
func (q *ExportQueue) Submit(docID uint, title, content, fileExt, exportType string) string {
	jobID := fmt.Sprintf("job-%d-%d", time.Now().UnixNano(), docID)
	job := &ExportJob{
		ID:      jobID,
		DocID:   docID,
		Title:   title,
		Type:    exportType,
		Status:  "pending",
		content: content,
		fileExt: fileExt,
	}
	q.mu.Lock()
	q.jobs[jobID] = job
	q.mu.Unlock()
	q.ch <- job
	return jobID
}

// Status returns a snapshot copy; nil if not found
func (q *ExportQueue) Status(jobID string) *ExportJob {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if j, ok := q.jobs[jobID]; ok {
		cp := *j
		return &cp
	}
	return nil
}

// MaxConcurrent peak running jobs so far (for tests)
func (q *ExportQueue) MaxConcurrent() int64 { return q.maxActive.Load() }

// worker pulls jobs until ctx cancels or channel closes
func (q *ExportQueue) worker(ctx context.Context) {
	for job := range q.ch {
		q.runOne(ctx, job)
	}
}

func (q *ExportQueue) runOne(parent context.Context, job *ExportJob) {
	q.setStatus(job, "running", "", "")

	// record in-flight concurrency and peak
	cur := q.active.Add(1)
	for {
		peak := q.maxActive.Load()
		if cur <= peak || q.maxActive.CompareAndSwap(peak, cur) {
			break
		}
	}
	defer q.active.Add(-1)

	outDir := filepath.Join(q.outRoot, job.ID)
	_ = os.MkdirAll(outDir, 0o755)

	// per-job timeout prevents one hung soffice from killing workers
	workCtx, cancel := context.WithTimeout(parent, q.timeout)
	defer cancel()

	out, err := q.convert(workCtx, ExportWork{
		DocID:   job.DocID,
		Title:   job.Title,
		Content: job.content,
		FileExt: job.fileExt,
		Type:    job.Type,
		OutDir:  outDir,
	})
	if err != nil {
		q.setStatus(job, "error", "", err.Error())
		return
	}
	q.setStatus(job, "done", out, "")
}

func (q *ExportQueue) setStatus(job *ExportJob, status, result, errMsg string) {
	q.mu.Lock()
	job.Status = status
	job.Result = result
	job.Err = errMsg
	q.mu.Unlock()
}

// sofficeConvert default converter: soffice --headless to pdf/docx/xlsx.
// logic mirrors handler.ExportDoc (pdf wraps A4 Flat ODT first).
func (q *ExportQueue) sofficeConvert(ctx context.Context, w ExportWork) (string, error) {
	srcPath := filepath.Join(w.OutDir, fmt.Sprintf("doc-%d.%s", w.DocID, w.FileExt))
	convertTo := w.Type
	if w.Type == "pdf" {
		srcPath = filepath.Join(w.OutDir, fmt.Sprintf("doc-%d.fodt", w.DocID))
		if err := os.WriteFile(srcPath, []byte(buildFODT(w.Content)), 0o644); err != nil {
			return "", err
		}
		convertTo = "pdf:writer_pdf_Export"
	} else {
		if err := os.WriteFile(srcPath, []byte(w.Content), 0o644); err != nil {
			return "", err
		}
	}

	cmd := exec.CommandContext(ctx, "soffice", "--headless",
		"--convert-to", convertTo, "--outdir", w.OutDir, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("soffice: %v: %s", err, string(out))
	}
	outPath := filepath.Join(w.OutDir, fmt.Sprintf("doc-%d.%s", w.DocID, w.Type))
	if _, err := os.Stat(outPath); err != nil {
		return "", fmt.Errorf("converted file missing: %w", err)
	}
	return outPath, nil
}

// buildFODT wraps body into A4 Flat ODT (mirrors handler; standalone copy)
func buildFODT(content string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<office:document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0" office:version="1.2" office:mimetype="application/vnd.oasis.opendocument.text">`)
	b.WriteString(`<office:automatic-styles><style:page-layout style:name="pm1"><style:page-layout-properties fo:page-width="21cm" fo:page-height="29.7cm" fo:margin-top="2cm" fo:margin-bottom="2cm" fo:margin-left="2cm" fo:margin-right="2cm"/></style:page-layout></office:automatic-styles>`)
	b.WriteString(`<office:master-styles><style:master-page style:name="Standard" style:page-layout-name="pm1"/></office:master-styles>`)
	b.WriteString(`<office:body><office:text>`)
	for _, line := range strings.Split(content, "\n") {
		trim := strings.TrimSpace(strings.TrimRight(line, "\r"))
		if trim == "" {
			continue
		}
		level, body := 0, trim
		switch {
		case strings.HasPrefix(trim, "### "):
			level, body = 3, trim[4:]
		case strings.HasPrefix(trim, "## "):
			level, body = 2, trim[3:]
		case strings.HasPrefix(trim, "# "):
			level, body = 1, trim[2:]
		}
		esc := xmlEscape(body)
		if level > 0 {
			fmt.Fprintf(&b, `<text:h text:outline-level="%d">%s</text:h>`, level, esc)
		} else {
			fmt.Fprintf(&b, `<text:p>%s</text:p>`, esc)
		}
	}
	b.WriteString(`</office:text></office:body></office:document>`)
	return b.String()
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

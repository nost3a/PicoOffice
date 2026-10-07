package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// concurrency: 5 jobs, 2 workers; at most 2 running, rest pending.
func TestExportQueueConcurrencyLimit(t *testing.T) {
	q := NewExportQueue(2, t.TempDir())

	release := make(chan struct{})
	started := make(chan struct{}, 5)
	q.convert = func(ctx context.Context, w ExportWork) (string, error) {
		started <- struct{}{}
		<-release // block until test releases
		return filepath.Join(w.OutDir, "out.pdf"), nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q.Start(ctx)

	ids := make([]string, 5)
	for i := 0; i < 5; i++ {
		ids[i] = q.Submit(uint(i+1), "doc", "content", "txt", "pdf")
	}

	// only 2 workers, exactly 2 jobs started
	for i := 0; i < 2; i++ {
		<-started
	}
	time.Sleep(100 * time.Millisecond)

	if got := q.MaxConcurrent(); got != 2 {
		t.Fatalf("max concurrent = %d, want 2 (worker 上限)", got)
	}
	// the last 3 must still be queued
	for i := 2; i < 5; i++ {
		if st := q.Status(ids[i]); st == nil || st.Status != "pending" {
			t.Fatalf("job[%d] status = %v, want pending", i, st)
		}
	}

	close(release)
	// wait for all to finish
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		done := true
		for _, id := range ids {
			if q.Status(id).Status != "done" {
				done = false
				break
			}
		}
		if done {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	for _, id := range ids {
		if st := q.Status(id); st.Status != "done" {
			t.Fatalf("job %s status=%s, want done", id, st.Status)
		}
	}
}

// per-job timeout: if converter blocks, job becomes error after timeout.
func TestExportQueueTimeout(t *testing.T) {
	q := NewExportQueue(1, t.TempDir())
	q.timeout = 80 * time.Millisecond
	q.convert = func(ctx context.Context, w ExportWork) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q.Start(ctx)

	id := q.Submit(1, "doc", "content", "txt", "pdf")
	deadline := time.Now().Add(3 * time.Second)
	var st *ExportJob
	for time.Now().Before(deadline) {
		st = q.Status(id)
		if st.Status == "error" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if st == nil || st.Status != "error" {
		t.Fatalf("status=%v, want error after timeout", st)
	}
}

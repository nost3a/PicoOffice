package service

import (
	"encoding/json"
	"errors"
	"sync"

	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/model"
)

// ErrClientAhead client version ahead of server, abnormal
var ErrClientAhead = errors.New("client version ahead of server")

// one lock per doc, serializes step commits to avoid version race
var (
	docLocksMu sync.Mutex
	docLocks   = map[uint]*sync.Mutex{}
)

func docLock(docID uint) *sync.Mutex {
	docLocksMu.Lock()
	defer docLocksMu.Unlock()
	m, ok := docLocks[docID]
	if !ok {
		m = &sync.Mutex{}
		docLocks[docID] = m
	}
	return m
}

// LatestVersion current version; 0 when no steps
func LatestVersion(db *gorm.DB, docID uint) int {
	var v int
	db.Model(&model.CollabStep{}).
		Where("doc_id = ?", docID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&v)
	return v
}

// ReceiveResult result of ReceiveSteps
type ReceiveResult struct {
	Accepted bool            `json:"accepted"`
	Version  int             `json:"version"`           // server's latest version
	Steps    []model.CollabStep `json:"steps,omitempty"` // on reject: existing steps in (clientVersion, latest] for client transform
}

// ReceiveSteps ingests a batch of client steps.
//   - clientVersion == latest: persist in order, bump version, return {accepted:true, version:<new>
//   - clientVersion < latest: no-op, return existing steps in range for client transform
//   - clientVersion > latest: treat as error
func ReceiveSteps(db *gorm.DB, docID uint, clientVersion int, steps []json.RawMessage, uid uint) (*ReceiveResult, error) {
	mu := docLock(docID)
	mu.Lock()
	defer mu.Unlock()

	latest := LatestVersion(db, docID)
	switch {
	case clientVersion < latest:
		var missing []model.CollabStep
		db.Where("doc_id = ? AND version > ?", docID, clientVersion).
			Order("version ASC").Find(&missing)
		return &ReceiveResult{Accepted: false, Version: latest, Steps: missing}, nil
	case clientVersion > latest:
		return nil, ErrClientAhead
	}

	// clientVersion == latest, persist in order
	tx := db.Begin()
	v := latest
	for _, s := range steps {
		v++
		row := model.CollabStep{
			DocID:    docID,
			Version:  v,
			StepData: string(s),
			UserID:   uid,
		}
		if err := tx.Create(&row).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &ReceiveResult{Accepted: true, Version: v}, nil
}

// StepsAfter returns steps after a version (for replay on load)
func StepsAfter(db *gorm.DB, docID uint, afterVersion int) (int, []model.CollabStep) {
	latest := LatestVersion(db, docID)
	var list []model.CollabStep
	db.Where("doc_id = ? AND version > ?", docID, afterVersion).
		Order("version ASC").Find(&list)
	return latest, list
}

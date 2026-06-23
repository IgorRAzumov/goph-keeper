package fake

import (
	"context"
	"sort"
	"sync"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/record/model"
)

// RecordRepo — in-memory RecordStore для тестов.
type RecordRepo struct {
	mu      sync.Mutex
	records map[string]map[string]*model.Record // owner -> id -> record
}

// NewRecordRepo создаёт пустой in-memory репозиторий записей.
func NewRecordRepo() *RecordRepo {
	return &RecordRepo{records: map[string]map[string]*model.Record{}}
}

func (r *RecordRepo) Update(_ context.Context, ownerID string, rec *model.Record) error {
	conflicts, err := r.UpdateBatch(context.Background(), ownerID, []*model.Record{rec})
	if err != nil {
		return err
	}
	if len(conflicts) > 0 {
		return common.ErrConflict
	}
	return nil
}

func (r *RecordRepo) UpdateBatch(_ context.Context, ownerID string, records []*model.Record) ([]*model.Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ownerID == "" {
		return nil, common.ErrInvalidInput
	}
	if r.records[ownerID] == nil {
		r.records[ownerID] = map[string]*model.Record{}
	}
	var conflicts []*model.Record
	for _, rec := range records {
		if rec == nil || rec.ID == "" {
			return nil, common.ErrInvalidInput
		}
		stored, ok := r.records[ownerID][rec.ID]
		if !ok {
			copyRec := *rec
			copyRec.OwnerID = ownerID
			r.records[ownerID][rec.ID] = &copyRec
			continue
		}
		if rec.Version <= stored.Version {
			copyStored := *stored
			conflicts = append(conflicts, &copyStored)
			continue
		}
		copyRec := *rec
		copyRec.OwnerID = ownerID
		r.records[ownerID][rec.ID] = &copyRec
	}
	return conflicts, nil
}

func (r *RecordRepo) Get(_ context.Context, ownerID, recordID string) (*model.Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ownerRecords, ok := r.records[ownerID]
	if !ok {
		return nil, common.ErrNotFound
	}
	rec, ok := ownerRecords[recordID]
	if !ok {
		return nil, common.ErrNotFound
	}
	copyRec := *rec
	return &copyRec, nil
}

func (r *RecordRepo) ListSince(_ context.Context, ownerID string, sinceVersion int64) ([]*model.Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ownerRecords := r.records[ownerID]
	out := make([]*model.Record, 0)
	for _, rec := range ownerRecords {
		if rec.Version > sinceVersion {
			copyRec := *rec
			out = append(out, &copyRec)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

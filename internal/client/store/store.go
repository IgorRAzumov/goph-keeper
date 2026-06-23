package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"sort"

	"goph-keeper/internal/client/model"
)

// Record — локальная запись (ciphertext как на сервере).
type Record struct {
	ID         string           `json:"id"`
	Type       model.RecordType `json:"type"`
	Meta       string           `json:"meta"`
	Ciphertext []byte           `json:"ciphertext"`
	Version    int64            `json:"version"`
	Deleted    bool             `json:"deleted"`
	Dirty      bool             `json:"dirty"`
}

// Data — локальное хранилище.
type Data struct {
	LastSyncVersion int64             `json:"last_sync_version"`
	Records         map[string]Record `json:"records"`
}

// Open загружает data.json или создаёт пустое хранилище.
func Open(path string) (*Data, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Data{Records: map[string]Record{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var data Data
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("store: parse: %w", err)
	}
	if data.Records == nil {
		data.Records = map[string]Record{}
	}
	return &data, nil
}

// Save сохраняет data.json.
func (data *Data) Save(path string) error {
	if data == nil {
		return errors.New("store: nil data")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

// MaxVersion возвращает максимальный version среди всех записей.
func (data *Data) MaxVersion() int64 {
	var maxInt64 int64
	for _, record := range data.Records {
		if record.Version > maxInt64 {
			maxInt64 = record.Version
		}
	}
	return maxInt64
}

// NextVersion возвращает version для новой/обновлённой записи.
func (data *Data) NextVersion() int64 {
	return data.MaxVersion() + 1
}

// ListActive перебирает неудалённые записи в порядке возрастания meta.
func (data *Data) ListActive() iter.Seq[Record] {
	// Сортировка требует материализации: пройти map в нужном порядке иначе нельзя.
	ordered := make([]Record, 0, len(data.Records))
	for _, record := range data.Records {
		if !record.Deleted {
			ordered = append(ordered, record)
		}
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Meta < ordered[j].Meta
	})
	return func(yield func(Record) bool) {
		for _, record := range ordered {
			if !yield(record) {
				return
			}
		}
	}
}

// Get возвращает запись по id.
func (data *Data) Get(id string) (Record, error) {
	record, ok := data.Records[id]
	if !ok || record.Deleted {
		return Record{}, model.ErrNotFound
	}
	return record, nil
}

// Put сохраняет запись и помечает dirty.
func (data *Data) Put(record Record) {
	record.Dirty = true
	data.Records[record.ID] = record
}

// Delete помечает tombstone.
func (data *Data) Delete(id string) error {
	record, ok := data.Records[id]
	if !ok || record.Deleted {
		return model.ErrNotFound
	}
	record.Deleted = true
	record.Version = data.NextVersion()
	record.Dirty = true
	data.Records[id] = record
	return nil
}

// DirtyRecords перебирает записи, требующие push. Обход ленивый и без аллокаций:
// потребитель может остановиться через break, не собирая весь слайс.
func (data *Data) DirtyRecords() iter.Seq[Record] {
	return func(yield func(Record) bool) {
		for _, record := range data.Records {
			if record.Dirty {
				if !yield(record) {
					return
				}
			}
		}
	}
}

// MergeRemote применяет запись с сервера (LWW по version).
func (data *Data) MergeRemote(record Record) {
	local, ok := data.Records[record.ID]
	if ok && local.Version > record.Version {
		return
	}
	record.Dirty = false
	data.Records[record.ID] = record
	if record.Version > data.LastSyncVersion {
		data.LastSyncVersion = record.Version
	}
}

// MarkPushed снимает dirty после успешного push.
func (data *Data) MarkPushed(ids ...string) {
	for _, id := range ids {
		record, ok := data.Records[id]
		if !ok {
			continue
		}
		record.Dirty = false
		data.Records[id] = record
		if record.Version > data.LastSyncVersion {
			data.LastSyncVersion = record.Version
		}
	}
}

// RebaseDirtyVersions переназначает version у dirty-записей на значения строго выше
// текущего максимума. Это помогает выйти из конфликтов синхронизации, когда сервер
// уже принял более высокие версии от другого клиента.
func (data *Data) RebaseDirtyVersions() []string {
	if data == nil {
		return nil
	}
	next := data.MaxVersion() + 1
	rebased := make([]string, 0)
	for id, record := range data.Records {
		if !record.Dirty {
			continue
		}
		if record.Version >= next {
			next = record.Version + 1
			continue
		}
		record.Version = next
		record.Dirty = true
		data.Records[id] = record
		rebased = append(rebased, id)
		next++
	}
	return rebased
}

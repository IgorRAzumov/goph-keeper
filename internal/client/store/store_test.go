package store_test

import (
	"path/filepath"
	"testing"

	"goph-keeper/internal/client/model"
	"goph-keeper/internal/client/store"
)

func TestStorePutListDelete(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "data.json")
	data, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	data.Put(store.Record{
		ID: "r1", Type: model.RecordTypeText, Meta: "note",
		Ciphertext: []byte("x"), Version: 1,
	})
	if err := data.Save(path); err != nil {
		t.Fatal(err)
	}
	reloaded, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.ListActive()) != 1 {
		t.Fatalf("expected 1 record")
	}
	if err := reloaded.Delete("r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := reloaded.Get("r1"); err != model.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestMergeRemoteKeepsNewerVersion(t *testing.T) {
	t.Parallel()

	data := &store.Data{Records: map[string]store.Record{}}
	data.Put(store.Record{ID: "r1", Version: 5, Ciphertext: []byte("local")})
	data.MergeRemote(store.Record{ID: "r1", Version: 3, Ciphertext: []byte("old")})
	if string(data.Records["r1"].Ciphertext) != "local" {
		t.Fatal("expected local to win")
	}
	data.MergeRemote(store.Record{ID: "r1", Version: 10, Ciphertext: []byte("remote")})
	if string(data.Records["r1"].Ciphertext) != "remote" {
		t.Fatal("expected remote to win")
	}
}

func TestDirtyRecordsAndMarkPushed(t *testing.T) {
	t.Parallel()

	data := &store.Data{Records: map[string]store.Record{}}
	data.Put(store.Record{ID: "r1", Version: 1, Ciphertext: []byte("x")})
	dirty := data.DirtyRecords()
	if len(dirty) != 1 {
		t.Fatalf("expected 1 dirty, got %d", len(dirty))
	}
	data.MarkPushed("r1")
	if len(data.DirtyRecords()) != 0 {
		t.Fatal("expected no dirty after push")
	}
	if data.LastSyncVersion != 1 {
		t.Fatalf("expected last sync version 1, got %d", data.LastSyncVersion)
	}
}

func TestNextVersion(t *testing.T) {
	t.Parallel()

	data := &store.Data{Records: map[string]store.Record{
		"a": {Version: 3},
		"b": {Version: 7},
	}}
	if data.NextVersion() != 8 {
		t.Fatalf("expected 8, got %d", data.NextVersion())
	}
}

func TestRebaseDirtyVersions(t *testing.T) {
	t.Parallel()

	data := &store.Data{Records: map[string]store.Record{
		"clean": {ID: "clean", Version: 10, Dirty: false},
		"d1":    {ID: "d1", Version: 2, Dirty: true},
		"d2":    {ID: "d2", Version: 10, Dirty: true},
	}}

	rebased := data.RebaseDirtyVersions()
	if len(rebased) != 2 {
		t.Fatalf("expected two rebased records, got %d", len(rebased))
	}
	if data.Records["d1"].Version <= 10 {
		t.Fatalf("expected rebased version > 10, got %d", data.Records["d1"].Version)
	}
	if data.Records["d2"].Version <= 10 {
		t.Fatalf("expected rebased version > 10, got %d", data.Records["d2"].Version)
	}
}

package persistence

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dump.json")
	original := Snapshot{
		Values: map[string]string{
			"name": "chanyut",
		},
		Expirations: map[string]time.Time{
			"name": time.Now().Add(time.Minute).UTC(),
		},
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := loaded.Values["name"]; got != "chanyut" {
		t.Fatalf("Load() value = %q, want %q", got, "chanyut")
	}

	if _, ok := loaded.Expirations["name"]; !ok {
		t.Fatal("Load() expiration missing, want present")
	}
}

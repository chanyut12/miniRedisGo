package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Snapshot is the JSON-serializable form of the store state.
type Snapshot struct {
	Values      map[string]string    `json:"values"`
	Expirations map[string]time.Time `json:"expirations"`
}

// Save writes the snapshot to disk as JSON.
func Save(path string, snapshot Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0o644)
}

// Load reads a snapshot JSON file from disk.
func Load(path string) (Snapshot, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, err
	}

	snapshot := Snapshot{
		Values:      make(map[string]string),
		Expirations: make(map[string]time.Time),
	}
	if len(payload) == 0 {
		return snapshot, nil
	}

	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return Snapshot{}, err
	}

	if snapshot.Values == nil {
		snapshot.Values = make(map[string]string)
	}
	if snapshot.Expirations == nil {
		snapshot.Expirations = make(map[string]time.Time)
	}

	return snapshot, nil
}

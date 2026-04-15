package store

import (
	"sync"
	"time"

	"github.com/chanyut12/miniRedisGo/internal/persistence"
)

// Store is the concrete in-memory storage implementation used by the server.
type Store struct {
	mu          sync.RWMutex
	values      map[string]string
	expirations map[string]time.Time
}

// New constructs the concrete store implementation.
func New() *Store {
	return &Store{
		values:      make(map[string]string),
		expirations: make(map[string]time.Time),
	}
}

// Set stores a value for the key and clears any existing expiration.
func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[key] = value
	delete(s.expirations, key)
}

// Get returns the current value for a key when it exists.
func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.deleteExpiredLocked(key) {
		return "", false
	}

	value, ok := s.values[key]
	return value, ok
}

// Del removes a key and reports whether anything was deleted.
func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.deleteExpiredLocked(key) {
		return false
	}

	if _, ok := s.values[key]; !ok {
		return false
	}

	delete(s.values, key)
	delete(s.expirations, key)
	return true
}

// Exists reports whether a key is present in the store.
func (s *Store) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.deleteExpiredLocked(key) {
		return false
	}

	_, ok := s.values[key]
	return ok
}

// Expire stores a TTL for the key.
func (s *Store) Expire(key string, seconds int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.deleteExpiredLocked(key) {
		return false
	}

	if _, ok := s.values[key]; !ok {
		return false
	}

	s.expirations[key] = time.Now().Add(time.Duration(seconds) * time.Second)
	return true
}

// TTL reports the remaining TTL or Redis-like sentinel values.
func (s *Store) TTL(key string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.deleteExpiredLocked(key) {
		return -2
	}

	if _, ok := s.values[key]; !ok {
		return -2
	}

	expiresAt, ok := s.expirations[key]
	if !ok {
		return -1
	}

	return int64(time.Until(expiresAt).Seconds())
}

// DeleteExpired removes all expired keys and returns the number deleted.
func (s *Store) DeleteExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	deleted := 0
	now := time.Now()
	for key, expiresAt := range s.expirations {
		if now.Before(expiresAt) {
			continue
		}

		delete(s.values, key)
		delete(s.expirations, key)
		deleted++
	}

	return deleted
}

// SnapshotData returns a persistence-ready copy of the non-expired store state.
func (s *Store) SnapshotData() persistence.Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	snapshot := persistence.Snapshot{
		Values:      make(map[string]string),
		Expirations: make(map[string]time.Time),
	}

	for key, value := range s.values {
		expiresAt, hasTTL := s.expirations[key]
		if hasTTL && !now.Before(expiresAt) {
			delete(s.expirations, key)
			delete(s.values, key)
			continue
		}

		snapshot.Values[key] = value
		if hasTTL {
			snapshot.Expirations[key] = expiresAt
		}
	}

	return snapshot
}

// LoadSnapshot replaces the current store state with a snapshot, skipping
// entries that are already expired at load time.
func (s *Store) LoadSnapshot(snapshot persistence.Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values = make(map[string]string)
	s.expirations = make(map[string]time.Time)

	now := time.Now()
	for key, value := range snapshot.Values {
		expiresAt, hasTTL := snapshot.Expirations[key]
		if hasTTL && !now.Before(expiresAt) {
			continue
		}

		s.values[key] = value
		if hasTTL {
			s.expirations[key] = expiresAt
		}
	}
}

func (s *Store) deleteExpiredLocked(key string) bool {
	expiresAt, ok := s.expirations[key]
	if !ok {
		return false
	}

	if time.Now().Before(expiresAt) {
		return false
	}

	delete(s.values, key)
	delete(s.expirations, key)
	return true
}

func (s *Store) SetEx(key, value string, seconds int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
	s.expirations[key] = time.Now().Add(time.Duration(seconds) * time.Second)
}

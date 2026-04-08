package store

import (
	"testing"
	"time"

	"github.com/chanyut12/miniRedisGo/internal/persistence"
)

func TestStoreSetAndGet(t *testing.T) {
	store := New()
	store.Set("name", "chanyut")

	value, ok := store.Get("name")
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}

	if value != "chanyut" {
		t.Fatalf("Get() value = %q, want %q", value, "chanyut")
	}
}

func TestStoreGetMissing(t *testing.T) {
	store := New()

	if _, ok := store.Get("missing"); ok {
		t.Fatal("Get() ok = true, want false")
	}
}

func TestStoreDel(t *testing.T) {
	store := New()
	store.Set("name", "chanyut")

	if deleted := store.Del("name"); !deleted {
		t.Fatal("Del() = false, want true")
	}

	if deleted := store.Del("name"); deleted {
		t.Fatal("Del() second call = true, want false")
	}
}

func TestStoreExists(t *testing.T) {
	store := New()
	store.Set("name", "chanyut")

	if !store.Exists("name") {
		t.Fatal("Exists() = false, want true")
	}

	if store.Exists("missing") {
		t.Fatal("Exists() missing = true, want false")
	}
}

func TestStoreExpireAndTTL(t *testing.T) {
	store := New()
	store.Set("token", "abc123")

	if ok := store.Expire("token", 2); !ok {
		t.Fatal("Expire() = false, want true")
	}

	ttl := store.TTL("token")
	if ttl < 0 || ttl > 2 {
		t.Fatalf("TTL() = %d, want value between 0 and 2", ttl)
	}
}

func TestStoreLazyExpiration(t *testing.T) {
	store := New()
	store.Set("token", "abc123")

	if ok := store.Expire("token", 1); !ok {
		t.Fatal("Expire() = false, want true")
	}

	time.Sleep(1100 * time.Millisecond)

	if _, ok := store.Get("token"); ok {
		t.Fatal("Get() after expiration ok = true, want false")
	}

	if store.Exists("token") {
		t.Fatal("Exists() after expiration = true, want false")
	}

	if ttl := store.TTL("token"); ttl != -2 {
		t.Fatalf("TTL() after expiration = %d, want -2", ttl)
	}
}

func TestStoreSetClearsPreviousTTL(t *testing.T) {
	store := New()
	store.Set("token", "abc123")

	if ok := store.Expire("token", 5); !ok {
		t.Fatal("Expire() = false, want true")
	}

	store.Set("token", "new-value")

	if ttl := store.TTL("token"); ttl != -1 {
		t.Fatalf("TTL() after SET reset = %d, want -1", ttl)
	}
}

func TestStoreDeleteExpired(t *testing.T) {
	store := New()
	store.Set("expired", "yes")
	store.Set("alive", "yes")

	if ok := store.Expire("expired", 1); !ok {
		t.Fatal("Expire(expired) = false, want true")
	}

	if ok := store.Expire("alive", 5); !ok {
		t.Fatal("Expire(alive) = false, want true")
	}

	time.Sleep(1100 * time.Millisecond)

	if deleted := store.DeleteExpired(); deleted != 1 {
		t.Fatalf("DeleteExpired() = %d, want 1", deleted)
	}

	if store.Exists("expired") {
		t.Fatal("Exists(expired) = true, want false")
	}

	if !store.Exists("alive") {
		t.Fatal("Exists(alive) = false, want true")
	}
}

func TestStoreSnapshotDataSkipsExpired(t *testing.T) {
	store := New()
	store.Set("alive", "yes")
	store.Set("expired", "gone")

	if ok := store.Expire("expired", 1); !ok {
		t.Fatal("Expire(expired) = false, want true")
	}

	time.Sleep(1100 * time.Millisecond)

	snapshot := store.SnapshotData()
	if _, ok := snapshot.Values["expired"]; ok {
		t.Fatal("SnapshotData() includes expired key, want skipped")
	}

	if got := snapshot.Values["alive"]; got != "yes" {
		t.Fatalf("SnapshotData() alive value = %q, want %q", got, "yes")
	}
}

func TestStoreLoadSnapshotSkipsExpired(t *testing.T) {
	store := New()
	snapshot := persistence.Snapshot{
		Values: map[string]string{
			"alive":   "yes",
			"expired": "gone",
		},
		Expirations: map[string]time.Time{
			"expired": time.Now().Add(-time.Minute),
		},
	}

	store.LoadSnapshot(snapshot)

	if !store.Exists("alive") {
		t.Fatal("Exists(alive) = false, want true")
	}

	if store.Exists("expired") {
		t.Fatal("Exists(expired) = true, want false")
	}
}

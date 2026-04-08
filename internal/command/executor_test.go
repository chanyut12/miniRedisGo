package command

import (
	"strconv"
	"testing"
	"time"

	"github.com/chanyut12/miniRedisGo/internal/protocol"
	"github.com/chanyut12/miniRedisGo/internal/store"
)

func TestBasicExecutorExecutePing(t *testing.T) {
	executor := NewBasicExecutor(nil)

	got := executor.Execute(protocol.Command{Name: "PING"})
	if got != "PONG" {
		t.Fatalf("Execute() = %q, want %q", got, "PONG")
	}
}

func TestBasicExecutorExecutePingWrongArity(t *testing.T) {
	executor := NewBasicExecutor(nil)

	got := executor.Execute(protocol.Command{Name: "PING", Args: []string{"extra"}})
	want := "ERR wrong number of arguments for 'PING' command"
	if got != want {
		t.Fatalf("Execute() = %q, want %q", got, want)
	}
}

func TestBasicExecutorExecuteSetGetDelExists(t *testing.T) {
	executor := NewBasicExecutor(store.New())

	if got := executor.Execute(protocol.Command{Name: "SET", Args: []string{"name", "chanyut"}}); got != "OK" {
		t.Fatalf("SET Execute() = %q, want %q", got, "OK")
	}

	if got := executor.Execute(protocol.Command{Name: "GET", Args: []string{"name"}}); got != "chanyut" {
		t.Fatalf("GET Execute() = %q, want %q", got, "chanyut")
	}

	if got := executor.Execute(protocol.Command{Name: "EXISTS", Args: []string{"name"}}); got != "1" {
		t.Fatalf("EXISTS Execute() = %q, want %q", got, "1")
	}

	if got := executor.Execute(protocol.Command{Name: "DEL", Args: []string{"name"}}); got != "1" {
		t.Fatalf("DEL Execute() = %q, want %q", got, "1")
	}

	if got := executor.Execute(protocol.Command{Name: "GET", Args: []string{"name"}}); got != "(nil)" {
		t.Fatalf("GET after DEL Execute() = %q, want %q", got, "(nil)")
	}
}

func TestBasicExecutorExecuteExpireAndTTL(t *testing.T) {
	executor := NewBasicExecutor(store.New())

	if got := executor.Execute(protocol.Command{Name: "SET", Args: []string{"token", "abc123"}}); got != "OK" {
		t.Fatalf("SET Execute() = %q, want %q", got, "OK")
	}

	if got := executor.Execute(protocol.Command{Name: "EXPIRE", Args: []string{"token", "1"}}); got != "1" {
		t.Fatalf("EXPIRE Execute() = %q, want %q", got, "1")
	}

	ttlText := executor.Execute(protocol.Command{Name: "TTL", Args: []string{"token"}})
	ttl, err := strconv.ParseInt(ttlText, 10, 64)
	if err != nil {
		t.Fatalf("TTL Execute() parse error = %v", err)
	}

	if ttl < 0 || ttl > 1 {
		t.Fatalf("TTL Execute() = %d, want value between 0 and 1", ttl)
	}

	time.Sleep(1100 * time.Millisecond)

	if got := executor.Execute(protocol.Command{Name: "GET", Args: []string{"token"}}); got != "(nil)" {
		t.Fatalf("GET after expiration Execute() = %q, want %q", got, "(nil)")
	}
}

func TestBasicExecutorExecuteInvalidExpireTime(t *testing.T) {
	executor := NewBasicExecutor(store.New())

	got := executor.Execute(protocol.Command{Name: "EXPIRE", Args: []string{"token", "0"}})
	want := "ERR invalid expire time"
	if got != want {
		t.Fatalf("Execute() = %q, want %q", got, want)
	}
}

func TestBasicExecutorExecuteGetMissing(t *testing.T) {
	executor := NewBasicExecutor(store.New())

	got := executor.Execute(protocol.Command{Name: "GET", Args: []string{"missing"}})
	if got != "(nil)" {
		t.Fatalf("Execute() = %q, want %q", got, "(nil)")
	}
}

func TestBasicExecutorExecuteUnknownCommand(t *testing.T) {
	executor := NewBasicExecutor(nil)

	got := executor.Execute(protocol.Command{Name: "UNKNOWN"})
	want := "ERR unknown command 'UNKNOWN'"
	if got != want {
		t.Fatalf("Execute() = %q, want %q", got, want)
	}
}

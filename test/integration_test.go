package test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appserver "github.com/chanyut12/miniRedisGo/internal/server"
)

func TestServerPing(t *testing.T) {
	addr, shutdown := startTestServer(t)
	defer shutdown()

	responses := runCommands(t, addr, "PING")
	if responses[0] != "PONG" {
		t.Fatalf("PING response = %q, want %q", responses[0], "PONG")
	}
}

func TestServerSetGetDelExists(t *testing.T) {
	addr, shutdown := startTestServer(t)
	defer shutdown()

	responses := runCommands(t, addr,
		"SET name chanyut",
		"GET name",
		"EXISTS name",
		"DEL name",
		"GET name",
	)

	want := []string{"OK", "chanyut", "1", "1", "(nil)"}
	for i := range want {
		if responses[i] != want[i] {
			t.Fatalf("response[%d] = %q, want %q", i, responses[i], want[i])
		}
	}
}

func TestServerExpireTTL(t *testing.T) {
	addr, shutdown := startTestServer(t)
	defer shutdown()

	conn := dialServer(t, addr)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	writeCommand(t, conn, "SET token abc123")
	assertResponse(t, reader, "OK")

	writeCommand(t, conn, "EXPIRE token 1")
	assertResponse(t, reader, "1")

	writeCommand(t, conn, "TTL token")
	ttlText := readResponse(t, reader)
	if ttlText != "0" && ttlText != "1" {
		t.Fatalf("TTL immediate response = %q, want %q or %q", ttlText, "0", "1")
	}

	time.Sleep(1100 * time.Millisecond)

	writeCommand(t, conn, "GET token")
	assertResponse(t, reader, "(nil)")

	writeCommand(t, conn, "TTL token")
	assertResponse(t, reader, "-2")
}

func TestServerInvalidCommand(t *testing.T) {
	addr, shutdown := startTestServer(t)
	defer shutdown()

	responses := runCommands(t, addr, "NOPE")
	want := "ERR unknown command 'NOPE'"
	if responses[0] != want {
		t.Fatalf("invalid command response = %q, want %q", responses[0], want)
	}
}

func startTestServer(t *testing.T) (string, func()) {
	t.Helper()

	port := reservePort(t)
	cfg := appserver.DefaultConfig()
	cfg.Host = "127.0.0.1"
	cfg.Port = port
	cfg.Debug = false
	cfg.CleanupInterval = 100 * time.Millisecond
	cfg.SnapshotPath = filepath.Join(t.TempDir(), "dump.json")

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := appserver.New(cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Run(ctx)
	}()

	waitForServer(t, cfg.Address())

	return cfg.Address(), func() {
		cancel()
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("server shutdown error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for server shutdown")
		}
	}
}

func reservePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port listen error: %v", err)
	}
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("listener addr type = %T, want *net.TCPAddr", listener.Addr())
	}

	return addr.Port
}

func waitForServer(t *testing.T, addr string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("server at %s did not start in time", addr)
}

func runCommands(t *testing.T, addr string, commands ...string) []string {
	t.Helper()

	conn := dialServer(t, addr)
	defer conn.Close()

	reader := bufio.NewReader(conn)
	responses := make([]string, 0, len(commands))
	for _, command := range commands {
		writeCommand(t, conn, command)
		responses = append(responses, readResponse(t, reader))
	}

	return responses
}

func dialServer(t *testing.T, addr string) net.Conn {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}

	return conn
}

func writeCommand(t *testing.T, conn net.Conn, command string) {
	t.Helper()

	if _, err := fmt.Fprintf(conn, "%s\n", command); err != nil {
		t.Fatalf("write command error: %v", err)
	}
}

func readResponse(t *testing.T, reader *bufio.Reader) string {
	t.Helper()

	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read response error: %v", err)
	}

	return strings.TrimSpace(line)
}

func assertResponse(t *testing.T, reader *bufio.Reader, want string) {
	t.Helper()

	if got := readResponse(t, reader); got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
}

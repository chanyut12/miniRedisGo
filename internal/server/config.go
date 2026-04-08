package server

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHost            = "localhost"
	defaultPort            = 6379
	defaultSnapshotPath    = "data/dump.json"
	defaultCleanupInterval = time.Second
)

// Config holds the runtime settings for the server process.
type Config struct {
	Host            string
	Port            int
	SnapshotPath    string
	CleanupInterval time.Duration
	Debug           bool
}

// DefaultConfig returns the baseline configuration used by the server.
func DefaultConfig() Config {
	return Config{
		Host:            defaultHost,
		Port:            defaultPort,
		SnapshotPath:    defaultSnapshotPath,
		CleanupInterval: defaultCleanupInterval,
		Debug:           debugFromEnv(),
	}
}

// Address formats the bind host and port for listener setup and logging.
func (c Config) Address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// Validate checks the config before the server starts real work.
func (c Config) Validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("host must not be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if strings.TrimSpace(c.SnapshotPath) == "" {
		return fmt.Errorf("snapshot path must not be empty")
	}
	if c.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup interval must be greater than zero")
	}
	return nil
}

func debugFromEnv() bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv("DEBUG")))
	switch raw {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

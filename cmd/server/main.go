package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanyut12/miniRedisGo/internal/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := server.DefaultConfig()

	flag.StringVar(&cfg.Host, "host", cfg.Host, "host to bind the server to")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "port to bind the server to")
	flag.StringVar(&cfg.SnapshotPath, "snapshot-path", cfg.SnapshotPath, "path to the snapshot file")
	flag.DurationVar(&cfg.CleanupInterval, "cleanup-interval", cfg.CleanupInterval, "interval between cleanup runs")
	flag.BoolVar(&cfg.Debug, "debug", cfg.Debug, "enable debug logging")
	flag.Parse()

	logger := newLogger(cfg.Debug)
	logger.Info(
		"loaded configuration",
		"host", cfg.Host,
		"port", cfg.Port,
		"snapshot_path", cfg.SnapshotPath,
		"cleanup_interval", cfg.CleanupInterval.String(),
		"debug", cfg.Debug,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := server.New(cfg, logger)
	return app.Run(ctx)
}

func newLogger(debug bool) *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	if debug {
		level.Set(slog.LevelDebug)
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}

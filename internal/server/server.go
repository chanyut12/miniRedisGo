package server

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chanyut12/miniRedisGo/internal/command"
	"github.com/chanyut12/miniRedisGo/internal/persistence"
	"github.com/chanyut12/miniRedisGo/internal/protocol"
	"github.com/chanyut12/miniRedisGo/internal/store"
)

// Server owns the high-level process lifecycle.
type Server struct {
	cfg          Config
	logger       *slog.Logger
	parser       protocol.Parser
	executor     command.Executor
	commandStore *store.Store
	listener     net.Listener
	connsMu      sync.Mutex
	conns        map[net.Conn]struct{}
	handlersWG   sync.WaitGroup
}

// New builds a server with its runtime dependencies.
func New(cfg Config, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	commandStore := store.New()

	return &Server{
		cfg:          cfg,
		logger:       logger.With("component", "server"),
		parser:       protocol.NewLineParser(),
		executor:     command.NewBasicExecutor(commandStore),
		commandStore: commandStore,
		conns:        make(map[net.Conn]struct{}),
	}
}

// NewWithListener builds a server using an already-created listener. This is
// primarily useful for tests that need to avoid port reservation races.
func NewWithListener(cfg Config, logger *slog.Logger, listener net.Listener) *Server {
	srv := New(cfg, logger)
	srv.listener = listener
	return srv
}

// Run starts the server lifecycle and accepts client connections until the
// context is canceled.
func (s *Server) Run(ctx context.Context) error {
	if err := s.cfg.Validate(); err != nil {
		return err
	}

	if err := s.loadSnapshot(); err != nil {
		return err
	}

	listener := s.listener
	if listener == nil {
		var err error
		listener, err = net.Listen("tcp", s.cfg.Address())
		if err != nil {
			return err
		}
	}
	defer listener.Close()

	s.logger.Info(
		"server listening",
		"addr", s.cfg.Address(),
		"snapshot_path", s.cfg.SnapshotPath,
		"cleanup_interval", s.cfg.CleanupInterval.String(),
		"debug", s.cfg.Debug,
	)

	go s.runCleanupLoop(ctx)

	go func() {
		<-ctx.Done()
		s.logger.Info("server shutdown requested", "reason", ctx.Err())
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			s.logger.Error("failed to close listener", "error", err)
		}
		s.closeActiveConnections()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				s.handlersWG.Wait()
				return s.saveSnapshot()
			}
			return err
		}

		s.trackConn(conn)
		s.handlersWG.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	logger := s.logger.With("remote_addr", remoteAddr)

	logger.Info("client connected")
	defer func() {
		s.untrackConn(conn)
		s.handlersWG.Done()
		logger.Info("client disconnected")
		_ = conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Debug("received raw command", "line", line)

		cmd, err := s.parser.ParseLine(line)
		if err != nil {
			response := protocol.ErrorResponse(err.Error())
			logger.Debug("parser error", "error", err, "response", response)
			if writeErr := writeResponse(conn, response); writeErr != nil {
				logger.Error("failed to write parser error response", "error", writeErr)
				return
			}
			continue
		}

		logger.Debug("parsed command", "name", cmd.Name, "args", strings.Join(cmd.Args, " "))

		response := s.executor.Execute(cmd)
		logger.Debug("command completed", "name", cmd.Name, "response", response)

		if err := writeResponse(conn, response); err != nil {
			logger.Error("failed to write response", "error", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("connection read failed", "error", err)
	}
}

func (s *Server) runCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deleted := s.commandStore.DeleteExpired()
			if deleted > 0 {
				s.logger.Info("background cleanup removed expired keys", "deleted", deleted)
				continue
			}

			s.logger.Debug("background cleanup tick", "deleted", deleted)
		}
	}
}

func (s *Server) loadSnapshot() error {
	s.logger.Info("loading snapshot", "path", s.cfg.SnapshotPath)

	snapshot, err := persistence.Load(s.cfg.SnapshotPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.logger.Info("snapshot file not found, starting with empty state", "path", s.cfg.SnapshotPath)
			return nil
		}
		return err
	}

	s.commandStore.LoadSnapshot(snapshot)
	s.logger.Info("snapshot loaded", "keys", len(snapshot.Values))
	return nil
}

func (s *Server) saveSnapshot() error {
	snapshot := s.commandStore.SnapshotData()
	s.logger.Info("saving snapshot", "path", s.cfg.SnapshotPath, "keys", len(snapshot.Values))

	if err := persistence.Save(s.cfg.SnapshotPath, snapshot); err != nil {
		return err
	}

	s.logger.Info("snapshot saved", "path", s.cfg.SnapshotPath, "keys", len(snapshot.Values))
	return nil
}

func writeResponse(conn net.Conn, response string) error {
	_, err := conn.Write([]byte(response + "\n"))
	return err
}

func (s *Server) trackConn(conn net.Conn) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()
	s.conns[conn] = struct{}{}
}

func (s *Server) untrackConn(conn net.Conn) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()
	delete(s.conns, conn)
}

func (s *Server) closeActiveConnections() {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()

	for conn := range s.conns {
		_ = conn.Close()
	}
}

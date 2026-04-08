package command

import (
	"fmt"
	"strconv"

	"github.com/chanyut12/miniRedisGo/internal/protocol"
)

// Store describes the storage behavior command handlers depend on.
type Store interface {
	Set(key, value string)
	Get(key string) (string, bool)
	Del(key string) bool
	Exists(key string) bool
	Expire(key string, seconds int64) bool
	TTL(key string) int64
}

// Executor runs a parsed command and returns the response string.
type Executor interface {
	Execute(cmd protocol.Command) string
}

// BasicExecutor handles supported commands and formats protocol responses.
type BasicExecutor struct {
	store Store
}

// NewBasicExecutor constructs the default command executor.
func NewBasicExecutor(store Store) *BasicExecutor {
	return &BasicExecutor{store: store}
}

// Execute routes the command to the matching handler.
func (e *BasicExecutor) Execute(cmd protocol.Command) string {
	switch cmd.Name {
	case "PING":
		return e.executePing(cmd)
	case "SET":
		return e.executeSet(cmd)
	case "GET":
		return e.executeGet(cmd)
	case "DEL":
		return e.executeDel(cmd)
	case "EXISTS":
		return e.executeExists(cmd)
	case "EXPIRE":
		return e.executeExpire(cmd)
	case "TTL":
		return e.executeTTL(cmd)
	default:
		return protocol.ErrorResponse(fmt.Sprintf("unknown command '%s'", cmd.Name))
	}
}

func (e *BasicExecutor) executePing(cmd protocol.Command) string {
	if len(cmd.Args) != 0 {
		return protocol.ErrorResponse("wrong number of arguments for 'PING' command")
	}

	return "PONG"
}

func (e *BasicExecutor) executeSet(cmd protocol.Command) string {
	if len(cmd.Args) != 2 {
		return protocol.ErrorResponse("wrong number of arguments for 'SET' command")
	}

	if e.store == nil {
		return protocol.ErrorResponse("store not configured")
	}

	e.store.Set(cmd.Args[0], cmd.Args[1])
	return "OK"
}

func (e *BasicExecutor) executeGet(cmd protocol.Command) string {
	if len(cmd.Args) != 1 {
		return protocol.ErrorResponse("wrong number of arguments for 'GET' command")
	}

	if e.store == nil {
		return protocol.ErrorResponse("store not configured")
	}

	value, ok := e.store.Get(cmd.Args[0])
	if !ok {
		return "(nil)"
	}

	return value
}

func (e *BasicExecutor) executeDel(cmd protocol.Command) string {
	if len(cmd.Args) != 1 {
		return protocol.ErrorResponse("wrong number of arguments for 'DEL' command")
	}

	if e.store == nil {
		return protocol.ErrorResponse("store not configured")
	}

	if e.store.Del(cmd.Args[0]) {
		return "1"
	}

	return "0"
}

func (e *BasicExecutor) executeExists(cmd protocol.Command) string {
	if len(cmd.Args) != 1 {
		return protocol.ErrorResponse("wrong number of arguments for 'EXISTS' command")
	}

	if e.store == nil {
		return protocol.ErrorResponse("store not configured")
	}

	if e.store.Exists(cmd.Args[0]) {
		return "1"
	}

	return "0"
}

func (e *BasicExecutor) executeExpire(cmd protocol.Command) string {
	if len(cmd.Args) != 2 {
		return protocol.ErrorResponse("wrong number of arguments for 'EXPIRE' command")
	}

	if e.store == nil {
		return protocol.ErrorResponse("store not configured")
	}

	seconds, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil || seconds <= 0 {
		return protocol.ErrorResponse("invalid expire time")
	}

	if e.store.Expire(cmd.Args[0], seconds) {
		return "1"
	}

	return "0"
}

func (e *BasicExecutor) executeTTL(cmd protocol.Command) string {
	if len(cmd.Args) != 1 {
		return protocol.ErrorResponse("wrong number of arguments for 'TTL' command")
	}

	if e.store == nil {
		return protocol.ErrorResponse("store not configured")
	}

	return strconv.FormatInt(e.store.TTL(cmd.Args[0]), 10)
}

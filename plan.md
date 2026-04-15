# MiniRedisGo Main Branch Implementation Plan With Head First Design Principles

## Summary

Implement the `main` submission scope on `develop` using Head First Design Patterns principles as coding rules, not as forced pattern ceremony. The goal is a small, debuggable Go system with clear object boundaries, low coupling, and replaceable behaviors: TCP server, line-based protocol, `PING/SET/GET/DEL/EXISTS/EXPIRE/TTL`, `map + sync.RWMutex`, lazy expiration plus background cleanup, and JSON snapshot persistence.

Design defaults locked for `main`:

- Redis-like responses where practical
- Load snapshot on startup, save on graceful shutdown
- Lazy expiration plus periodic cleanup
- Debug logging via `log/slog` with debug mode
- Head First principles applied:
  - encapsulate what varies
  - program to interfaces where substitution matters
  - favor composition over inheritance
  - open for extension, closed for core rewrites
  - depend on abstractions at subsystem boundaries

## Design Rules From Head First

### Core principles to enforce

- Encapsulate what varies:
  - parsing logic stays in `protocol`
  - command execution stays in `command`
  - storage and expiration stay in `store`
  - snapshot I/O stays in `persistence`
- Program to interfaces:
  - command executor depends on a `Store` interface, not a concrete store type
  - server depends on a `Handler` or executor interface where useful
  - persistence integration uses a small saver/loader abstraction only if it improves testability
- Favor composition over inheritance:
  - `Server` owns logger, store, parser, executor, and persistence collaborators
  - avoid embedding big structs just to share behavior
- Open-closed principle:
  - adding a new command should mean adding a handler, not rewriting the whole dispatcher
  - adding a new persistence format later should not force server-layer rewrites
- Dependency inversion:
  - high-level flow in server should orchestrate abstractions
  - low-level file I/O and map internals stay behind package APIs

### Patterns to use lightly and intentionally

- Strategy:
  - parser/executor/store behavior remains swappable in tests
  - cleanup policy and persistence policy can be constructor-injected if needed
- Command:
  - each supported command should map to a small handler path with explicit inputs and outputs
  - no giant `if` block with mixed parsing, business logic, and formatting
- Template Method idea:
  - request lifecycle should be consistent:
    1. read line
    2. parse
    3. validate
    4. execute
    5. format response
    6. log result
- Observer is not needed as a formal pattern in `main`; use logging instead of building event subscribers

## Implementation Changes

### Project structure and responsibilities

- `cmd/server/main.go`
  - bootstraps config, logger, dependencies, signal handling
- `internal/server`
  - listener, connection loop, request lifecycle orchestration
- `internal/protocol`
  - parse raw line into `Command` value object
  - format standard responses and errors
- `internal/command`
  - dispatch parsed commands to handlers
  - convert store return values into wire responses
- `internal/store`
  - own all state, locking, TTL rules, lazy expiration, cleanup
- `internal/persistence`
  - load/save snapshot JSON only

### Public interfaces and types

- `protocol.Command`
  - fields: `Name string`, `Args []string`, `Raw string`
- `protocol.Parser` interface
  - `ParseLine(line string) (Command, error)`
- `command.Executor` interface
  - `Execute(cmd protocol.Command) string`
- `command.Store` interface
  - `Set(key, value string)`
  - `Get(key string) (string, bool)`
  - `Del(key string) bool`
  - `Exists(key string) bool`
  - `Expire(key string, seconds int64) bool`
  - `TTL(key string) int64`
- `store.Store`
  - concrete implementation using map + mutex
- `persistence.Snapshot`
  - contains values and expiration timestamps
- `persistence.Repository` interface is optional; only add it if needed for testing without file I/O

### Request flow

- Server reads one line from a connection
- Parser returns a `Command` value object
- Executor dispatches by command name to one handler function per command
- Handler calls the store abstraction
- Response formatter returns one line string
- Server writes response and logs the outcome

### Command semantics

- `PING` -> `PONG`
- `SET key value` -> `OK`
- `GET key` -> value or `(nil)`
- `DEL key` -> `1` or `0`
- `EXISTS key` -> `1` or `0`
- `EXPIRE key seconds` -> `1` or `0`
- `TTL key` -> remaining seconds, `-1`, or `-2`
- Invalid command or arity -> `ERR <message>`
- `EXPIRE` with non-positive seconds -> `ERR invalid expire time`
- `SET` clears old TTL on the same key
- Values are single-token strings only in `main`

### Store behavior

- Internal state:
  - `map[string]string`
  - `map[string]time.Time`
  - one `sync.RWMutex`
- Expiration rules are fully encapsulated in store methods
- `Get`, `Exists`, `TTL`, and `Expire` must perform lazy expiration checks
- `DeleteExpired() int` is the only method used by the cleanup goroutine
- Snapshot generation excludes expired keys
- Snapshot loading skips expired entries

### Logging and debugger-readiness

- Use `log/slog`
- Debug mode enabled by flag or env
- Required logs:
  - startup config and listening address
  - snapshot load/save counts and errors
  - client connect/disconnect
  - raw command received in debug mode
  - parsed command in debug mode
  - handler result in debug mode
  - store mutations and lazy expiration in debug mode
  - cleanup deletion counts
- Keep functions short enough that stepping through in a debugger is straightforward
- Prefer explicit returns over hidden side effects
- Keep command handlers thin and deterministic

## Branch Strategy

### `main` branch goals

- `main` must always stay in a demo-ready state
- Only merge code into `main` when the full submission flow works end to end
- `main` scope is limited to:
  - TCP server
  - `PING`, `SET`, `GET`, `DEL`, `EXISTS`, `EXPIRE`, `TTL`
  - `map + sync.RWMutex`
  - lazy expiration and background cleanup
  - JSON snapshot load on startup and save on graceful shutdown
  - debug logging and tests required for confidence
- `main` must not contain half-finished advanced features

### `develop` branch goals

- `develop` is the active implementation branch
- Build all `main` work on `develop` first
- After `main` MVP is complete, keep future work on `develop`
- `develop` is where to continue larger upgrades such as:
  - RESP protocol
  - append-only file
  - `INCR`, `DECR`, `SETEX`
  - list and hash data types
  - authentication
  - Docker polish
  - benchmarks

### Merge rule

- Work in `develop`
- Merge `develop` into `main` only after:
  - unit tests pass
  - integration tests pass
  - manual `nc` demo passes
  - restart and snapshot restore pass
  - README matches real behavior

### Suggested commit flow

- `feat: scaffold project structure and bootstrap server`
- `feat: implement TCP listener and connection handler`
- `feat: add parser and PING command`
- `feat: implement SET GET DEL EXISTS commands`
- `feat: add TTL support with lazy expiration`
- `feat: add background cleanup worker`
- `feat: add snapshot persistence on startup and shutdown`
- `test: add unit and integration coverage`
- `docs: update README for main branch delivery`

## Human-Style Build Checklist

### Phase 0: foundation

- [x] Initialize `go.mod`
- [x] Create folder layout
- [x] Add config struct and defaults
- [x] Add logger setup with debug mode
- [x] Add empty server bootstrap path that compiles

### Phase 1: vertical slice 1

- [x] Implement TCP listener
- [x] Implement per-connection goroutine
- [x] Implement line reader/writer
- [x] Implement parser for `PING`
- [x] Implement executor with `PING`
- [x] Verify `nc localhost 6379` then `PING` returns `PONG`
- [x] Add logs for connect, disconnect, request, response

### Phase 2: core storage slice

- [x] Implement concrete store with mutex
- [x] Add `Set`
- [x] Add `Get`
- [x] Add `Del`
- [x] Add `Exists`
- [x] Add unit tests for all four methods
- [x] Wire `SET/GET/DEL/EXISTS` through executor
- [x] Manual test full command roundtrip via TCP

### Phase 3: TTL slice

- [x] Add expiration map
- [x] Add lazy expiration helper inside store
- [x] Implement `Expire`
- [x] Implement `TTL`
- [x] Ensure `SET` clears existing expiration
- [x] Add unit tests for expiry semantics
- [x] Wire `EXPIRE/TTL` through executor
- [x] Manual test with real waiting and verify outputs

### Phase 4: cleanup slice

- [x] Add cleanup ticker goroutine
- [x] Implement `DeleteExpired() int`
- [x] Log cleanup activity
- [x] Verify expired keys disappear even without direct access

### Phase 5: persistence slice

- [x] Define snapshot struct
- [x] Implement JSON `Save`
- [x] Implement JSON `Load`
- [x] Load snapshot at startup if present
- [x] Save snapshot on graceful shutdown
- [x] Add persistence unit tests
- [x] Manual restart test confirms restore works

### Phase 6: polish

- [x] Standardize all error messages
- [x] Add integration tests for command flow
- [x] Run `go test ./...`
- [x] Run `go test -race ./...` if available
- [x] Update README to match actual implementation only
- [ ] Merge to `main` only after full manual demo succeeds

## Branch Checklists

### `main` checklist

- [x] Server starts on `localhost:6379`
- [x] Multiple clients can connect concurrently
- [x] `PING` works
- [x] `SET` works
- [x] `GET` works
- [x] `DEL` works
- [x] `EXISTS` works
- [x] `EXPIRE` works
- [x] `TTL` works
- [x] Expired keys behave as missing keys
- [x] Background cleanup runs without breaking command behavior
- [x] Snapshot loads on startup
- [x] Snapshot saves on graceful shutdown
- [x] Debug logs help trace request lifecycle
- [x] Unit tests pass
- [x] Integration tests pass
- [x] Manual `nc` demo passes
- [x] README matches implemented feature set

### `develop` backlog after `main`

- [ ] RESP protocol support
- [ ] Append-only file persistence
- [x] `SETEX`
- [ ] `INCR`
- [ ] `DECR`
- [ ] List data type
- [ ] Hash data type
- [ ] Authentication
- [ ] Docker polish
- [ ] Benchmarks

## Test Plan

### Unit tests

- Parser:
  - valid command parsing
  - bad arity cases handled at executor/validator layer
  - command name normalization
- Store:
  - set/get existing and missing keys
  - delete and exists behavior
  - expire existing vs missing key
  - TTL returns positive, `-1`, `-2`
  - expired keys are treated as missing
  - `SET` removes previous TTL
  - cleanup deletes expired keys only
- Persistence:
  - save then load roundtrip
  - expired keys not persisted/restored

### Integration tests

- TCP server handles:
  - `PING`
  - `SET/GET`
  - `DEL/EXISTS`
  - `EXPIRE/TTL`
  - invalid command
  - invalid argument count
- One concurrent test with multiple clients issuing commands
- Optional race run using `go test -race ./...`

### Manual acceptance scenarios

- `PING` -> `PONG`
- `SET name chanyut` -> `OK`
- `GET name` -> `chanyut`
- `EXISTS name` -> `1`
- `DEL name` -> `1`
- `GET name` -> `(nil)`
- `EXPIRE token 5` -> `1`
- `TTL token` -> positive integer
- after expiry, `GET token` -> `(nil)`
- restart server, confirm non-expired key is restored
- debug mode shows enough logs to trace one full command end to end

## Assumptions

- This plan uses Head First principles as design guidance, not as a requirement to implement every named pattern formally.
- Go remains idiomatic Go first; do not introduce unnecessary interfaces or factories where a concrete type is simpler and clearer.
- RESP, AOF, richer data types, auth, and advanced observability remain out of scope for `main`.
- Values remain single-token strings in `main` to keep parser complexity aligned with the one-week scope.

# MiniRedisGo

MiniRedisGo is a Redis-inspired in-memory key-value database built with Go.
It supports TCP client connections, concurrent access, TTL-based expiration, and optional persistence to disk.

## Features

- TCP server
- In-memory key-value store
- Concurrent multi-client support
- Basic Redis-like commands
- TTL expiration
- Background cleanup for expired keys
- Snapshot-based persistence
- Easy manual testing with `nc` or `telnet`

## Tech Stack

- **Language:** Go
- **Protocol:** TCP
- **Concurrency:** Goroutines, `sync.RWMutex`
- **Storage:** In-memory map
- **Persistence:** JSON snapshot / Append-only log (planned)
- **Testing:** Go testing package
- **Containerization:** Docker

## Supported Commands

| Command | Description | Example |
| --- | --- | --- |
| `PING` | Check server status | `PING` |
| `SET key value` | Store value | `SET name chanyut` |
| `GET key` | Retrieve value | `GET name` |
| `DEL key` | Delete key | `DEL name` |
| `EXISTS key` | Check if key exists | `EXISTS name` |
| `EXPIRE key seconds` | Set expiration | `EXPIRE name 60` |
| `TTL key` | Get remaining TTL | `TTL name` |

## Example Session

```text
PING
PONG

SET name chanyut
OK

GET name
chanyut

EXISTS name
1

EXPIRE name 5
1

TTL name
4

GET name
chanyut

After expiration:

GET name
(nil)
```

## Architecture

```text
Client
  ↓
TCP Server Listener
  ↓
Connection Handler
  ↓
Command Parser
  ↓
Command Executor
  ↓
In-Memory Store
  ↓
Persistence Layer
```

## Project Structure

```text
mini-redis-go/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── server/
│   ├── protocol/
│   ├── command/
│   ├── store/
│   └── persistence/
├── data/
├── test/
├── Dockerfile
├── README.md
└── go.mod
```

## How to Run

1. Clone the project.

```bash
git clone https://github.com/your-username/mini-redis-go.git
cd mini-redis-go
```

2. Run the server.

```bash
go run ./cmd/server
```

By default, the server listens on `localhost:6379`.

## How to Test Manually

Using `nc`:

```bash
nc localhost 6379
```

Then type commands:

```text
PING
SET language go
GET language
EXPIRE language 10
TTL language
```

## Persistence

MiniRedisGo supports snapshot-based persistence.

- Data is stored in memory during runtime.
- Periodically or on shutdown, state can be saved to disk.
- On restart, state can be restored from a snapshot.

Planned upgrade:

- Append-only log (AOF)

## Design Decisions

### Why TCP instead of HTTP?

Redis-like systems typically use raw TCP for lightweight communication and long-lived connections.

### Why Go?

Go provides simple concurrency, strong standard library support, and excellent performance for systems projects.

### Why `sync.RWMutex`?

The store is shared across multiple clients, so synchronization is required to avoid race conditions.

### How does TTL work?

TTL is implemented using:

- Lazy expiration during key access
- Optional background cleanup

## Future Improvements

- RESP protocol support
- `SETEX`, `INCR`, `DECR`
- Append-only persistence
- List and hash data types
- Authentication
- Pub/Sub
- Replication
- Benchmarks

## Learning Goals

This project is built to practice:

- TCP networking
- Concurrency and synchronization
- Protocol parsing
- In-memory data structures
- Database persistence concepts
- Backend systems design

## Portfolio Description

Built a Redis-inspired in-memory key-value database in Go with a custom TCP protocol, concurrent client handling, TTL-based expiration, and persistence support.

## Suggested MVP Scope

To avoid overbuilding, start with this:

### Phase 1

- TCP server
- `PING`

### Phase 2

- `SET`
- `GET`
- `DEL`
- `EXISTS`

### Phase 3

- Concurrency with goroutines and mutex

### Phase 4

- `EXPIRE`
- `TTL`
- Lazy expiration

### Phase 5

- Background cleanup

### Phase 6

- Persistence snapshot

## Resume / Portfolio Summary

You can describe it like this:

> Built a Redis-inspired in-memory key-value database in Go using raw TCP sockets, goroutines, and mutex-based concurrency control, with TTL expiration and snapshot persistence.

Or shorter:

> Developed a mini Redis-like in-memory database in Go with concurrent client handling, key expiration, and disk persistence.

## Recommendation

The best next move is:

1. Finalize the SRS
2. Create project folders
3. Implement MVP commands first
4. Add TTL
5. Add persistence
6. Polish README and demo

Possible next deliverables:

- ER-ish data structure design
- Command spec
- Roadmap
- Starter Go code

## License

MIT

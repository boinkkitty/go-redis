# 🧠 In-Memory Redis Replica (with AOF Persistence)

A lightweight in-memory database written in Go that replicates basic Redis behavior. This server supports several core Redis commands and provides **append-only persistence** using a **mutex** to ensure thread safety.

---

## 🚀 Features

- 🧠 In-memory key-value store
- 💾 Append-Only File (AOF) persistence
- 🔒 Mutex-protected concurrent writes
- 🎯 Redis-compatible command interface
- 🛠️ Supports basic string and hash operations
- 📡 Listens on the default Redis port: `6379`

---

## 💬 Supported Commands

| Command | Description                           | Usage                                    |
|-------|---------------------------------------|------------------------------------------|
| `PING` | Health check (returns `PONG`)         | `PING [message]`                         |
| `SET` | Sets a string key                   | `SET key val`                            |
| `GET` | Gets a string value                 | `GET key`                                |
| `HSET` | Sets a field in a hash       | `HSET key field1 val1 [field2 val2 ...]` |
| `HGET` | Gets a field value in a hash | `HGET key field`                         |
| `HGETALL` | Gets all fields/values in hash| `HGETALL key`                             |

Example usage in `handlers.go`:
```go
var Handlers = map[string]func([]Value) Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
}

```

---

## Installation

### Server

1. Clone the repository:
   ```bash
   git clone https://github.com/boinkkitty/go-redis.git
   ```
2. Run the server:
   ```bash
   cd go-redis
   go run *.go
   ```

### Client

1. Run client:
   ```bash
   redis-cli
   ```
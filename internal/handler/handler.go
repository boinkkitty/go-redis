package handler

import (
	"sync"

	"github.com/boinkkitty/go-redis/internal/resp"
)

// Handlers maps command names to their handler functions.
var Handlers = map[string]func([]resp.Value) resp.Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
	// "DEL":     del,
}

// ping handles the PING command. Returns PONG or echoes the argument.
func ping(args []resp.Value) resp.Value {
	if len(args) == 0 {
		return resp.Value{Typ: "string", Str: "PONG"}
	}
	return resp.Value{Typ: "string", Str: args[0].Bulk}
}

// SETs stores string key-value pairs.
var SETs = map[string]string{}

// SETsMu protects SETs for concurrent access.
var SETsMu = sync.RWMutex{}

// set handles the SET command. Sets a string key to a value.
func set(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'SET' command"}
	}
	key := args[0].Bulk
	value := args[1].Bulk
	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()
	return resp.Value{Typ: "string", Str: "OK"}
}

// get handles the GET command. Gets the value of a string key.
func get(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'GET' command"}
	}
	key := args[0].Bulk
	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()
	if !ok {
		return resp.Value{Typ: "null"}
	}
	return resp.Value{Typ: "bulk", Bulk: value}
}

// HSETs stores hash key-value pairs.
var HSETs = map[string]map[string]string{}

// HSETsMu protects HSETs for concurrent access.
var HSETsMu = sync.RWMutex{}

// hset handles the HSET command. Sets fields in a hash.
func hset(args []resp.Value) resp.Value {
	if len(args) < 3 || len(args)%2 == 0 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'HSET' command"}
	}
	hashKey := args[0].Bulk
	HSETsMu.Lock()
	defer HSETsMu.Unlock()
	if _, ok := HSETs[hashKey]; !ok {
		HSETs[hashKey] = make(map[string]string)
	}
	for i := 1; i < len(args); i += 2 {
		field := args[i].Bulk
		value := args[i+1].Bulk
		HSETs[hashKey][field] = value
	}
	return resp.Value{Typ: "string", Str: "OK"}
}

// hget handles the HGET command. Gets a field from a hash.
func hget(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'HGET' command"}
	}
	hash := args[0].Bulk
	key := args[1].Bulk
	HSETsMu.RLock()
	value, ok := HSETs[hash][key]
	HSETsMu.RUnlock()
	if !ok {
		return resp.Value{Typ: "null"}
	}
	return resp.Value{Typ: "bulk", Bulk: value}
}

// hgetall handles the HGETALL command. Gets all fields and values from a hash.
func hgetall(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Typ: "error", Str: "ERR wrong number of arguments for 'HGETALL' command"}
	}
	hash := args[0].Bulk
	HSETsMu.RLock()
	fields, ok := HSETs[hash]
	HSETsMu.RUnlock()
	if !ok {
		return resp.Value{Typ: "array", Array: []resp.Value{}}
	}
	arr := make([]resp.Value, 0, len(fields)*2)
	for field, val := range fields {
		arr = append(arr, resp.Value{Typ: "bulk", Bulk: field})
		arr = append(arr, resp.Value{Typ: "bulk", Bulk: val})
	}
	return resp.Value{Typ: "array", Array: arr}
}

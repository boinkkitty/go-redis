package main

import "sync"

var Handlers = map[string]func([]Value) Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "String", str: "PONG"}
	}
	return Value{typ: "string", str: args[0].bulk}
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	key := args[0].bulk
	value := args[1].bulk

	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	key := args[0].bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

var HSETs = map[string]map[string]string{}
var HSETsMu = sync.RWMutex{}

func hset(args []Value) Value {
	if len(args) < 3 || len(args)%2 == 0 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	hashKey := args[0].bulk

	HSETsMu.Lock()
	defer HSETsMu.Unlock()

	// Create the hash if it doesn't exist yet
	if _, ok := HSETs[hashKey]; !ok {
		HSETs[hashKey] = make(map[string]string)
	}

	// Loop through each field-value pair
	for i := 1; i < len(args); i += 2 {
		field := args[i].bulk
		value := args[i+1].bulk
		HSETs[hashKey][field] = value
	}

	return Value{typ: "string", str: "OK"}
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	hash := args[0].bulk
	key := args[1].bulk

	HSETsMu.RLock()
	value, ok := HSETs[hash][key]
	HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

func hgetall(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	hash := args[0].bulk

	HSETsMu.RLock()
	fields, ok := HSETs[hash]
	HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "array", array: []Value{}}
	}

	arr := make([]Value, 0, len(fields)*2)
	for field, val := range fields {
		arr = append(arr, Value{typ: "bulk", bulk: field})
		arr = append(arr, Value{typ: "bulk", bulk: val})
	}

	return Value{typ: "array", array: arr}
}

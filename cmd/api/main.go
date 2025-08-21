// Entrypoint for the Redis-like server
package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/boinkkitty/go-redis/internal/aof"
	"github.com/boinkkitty/go-redis/internal/handler"
	"github.com/boinkkitty/go-redis/internal/resp"
)

const TCP_NETWORK = "tcp"
const REDIS_PORT = ":6379"

func main() {
	fmt.Println("Listening on port", REDIS_PORT)

	l, err := net.Listen(TCP_NETWORK, REDIS_PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	aofInst, err := aof.NewAof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aofInst.Close()

	aofInst.Read(func(value resp.Value) {
		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]
		handlerFunc, ok := handler.Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}
		handlerFunc(args)
	})

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn, aofInst)
	}
}

func handleConnection(conn net.Conn, aofInst *aof.Aof) {
	defer conn.Close()
	for {
		respReader := resp.NewResp(conn)
		value, err := respReader.Read()
		if err != nil {
			fmt.Println(err)
			return
		}
		if value.Typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}
		if len(value.Array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}
		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]
		writer := resp.NewWriter(conn)
		handlerFunc, ok := handler.Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(resp.Value{Typ: "string", Str: ""})
			continue
		}
		if command == "SET" || command == "HSET" {
			aofInst.Write(value)
		}
		result := handlerFunc(args)
		writer.Write(result)
	}
}

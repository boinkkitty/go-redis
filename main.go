package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

const TCP_NETWORK = "tcp"
const REDIS_PORT = ":6379"

func main() {
	fmt.Printf("Listening on port%s", REDIS_PORT)

	// Create a new server
	l, err := net.Listen(TCP_NETWORK, REDIS_PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Listen for connections
	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close() // Close connection once done

	for {
		buf := make([]byte, 1024)

		// read message from client
		_, err = conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error reading from client: ", err.Error())
			os.Exit(1)
		}

		// ignore request and send back a PONG
		conn.Write([]byte("+OK\r\n"))
	}
}

/*
func main() {
	input := "$5\r\nDaryl\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	b, _ := reader.ReadByte()

	if b != '$' {
		fmt.Println("Invalid type, expected bulk stirng")
		os.Exit(1)
	}

	size, _ := reader.ReadByte()
	strSize, _ := strconv.ParseInt(string(size), 10, 64)

	// Consume /r/n
	reader.ReadByte()
	reader.ReadByte()

	name := make([]byte, strSize)
	reader.Read(name)

	fmt.Println(string(name))
}
*/

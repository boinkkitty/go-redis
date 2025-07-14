.PHONY: server client build

server:
	go run server/*.go

client:
	redis-cli -p 6379

build:
	go build -o bin/server server/*.go

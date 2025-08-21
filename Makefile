.PHONY: server client build clean


server:
	go run ./cmd/api

client:
	redis-cli -p 6379

build:
	go build -o bin/server ./cmd/api

clean:
	rm -rf bin/

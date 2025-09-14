package aof

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"

	"github.com/boinkkitty/go-redis/internal/resp"
)

// Aof provides append-only file persistence for Redis-like commands.
// It is safe for concurrent use.
type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

// NewAof opens or creates an append-only file at the given path.
// It starts a background goroutine to periodically sync the file to disk.
func NewAof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}

	go func() {
		for {
			aof.mu.Lock()

			aof.file.Sync()
			aof.mu.Unlock()
			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

// Close closes the underlying append-only file.
func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.file.Close()
}

// Write appends a RESP value to the append-only file.
// It is safe for concurrent use.
func (aof *Aof) Write(value resp.Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

// Read replays all RESP values from the append-only file,
// calling the provided callback for each value.
// The callback is called in the order the values appear in the file.
func (aof *Aof) Read(callback func(value resp.Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	aof.file.Seek(0, io.SeekStart)
	reader := resp.NewResp(aof.file)
	for {
		value, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		callback(value)
	}
	return nil
}

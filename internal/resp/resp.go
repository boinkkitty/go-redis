// Package resp implements the RESP (REdis Serialization Protocol) parser and serializer.
package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// RESP type byte constants.
const (
	STRING  = '+' // Simple string
	ERROR   = '-' // Error
	INTEGER = ':' // Integer
	BULK    = '$' // Bulk string
	ARRAY   = '*' // Array
)

// Value represents a RESP value (string, error, integer, bulk, array).
type Value struct {
	Typ   string  // Type of value: "string", "error", "integer", "bulk", "array", "null"
	Str   string  // String value
	Num   int     // Integer value
	Bulk  string  // Bulk string value
	Array []Value // Array of values
}

// Resp reads RESP values from an io.Reader.
type Resp struct {
	reader *bufio.Reader
}

// NewResp creates a new RESP reader from an io.Reader.
func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

// readLine reads a line ending with \r\n from the underlying reader.
func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}

		// Add byte
		n += 1
		line = append(line, b)

		// Stop when we read \r\n
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}

	// Return without \r\n
	return line[:len(line)-2], n, nil
}

// readInteger reads an integer from the underlying reader.
func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}

	return int(i64), n, nil
}

// Read reads a RESP value from the underlying reader.
func (r *Resp) Read() (Value, error) {
	b, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch b {
	case STRING:
		line, _, err := r.readLine()
		if err != nil {
			return Value{}, err
		}
		return Value{Typ: "string", Str: string(line)}, nil
	case ERROR:
		line, _, err := r.readLine()
		if err != nil {
			return Value{}, err
		}
		return Value{Typ: "error", Str: string(line)}, nil
	case INTEGER:
		x, _, err := r.readInteger()
		if err != nil {
			return Value{}, err
		}
		return Value{Typ: "integer", Num: x}, nil
	case BULK:
		n, _, err := r.readInteger()
		if err != nil {
			return Value{}, err
		}
		if n == -1 {
			return Value{Typ: "null"}, nil
		}
		buf := make([]byte, n+2)
		_, err = io.ReadFull(r.reader, buf)
		if err != nil {
			return Value{}, err
		}
		return Value{Typ: "bulk", Bulk: string(buf[:n])}, nil
	case ARRAY:
		n, _, err := r.readInteger()
		if err != nil {
			return Value{}, err
		}
		if n == -1 {
			return Value{Typ: "null"}, nil
		}
		arr := make([]Value, n)
		for i := 0; i < n; i++ {
			v, err := r.Read()
			if err != nil {
				return Value{}, err
			}
			arr[i] = v
		}
		return Value{Typ: "array", Array: arr}, nil
	default:
		return Value{}, fmt.Errorf("invalid RESP type: %c", b)
	}
}

// Marshal serializes a Value into RESP bytes.
func (v Value) Marshal() []byte {
	switch v.Typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshalNull()
	case "error":
		return v.marshalError()
	default:
		return []byte{}
	}
}

// marshalString serializes a simple string Value.
func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

// marshalBulk serializes a bulk string Value.
func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.Bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Bulk...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

// marshalArray serializes an array Value.
func (v Value) marshalArray() []byte {
	var bytes []byte
	length := len(v.Array)
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(length)...)
	bytes = append(bytes, '\r', '\n')
	for i := 0; i < length; i++ {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}
	return bytes
}

// marshalError serializes an error Value.
func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

// marshalNull serializes a null Value.
func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

// Writer writes RESP values to an io.Writer.
type Writer struct {
	writer io.Writer
}

// NewWriter creates a new RESP writer from an io.Writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

// Write writes a Value to the underlying writer in RESP format.
func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

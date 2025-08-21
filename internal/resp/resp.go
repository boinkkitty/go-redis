package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

// For serialization and deserialization
type Value struct {
	Typ   string  // Exported
	Str   string  // Exported
	Num   int     // Exported
	Bulk  string  // Exported
	Array []Value // Exported
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

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

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.Bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Bulk...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

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

func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

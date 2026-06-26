package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

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
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
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
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch ValueType(_type) {
	case SimpleString:
		return r.readSimpleString()
	case BulkString:
		return r.readBulk()

	case Integer:
		return r.readIntegerResp()

	case BigNumber:

	case Double:

	case Boolean:
		return r.readBoolean()

	case Array:
		return r.readArray()

	case NULL:
		return r.readNull()

	case Map:
		return r.readMap()
	default:
	}
	return Value{}, fmt.Errorf("invalid case here")
}

func (r *Resp) readNull() (Value, error) {
	_, _, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	return Value{typ: NULL}, nil
}

func (r *Resp) readSimpleString() (Value, error) {
	line, _, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	return Value{typ: SimpleString, str: string(line)}, nil
}

func (r *Resp) readSimpleError() (val Value, err error) {
	// -Error message\r\n

	line, _, err := r.readLine()

	if err != nil {
		return val, err
	}

	return Value{typ: SimpleError, str: string(line)}, nil
}

func (r *Resp) readBulkError() (val Value, err error) {
	// !<length>\r\n<error>\r\n

	line, _, err := r.readLine()
	if err != nil {
		return val, err
	}
	number, err := strconv.Atoi(string(line))

	if err != nil {
		return val, err
	}

	buf := make([]byte, number)

	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return Value{}, err
	}
	if _, _, err := r.readLine(); err != nil {
		return Value{}, err
	}

	return Value{typ: BulkError, str: string(buf)}, nil

}

func (r *Resp) readBoolean() (val Value, err error) {
	// #<t|f>\r\n

	line, n, err := r.readLine()

	if err != nil {
		return Value{}, err
	}

	if n != 3 {
		return Value{}, fmt.Errorf("booleans expect t or f")
	}

	lineStr := string(line)

	if lineStr != "t" && lineStr != "f" {
		return Value{}, fmt.Errorf("booleans expect t or f")
	}

	return Value{typ: Boolean, boolean: lineStr == "t"}, nil
}

func (r *Resp) readIntegerResp() (val Value, err error) {

	// :[<+|->]<value>\r\n
	x, _, err := r.readInteger()

	if err != nil {
		return val, err
	}

	return Value{typ: Integer, num: x}, nil
}

func (r *Resp) readBulk() (Value, error) {
	// $<length>\r\n<data>\r\n
	size, _, err := r.readInteger()

	if err != nil {
		return Value{}, err
	}

	buf := make([]byte, size)
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return Value{}, err
	}
	if _, _, err := r.readLine(); err != nil {
		return Value{}, err
	}
	return Value{typ: BulkString, str: string(buf)}, nil

}

func (r *Resp) readArray() (Value, error) {
	// <number-of-elements>\r\n<element-1>...<element-n>
	v := Value{}

	len, _, err := r.readInteger()

	if err != nil {
		return v, err
	}

	v.typ = Array
	v.array = make([]Value, 0)

	for i := 0; i < len; i += 1 {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		v.array = append(v.array, val)

	}

	return v, nil

}

func (r *Resp) readMap() (Value, error) {
	len, _, err := r.readInteger()

	if err != nil {
		return Value{}, err
	}

	val := Value{}
	val.typ = Map
	val.MAP = make([][2]Value, 0)

	for i := 0; i < len; i += 1 {
		key, err := r.Read()

		if err != nil {
			return val, err
		}

		value, err := r.Read()
		if err != nil {
			return val, err
		}

		val.MAP = append(val.MAP, [2]Value{key, value})

	}

	return val, nil

}

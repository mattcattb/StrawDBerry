package redis

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type valueType byte

const (
	simpleString valueType = '+'
	simpleError  valueType = '-'
	integer      valueType = ':'
	bulkString   valueType = '$'
	bulkError    valueType = '!'
	array        valueType = '*'
	boolean      valueType = '#'
	mapValue     valueType = '%'
	nullValue    valueType = '_'
	double       valueType = ','
	bigNumber    valueType = '('
)

type Value struct {
	typ     valueType
	str     string
	num     int
	array   []Value
	boolean bool
	pairs   [][2]Value
}

func SimpleString(s string) Value {
	return Value{typ: simpleString, str: s}
}

func BulkString(s string) Value {
	return Value{typ: bulkString, str: s}
}

func Integer(n int) Value {
	return Value{typ: integer, num: n}
}

func Array(values []Value) Value {
	return Value{typ: array, array: values}
}

func Null() Value {
	return Value{typ: nullValue}
}

func SimpleError(s string) Value {
	return Value{typ: simpleError, str: s}
}

func Error(s string) Value {
	return SimpleError(s)
}

func Boolean(b bool) Value {
	return Value{typ: boolean, boolean: b}
}

func Map(pairs [][2]Value) Value {
	return Value{typ: mapValue, pairs: pairs}
}

func (v Value) Array() ([]Value, bool) {
	return v.array, v.typ == array
}

func (v Value) BulkString() (string, bool) {
	return v.str, v.typ == bulkString
}

func (v Value) SimpleString() (string, bool) {
	return v.str, v.typ == simpleString
}

func (v Value) Integer() (int, bool) {
	return v.num, v.typ == integer
}

func (v Value) IsNull() bool {
	return v.typ == nullValue
}

func (v Value) Map() ([][2]Value, bool) {
	return v.pairs, v.typ == mapValue
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
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			if line[len(line)-1] != '\n' {
				return nil, 0, ErrJsonDecode

			}
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

	switch valueType(_type) {
	case simpleString:
		return r.readSimpleString()
	case simpleError:
		return r.readSimpleError()
	case bulkString:
		return r.readBulk()
	case bulkError:
		return r.readBulkError()

	case integer:
		return r.readIntegerResp()

	case bigNumber:

	case double:

	case boolean:
		return r.readBoolean()

	case array:
		return r.readArray()

	case nullValue:
		return r.readNull()

	case mapValue:
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

	return Null(), nil
}

func (r *Resp) readSimpleString() (Value, error) {
	line, _, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	return SimpleString(string(line)), nil
}

func (r *Resp) readSimpleError() (val Value, err error) {
	// -Error message\r\n

	line, _, err := r.readLine()

	if err != nil {
		return val, err
	}

	return SimpleError(string(line)), nil
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

	return Value{typ: bulkError, str: string(buf)}, nil

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

	return Boolean(lineStr == "t"), nil
}

func (r *Resp) readIntegerResp() (val Value, err error) {

	// :[<+|->]<value>\r\n
	x, _, err := r.readInteger()

	if err != nil {
		return val, err
	}

	return Integer(x), nil
}

func (r *Resp) readBulk() (Value, error) {
	// $<length>\r\n<data>\r\n
	size, _, err := r.readInteger()

	if err != nil {
		return Value{}, err
	}

	if size == -1 {
		return Null(), nil
	}
	if size < -1 {
		return Value{}, fmt.Errorf("invalid bulk string length %d", size)
	}

	buf := make([]byte, size)
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return Value{}, err
	}
	if _, _, err := r.readLine(); err != nil {
		return Value{}, err
	}
	return BulkString(string(buf)), nil

}

func (r *Resp) readArray() (Value, error) {
	// <number-of-elements>\r\n<element-1>...<element-n>
	v := Value{}

	len, _, err := r.readInteger()

	if err != nil {
		return v, err
	}

	if len == -1 {
		return Null(), nil
	}
	if len < -1 {
		return Value{}, fmt.Errorf("invalid array length %d", len)
	}

	v.typ = array
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
	val.typ = mapValue
	val.pairs = make([][2]Value, 0)

	for i := 0; i < len; i += 1 {
		key, err := r.Read()

		if err != nil {
			return val, err
		}

		value, err := r.Read()
		if err != nil {
			return val, err
		}

		val.pairs = append(val.pairs, [2]Value{key, value})

	}

	return val, nil

}

func (v Value) Marshal() []byte {
	switch v.typ {

	case bulkError:
		return v.marshalBulkError()

	case simpleError:
		return v.marshalSimpleError()

	case array:
		return v.marshalArray()

	case integer:
		return v.marshalInteger()

	case bulkString:
		return v.marshalBulk()

	case simpleString:
		return v.marshalString()

	case boolean:
		return v.marshalBoolean()

	case nullValue:
		return v.marshalNull()

	case mapValue:
		return v.marshalMap()
	}

	return []byte{}
}
func (v Value) marshalSimpleError() []byte {

	// -Error message\r\n
	bytes := []byte{byte(v.typ)}

	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes

}

func (v Value) marshalBulkError() []byte {
	// !<length>\r\n<error>\r\n

	bytes := []byte{byte(v.typ)}

	bytes = append(bytes, (strconv.Itoa(len(v.str)))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	// *<number-of-elements>\r\n<element-1>...<element-n>

	bytes := []byte{byte(array)}
	bytes = append(bytes, (strconv.Itoa(len(v.array)))...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len(v.array); i += 1 {
		val := v.array[i]
		bytes = append(bytes, val.Marshal()...)
	}

	return bytes
}
func (v Value) marshalString() []byte {
	// +OK\r\n

	bytes := []byte{byte(simpleString)}

	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalBoolean() []byte {
	// #<t|f>\r\n
	bytes := []byte{byte(boolean)}

	var boolChar = "f"
	if v.boolean {
		boolChar = "t"
	}
	bytes = append(bytes, boolChar...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalBulk() []byte {
	// $<length>\r\n<data>\r\n

	bytes := []byte{byte(bulkString)}
	bytes = append(bytes, (strconv.Itoa(len(v.str)))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalInteger() []byte {
	// :[<+|->]<value>\r\n

	bytes := []byte{byte(integer)}

	bytes = append(bytes, (strconv.Itoa(v.num))...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalMap() []byte {
	// %<number-of-entries>\r\n<key-1><value-1>...<key-n><value-n>
	bytes := []byte{byte(mapValue)}

	numEntries := len(v.pairs)

	bytes = append(bytes, (strconv.Itoa(numEntries))...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < numEntries; i += 1 {
		entry := v.pairs[i]
		k := entry[0]
		v := entry[1]

		bytes = append(bytes, k.Marshal()...)
		bytes = append(bytes, v.Marshal()...)

	}

	return bytes

}
func (v Value) marshalNull() []byte {
	return []byte{byte(nullValue), '\r', '\n'}
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

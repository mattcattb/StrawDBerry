package resp

import (
	"io"
	"strconv"
)

func (v Value) Marshal() []byte {
	switch v.typ {

	case BulkError:
		return v.marshalBulkError()

	case SimpleError:
		return v.marshalSimpleError()

	case Array:
		return v.marshalArray()

	case Integer:
		return v.marshalInteger()

	case BulkString:
		return v.marshalBulk()

	case SimpleString:
		return v.marshalString()

	case Boolean:
		return v.marshalBoolean()

	case NULL:
		return v.marshalNull()

	case Map:
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

	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	// *<number-of-elements>\r\n<element-1>...<element-n>

	bytes := []byte{byte(Array)}
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

	bytes := []byte{byte(SimpleString)}

	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalBoolean() []byte {
	// #<t|f>\r\n
	bytes := []byte{byte(Boolean)}

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

	bytes := []byte{byte(BulkString)}
	bytes = append(bytes, (strconv.Itoa(len(v.str)))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalInteger() []byte {
	// :[<+|->]<value>\r\n

	bytes := []byte{byte(Integer)}

	var sign = "+"
	var val = v.num

	if v.num < 0 {
		val *= -1
		sign = "-"
	} else if v.num == 0 {
		sign = ""
	}

	bytes = append(bytes, sign...)
	bytes = append(bytes, (strconv.Itoa(val))...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalMap() []byte {
	// %<number-of-entries>\r\n<key-1><value-1>...<key-n><value-n>
	bytes := []byte{byte(Map)}

	numEntries := len(v.MAP)

	bytes = append(bytes, (strconv.Itoa(numEntries))...)

	for i := 0; i < numEntries; i += 1 {
		entry := v.MAP[i]
		k := entry[0]
		v := entry[1]

		bytes = append(bytes, k.Marshal()...)
		bytes = append(bytes, v.Marshal()...)

	}

	return bytes

}
func (v Value) marshalNull() []byte {
	return []byte{byte(NULL), '\r', '\n'}
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

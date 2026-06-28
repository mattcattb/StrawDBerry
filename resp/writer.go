package resp

import (
	"io"
	"strconv"
)

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

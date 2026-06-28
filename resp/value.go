package resp

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

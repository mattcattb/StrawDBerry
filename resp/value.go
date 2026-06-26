package resp

type ValueType byte

const (
	SimpleString ValueType = '+'
	SimpleError  ValueType = '-'
	Integer      ValueType = ':'
	BulkString   ValueType = '$'
	BulkError    ValueType = '!'
	Array        ValueType = '*'
	Boolean      ValueType = '#'
	Map          ValueType = '%'
	NULL         ValueType = '_'
	Double       ValueType = ','
	BigNumber    ValueType = '('
)

type Value struct {
	typ     ValueType
	str     string
	num     int
	array   []Value
	boolean bool
	MAP     [][2]Value
}

func (v Value) Array() ([]Value, bool) {
	return v.array, v.typ == Array
}

func (v Value) BulkString() (string, bool) {
	return v.str, v.typ == BulkString
}

func ErrorValue(s string) Value {
	return Value{typ: SimpleError}
}

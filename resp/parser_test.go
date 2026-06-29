package redis

import (
	"strings"
	"testing"
)

func TestReadSimpleString(t *testing.T) {
	r := NewResp(strings.NewReader("+OK\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	s, ok := value.SimpleString()
	if !ok {
		t.Fatal("value is not a simple string")
	}

	if s != "OK" {
		t.Fatalf("str = %q, want %q", s, "OK")
	}
}

func TestReadInteger(t *testing.T) {
	r := NewResp(strings.NewReader(":-42\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	n, ok := value.Integer()
	if !ok {
		t.Fatal("value is not an integer")
	}

	if n != -42 {
		t.Fatalf("num = %d, want %d", n, -42)
	}
}

func TestReadBulkString(t *testing.T) {
	r := NewResp(strings.NewReader("$5\r\nhello\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	s, ok := value.BulkString()
	if !ok {
		t.Fatal("value is not a bulk string")
	}

	if s != "hello" {
		t.Fatalf("bulk = %q, want %q", s, "hello")
	}
}

func TestReadNullBulkString(t *testing.T) {
	r := NewResp(strings.NewReader("$-1\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	if !value.IsNull() {
		t.Fatalf("value = %#v, want null", value)
	}
}

func TestReadArray(t *testing.T) {
	r := NewResp(strings.NewReader("*2\r\n+PING\r\n$5\r\nhello\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	values, ok := value.Array()
	if !ok {
		t.Fatal("value is not an array")
	}

	if len(values) != 2 {
		t.Fatalf("len(array) = %d, want %d", len(values), 2)
	}

	s, ok := values[0].SimpleString()
	if !ok || s != "PING" {
		t.Fatalf("first array value = %q, want %q", s, "PING")
	}

	s, ok = values[1].BulkString()
	if !ok || s != "hello" {
		t.Fatalf("second array value = %q, want %q", s, "hello")
	}
}

func TestReadMap(t *testing.T) {
	r := NewResp(strings.NewReader("%1\r\n+name\r\n$3\r\nmat\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	pairs, ok := value.Map()
	if !ok {
		t.Fatal("value is not a map")
	}

	if len(pairs) != 1 {
		t.Fatalf("len(map) = %d, want %d", len(pairs), 1)
	}

	pair := pairs[0]

	key, ok := pair[0].SimpleString()
	if !ok || key != "name" {
		t.Fatalf("map key = %q, want %q", key, "name")
	}

	val, ok := pair[1].BulkString()
	if !ok || val != "mat" {
		t.Fatalf("map value = %q, want %q", val, "mat")
	}
}

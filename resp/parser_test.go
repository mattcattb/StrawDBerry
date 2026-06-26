package resp

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

	if value.typ != "string" {
		t.Fatalf("typ = %q, want %q", value.typ, "string")
	}

	if value.str != "OK" {
		t.Fatalf("str = %q, want %q", value.str, "OK")
	}
}

func TestReadInteger(t *testing.T) {
	r := NewResp(strings.NewReader(":-42\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	if value.typ != "integer" {
		t.Fatalf("typ = %q, want %q", value.typ, "integer")
	}

	if value.num != -42 {
		t.Fatalf("num = %d, want %d", value.num, -42)
	}
}

func TestReadBulkString(t *testing.T) {
	r := NewResp(strings.NewReader("$5\r\nhello\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	if value.typ != "bulk-string" {
		t.Fatalf("typ = %q, want %q", value.typ, "bulk-string")
	}

	if value.bulk != "hello" {
		t.Fatalf("bulk = %q, want %q", value.bulk, "hello")
	}
}

func TestReadArray(t *testing.T) {
	r := NewResp(strings.NewReader("*2\r\n+PING\r\n$5\r\nhello\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	if value.typ != "array" {
		t.Fatalf("typ = %q, want %q", value.typ, "array")
	}

	if len(value.array) != 2 {
		t.Fatalf("len(array) = %d, want %d", len(value.array), 2)
	}

	if value.array[0].str != "PING" {
		t.Fatalf("first array value = %q, want %q", value.array[0].str, "PING")
	}

	if value.array[1].bulk != "hello" {
		t.Fatalf("second array value = %q, want %q", value.array[1].bulk, "hello")
	}
}

func TestReadMap(t *testing.T) {
	r := NewResp(strings.NewReader("%1\r\n+name\r\n$3\r\nmat\r\n"))

	value, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	if value.typ != "map" {
		t.Fatalf("typ = %q, want %q", value.typ, "map")
	}

	if len(value.MAP) != 1 {
		t.Fatalf("len(MAP) = %d, want %d", len(value.MAP), 1)
	}

	pair := value.MAP[0]

	if pair[0].str != "name" {
		t.Fatalf("map key = %q, want %q", pair[0].str, "name")
	}

	if pair[1].bulk != "mat" {
		t.Fatalf("map value = %q, want %q", pair[1].bulk, "mat")
	}
}

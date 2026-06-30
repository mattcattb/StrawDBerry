package redis

import (
	"strings"
	"testing"
)

func TestReadBasicRESPValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, v Value)
	}{
		{
			name:  "simple string",
			input: "+OK\r\n",
			check: func(t *testing.T, v Value) {
				t.Helper()
				got, ok := v.SimpleString()
				if !ok || got != "OK" {
					t.Fatalf("SimpleString() = %q, %v; want %q, true", got, ok, "OK")
				}
			},
		},
		{
			name:  "integer",
			input: ":-42\r\n",
			check: func(t *testing.T, v Value) {
				t.Helper()
				got, ok := v.Integer()
				if !ok || got != -42 {
					t.Fatalf("Integer() = %d, %v; want %d, true", got, ok, -42)
				}
			},
		},
		{
			name:  "bulk string",
			input: "$5\r\nhello\r\n",
			check: func(t *testing.T, v Value) {
				t.Helper()
				got, ok := v.BulkString()
				if !ok || got != "hello" {
					t.Fatalf("BulkString() = %q, %v; want %q, true", got, ok, "hello")
				}
			},
		},
		{
			name:  "null bulk string",
			input: "$-1\r\n",
			check: func(t *testing.T, v Value) {
				t.Helper()
				if !v.IsNull() {
					t.Fatalf("IsNull() = false; value = %#v", v)
				}
			},
		},
		{
			name:  "boolean true",
			input: "#t\r\n",
			check: func(t *testing.T, v Value) {
				t.Helper()
				if got := string(v.Marshal()); got != "#t\r\n" {
					t.Fatalf("Marshal() = %q, want %q", got, "#t\r\n")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResp(strings.NewReader(tt.input))

			got, err := r.Read()
			if err != nil {
				t.Fatal(err)
			}

			tt.check(t, got)
		})
	}
}

func TestReadArrayPreservesElementOrder(t *testing.T) {
	r := NewResp(strings.NewReader("*3\r\n$3\r\nGET\r\n$4\r\nname\r\n:7\r\n"))

	got, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	values, ok := got.Array()
	if !ok {
		t.Fatalf("Array() ok = false; value = %#v", got)
	}
	if len(values) != 3 {
		t.Fatalf("len(values) = %d, want 3", len(values))
	}

	command, ok := values[0].BulkString()
	if !ok || command != "GET" {
		t.Fatalf("values[0] = %q, %v; want %q, true", command, ok, "GET")
	}

	key, ok := values[1].BulkString()
	if !ok || key != "name" {
		t.Fatalf("values[1] = %q, %v; want %q, true", key, ok, "name")
	}

	n, ok := values[2].Integer()
	if !ok || n != 7 {
		t.Fatalf("values[2] = %d, %v; want %d, true", n, ok, 7)
	}

}

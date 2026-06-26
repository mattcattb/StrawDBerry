package resp

import "testing"

func TestMarshalSimpleString(t *testing.T) {
	value := Value{typ: "string", str: "OK"}

	got := string(value.Marshal())
	want := "+OK\r\n"

	if got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

func TestMarshalBulkString(t *testing.T) {
	value := Value{typ: "bulk-string", bulk: "hello"}

	got := string(value.Marshal())
	want := "$5\r\nhello\r\n"

	if got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

func TestMarshalArray(t *testing.T) {
	value := Value{
		typ: "array",
		array: []Value{
			{typ: "string", str: "PING"},
			{typ: "integer", num: 123},
		},
	}

	got := string(value.Marshal())
	want := "*2\r\n+PING\r\n:123\r\n"

	if got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

package resp

import "testing"

func TestMarshalSimpleString(t *testing.T) {
	value := SimpleString("OK")

	got := string(value.Marshal())
	want := "+OK\r\n"

	if got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

func TestMarshalBulkString(t *testing.T) {
	value := BulkString("hello")

	got := string(value.Marshal())
	want := "$5\r\nhello\r\n"

	if got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

func TestMarshalArray(t *testing.T) {
	value := Array([]Value{
		SimpleString("PING"),
		Integer(123),
	})

	got := string(value.Marshal())
	want := "*2\r\n+PING\r\n:123\r\n"

	if got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

func TestMarshalError(t *testing.T) {
	value := SimpleError("ERR bad things happened")

	got := string(value.Marshal())
	want := "-ERR bad things happened\r\n"

	if got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

func TestMarshalNull(t *testing.T) {
	value := Null()

	got := string(value.Marshal())
	want := "_\r\n"

	if got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

func TestMarshalMap(t *testing.T) {
	value := Map([][2]Value{
		{SimpleString("name"), BulkString("mat")},
	})

	got := string(value.Marshal())
	want := "%1\r\n+name\r\n$3\r\nmat\r\n"

	if got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}

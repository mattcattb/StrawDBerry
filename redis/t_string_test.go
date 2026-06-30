package redis

import "testing"

func newStringCommandTestClient() *Client {
	db := NewDb()
	server := NewServer(NewExec(db), nil, NewSH())

	return &Client{
		server: server,
		db:     db,
		tx:     &Transacion{},
		mode:   ModeNormal,
	}
}

func TestStringCommandSetThenGet(t *testing.T) {
	c := newStringCommandTestClient()

	set := Set(c, []Value{
		BulkString("name"),
		BulkString("mat"),
	})
	if set.Failed {
		t.Fatalf("SET failed: %#v", set.Reply)
	}

	gotStatus, ok := set.Reply.SimpleString()
	if !ok || gotStatus != "OK" {
		t.Fatalf("SET reply = %q, %v; want %q, true", gotStatus, ok, "OK")
	}

	get := Get(c, []Value{BulkString("name")})
	if get.Failed {
		t.Fatalf("GET failed: %#v", get.Reply)
	}

	got, ok := get.Reply.BulkString()
	if !ok || got != "mat" {
		t.Fatalf("GET reply = %q, %v; want %q, true", got, ok, "mat")
	}
}

func TestStringCommandGetMissingKeyReturnsNull(t *testing.T) {
	c := newStringCommandTestClient()

	got := Get(c, []Value{BulkString("missing")})
	if got.Failed {
		t.Fatalf("GET failed: %#v", got.Reply)
	}
	if !got.Reply.IsNull() {
		t.Fatalf("GET missing key reply = %#v, want null", got.Reply)
	}
}

func TestStringCommandIncrCreatesAndIncrementsValue(t *testing.T) {
	c := newStringCommandTestClient()

	first := Incr(c, []Value{BulkString("count")})
	got, ok := first.Reply.Integer()
	if first.Failed || !ok || got != 1 {
		t.Fatalf("first INCR = %d, failed=%v, ok=%v; want 1, false, true", got, first.Failed, ok)
	}

	second := Incr(c, []Value{BulkString("count")})
	got, ok = second.Reply.Integer()
	if second.Failed || !ok || got != 2 {
		t.Fatalf("second INCR = %d, failed=%v, ok=%v; want 2, false, true", got, second.Failed, ok)
	}
}

func TestStringCommandMGetPreservesOrderAndMissingValues(t *testing.T) {
	c := newStringCommandTestClient()

	Set(c, []Value{BulkString("first"), BulkString("one")})
	Set(c, []Value{BulkString("third"), BulkString("three")})

	res := MGet(c, []Value{
		BulkString("first"),
		BulkString("second"),
		BulkString("third"),
	})
	if res.Failed {
		t.Fatalf("MGET failed: %#v", res.Reply)
	}

	values, ok := res.Reply.Array()
	if !ok {
		t.Fatalf("MGET reply is not array: %#v", res.Reply)
	}
	if len(values) != 3 {
		t.Fatalf("len(values) = %d, want 3", len(values))
	}

	first, ok := values[0].BulkString()
	if !ok || first != "one" {
		t.Fatalf("values[0] = %q, %v; want %q, true", first, ok, "one")
	}
	if !values[1].IsNull() {
		t.Fatalf("values[1] = %#v, want null", values[1])
	}
	third, ok := values[2].BulkString()
	if !ok || third != "three" {
		t.Fatalf("values[2] = %q, %v; want %q, true", third, ok, "three")
	}
}

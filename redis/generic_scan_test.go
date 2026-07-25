package redis

import "testing"

func keyNames(t *testing.T, value Value) []string {
	t.Helper()
	values, ok := value.Array()
	if !ok {
		t.Fatalf("reply = %#v, want array", value)
	}

	keys := make([]string, len(values))
	for i, value := range values {
		key, ok := value.BulkString()
		if !ok {
			t.Fatalf("key %d = %#v, want bulk string", i, value)
		}
		keys[i] = key
	}
	return keys
}

func scanPage(t *testing.T, result CommandResult) (string, []string) {
	t.Helper()
	if result.Failed {
		t.Fatalf("SCAN failed: %#v", result.Reply)
	}

	values, ok := result.Reply.Array()
	if !ok || len(values) != 2 {
		t.Fatalf("SCAN reply = %#v, want [cursor, keys]", result.Reply)
	}
	cursor, ok := values[0].BulkString()
	if !ok {
		t.Fatalf("cursor = %#v, want bulk string", values[0])
	}
	return cursor, keyNames(t, values[1])
}

func TestKeysMatchesRedisGlobAndReturnsSortedKeys(t *testing.T) {
	c := newStringCommandTestClient()
	Set(c, []string{"user:2", "two"})
	Set(c, []string{"session:1", "one"})
	Set(c, []string{"user:1", "one"})

	result := Keys(c, []string{"user:*"})
	if result.Failed {
		t.Fatalf("KEYS failed: %#v", result.Reply)
	}
	got := keyNames(t, result.Reply)
	if len(got) != 2 || got[0] != "user:1" || got[1] != "user:2" {
		t.Fatalf("KEYS = %#v, want [user:1 user:2]", got)
	}
}

func TestScanPagesMatchingKeysWithStringCursors(t *testing.T) {
	c := newStringCommandTestClient()
	for _, key := range []string{"user:3", "session:1", "user:1", "user:2", "user:4"} {
		Set(c, []string{key, "value"})
	}

	cursor, keys := scanPage(t, Scan(c, []string{"0", "MATCH", "user:*", "COUNT", "2"}))
	if cursor != "2" || len(keys) != 2 || keys[0] != "user:1" || keys[1] != "user:2" {
		t.Fatalf("first SCAN page = cursor %q, keys %#v; want cursor 2 and first two users", cursor, keys)
	}

	cursor, keys = scanPage(t, Scan(c, []string{cursor, "MATCH", "user:*", "COUNT", "2"}))
	if cursor != "0" || len(keys) != 2 || keys[0] != "user:3" || keys[1] != "user:4" {
		t.Fatalf("second SCAN page = cursor %q, keys %#v; want cursor 0 and remaining users", cursor, keys)
	}
}

func TestScanRejectsInvalidCursorAndCount(t *testing.T) {
	c := newStringCommandTestClient()
	for _, args := range [][]string{{"invalid"}, {"0", "COUNT", "0"}} {
		if result := Scan(c, args); !result.Failed {
			t.Fatalf("SCAN %#v unexpectedly succeeded", args)
		}
	}
}

func TestScanIsRegisteredAndReturnsThePageResponseShape(t *testing.T) {
	c := newStringCommandTestClient()
	Set(c, []string{"user:1", "one"})

	result := c.HandleCommand(Array([]Value{
		BulkString("SCAN"),
		BulkString("0"),
		BulkString("MATCH"),
		BulkString("user:*"),
		BulkString("COUNT"),
		BulkString("100"),
	}))
	cursor, keys := scanPage(t, result)
	if cursor != "0" || len(keys) != 1 || keys[0] != "user:1" {
		t.Fatalf("registered SCAN = cursor %q, keys %#v; want cursor 0 and user:1", cursor, keys)
	}
}

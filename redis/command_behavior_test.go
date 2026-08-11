package redis

import (
	"sort"
	"testing"
)

type recordingAof struct {
	commands []Value
}

func (a *recordingAof) Append(v Value) error {
	a.commands = append(a.commands, v)
	return nil
}

func redisCommand(parts ...string) Value {
	values := make([]Value, len(parts))
	for i, part := range parts {
		values[i] = BulkString(part)
	}
	return Array(values)
}

func replyInteger(t *testing.T, result CommandResult) int {
	t.Helper()
	if result.Failed {
		t.Fatalf("command failed: %#v", result.Reply)
	}
	value, ok := result.Reply.Integer()
	if !ok {
		t.Fatalf("reply = %#v, want integer", result.Reply)
	}
	return value
}

func replyMembers(t *testing.T, result CommandResult) []string {
	t.Helper()
	if result.Failed {
		t.Fatalf("command failed: %#v", result.Reply)
	}
	values, ok := result.Reply.Array()
	if !ok {
		t.Fatalf("reply = %#v, want array", result.Reply)
	}
	members := make([]string, len(values))
	for i, value := range values {
		member, ok := value.BulkString()
		if !ok {
			t.Fatalf("member %d = %#v, want bulk string", i, value)
		}
		members[i] = member
	}
	sort.Strings(members)
	return members
}

func requireMembers(t *testing.T, result CommandResult, want ...string) {
	t.Helper()
	got := replyMembers(t, result)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("members = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("members = %#v, want %#v", got, want)
		}
	}
}

func TestHGetIsRegistered(t *testing.T) {
	c := newStringCommandTestClient()
	HSet(c, []string{"profile", "name", "mat"})

	result := c.HandleCommand(redisCommand("HGET", "profile", "name"))
	if result.Failed {
		t.Fatalf("HGET failed: %#v", result.Reply)
	}
	value, ok := result.Reply.BulkString()
	if !ok || value != "mat" {
		t.Fatalf("HGET reply = %q, %v; want %q, true", value, ok, "mat")
	}
}

func TestHSetCountsOnlyNewFields(t *testing.T) {
	c := newStringCommandTestClient()

	if got := replyInteger(t, HSet(c, []string{"profile", "name", "mat", "name", "matt", "role", "admin"})); got != 2 {
		t.Fatalf("first HSET count = %d, want 2", got)
	}
	if got := replyInteger(t, HSet(c, []string{"profile", "name", "matthew", "active", "yes"})); got != 1 {
		t.Fatalf("second HSET count = %d, want 1", got)
	}
}

func TestHDelAcceptsMultipleFieldsAndDeletesEmptyHash(t *testing.T) {
	c := newStringCommandTestClient()
	HSet(c, []string{"profile", "name", "mat", "role", "admin"})

	result := c.HandleCommand(redisCommand("HDEL", "profile", "name", "role"))
	if got := replyInteger(t, result); got != 2 {
		t.Fatalf("HDEL count = %d, want 2", got)
	}
	if _, exists := c.db.lookupKey("profile"); exists {
		t.Fatal("hash key still exists after deleting its final fields")
	}
}

func TestSRemReturnsDeletedCountPersistsAndDeletesEmptySet(t *testing.T) {
	c := newStringCommandTestClient()
	log := &recordingAof{}
	c.aof = log
	SAdd(c, []string{"tags", "go", "redis"})

	result := c.HandleCommand(redisCommand("SREM", "tags", "go", "redis", "missing"))
	if got := replyInteger(t, result); got != 2 {
		t.Fatalf("SREM count = %d, want 2", got)
	}
	if len(log.commands) != 1 {
		t.Fatalf("AOF command count = %d, want 1", len(log.commands))
	}
	if _, exists := c.db.lookupKey("tags"); exists {
		t.Fatal("set key still exists after deleting its final members")
	}
}

func TestSetMembershipAndAlgebraCommands(t *testing.T) {
	c := newStringCommandTestClient()
	SAdd(c, []string{"first", "a", "b", "c"})
	SAdd(c, []string{"second", "b", "c", "d"})
	SAdd(c, []string{"third", "c", "d", "e"})

	requireMembers(t, c.HandleCommand(redisCommand("SMEMBERS", "first")), "a", "b", "c")
	requireMembers(t, c.HandleCommand(redisCommand("SDIFF", "first", "second")), "a")
	requireMembers(t, c.HandleCommand(redisCommand("SINTER", "first", "second", "third")), "c")
	requireMembers(t, c.HandleCommand(redisCommand("SUNION", "first", "second", "third")), "a", "b", "c", "d", "e")

	requireMembers(t, c.HandleCommand(redisCommand("SMEMBERS", "missing")))
	requireMembers(t, c.HandleCommand(redisCommand("SINTER", "first", "missing")))
	requireMembers(t, c.HandleCommand(redisCommand("SUNION", "first", "missing")), "a", "b", "c")
}

func TestSetAlgebraRejectsWrongTypes(t *testing.T) {
	c := newStringCommandTestClient()
	SAdd(c, []string{"set", "member"})
	Set(c, []string{"string", "value"})

	for _, command := range [][]string{
		{"SMEMBERS", "string"},
		{"SDIFF", "set", "string"},
		{"SINTER", "set", "string"},
		{"SUNION", "set", "string"},
	} {
		t.Run(command[0], func(t *testing.T) {
			if result := c.HandleCommand(redisCommand(command...)); !result.Failed {
				t.Fatalf("%v unexpectedly succeeded", command)
			}
		})
	}
}

func TestSetGetReturnsOldValueAndKeepTTLRetainsExpiration(t *testing.T) {
	c := newStringCommandTestClient()
	if result := c.HandleCommand(redisCommand("SET", "name", "before", "PX", "5000")); result.Failed {
		t.Fatalf("initial SET failed: %#v", result.Reply)
	}
	before, _ := c.db.lookupKey("name")
	expiresAt := before.expiresAt

	result := c.HandleCommand(redisCommand("SET", "name", "after", "GET", "KEEPTTL"))
	if result.Failed {
		t.Fatalf("SET GET KEEPTTL failed: %#v", result.Reply)
	}
	old, ok := result.Reply.BulkString()
	if !ok || old != "before" {
		t.Fatalf("SET GET reply = %q, %v; want %q, true", old, ok, "before")
	}

	after, _ := c.db.lookupKey("name")
	if after.expiresAt != expiresAt {
		t.Fatalf("expiresAt = %d, want retained value %d", after.expiresAt, expiresAt)
	}
}

func TestSetRejectsUnknownConflictingAndNonPositiveExpirationOptions(t *testing.T) {
	tests := [][]string{
		{"SET", "unknown", "value", "BOGUS"},
		{"SET", "conflicting", "value", "EX", "1", "PX", "2"},
		{"SET", "zero", "value", "PX", "0"},
	}

	for _, command := range tests {
		t.Run(command[1], func(t *testing.T) {
			c := newStringCommandTestClient()
			if result := c.HandleCommand(redisCommand(command...)); !result.Failed {
				t.Fatalf("%v unexpectedly succeeded", command)
			}
			if _, exists := c.db.lookupKey(command[1]); exists {
				t.Fatalf("%q was created by rejected SET", command[1])
			}
		})
	}
}

func TestPersistOnlyMutatesAKeyWithExpiration(t *testing.T) {
	c := newStringCommandTestClient()
	log := &recordingAof{}
	c.aof = log
	Set(c, []string{"session", "value", "PX", "5000"})

	if got := replyInteger(t, c.HandleCommand(redisCommand("PERSIST", "session"))); got != 1 {
		t.Fatalf("first PERSIST = %d, want 1", got)
	}
	if got := replyInteger(t, c.HandleCommand(redisCommand("PERSIST", "session"))); got != 0 {
		t.Fatalf("second PERSIST = %d, want 0", got)
	}
	if len(log.commands) != 1 {
		t.Fatalf("AOF command count = %d, want 1", len(log.commands))
	}
}

func TestFlushAllIsReplayedByAOF(t *testing.T) {
	c := newStringCommandTestClient()
	log := &recordingAof{}
	c.aof = log

	c.HandleCommand(redisCommand("SET", "ghost", "value"))
	c.HandleCommand(redisCommand("FLUSHALL"))
	if len(log.commands) != 2 {
		t.Fatalf("AOF command count = %d, want SET and FLUSHALL", len(log.commands))
	}

	replayed := newStringCommandTestClient()
	for _, command := range log.commands {
		replayed.HandleCommand(command)
	}
	if snapshot := replayed.db.StatsSnapshot(); snapshot.keys != 0 {
		t.Fatalf("replayed key count = %d, want 0", snapshot.keys)
	}
}

func TestPingRejectsMoreThanOneArgument(t *testing.T) {
	c := newStringCommandTestClient()
	if result := c.HandleCommand(redisCommand("PING", "one", "two")); !result.Failed {
		t.Fatalf("PING with two arguments unexpectedly succeeded: %#v", result.Reply)
	}
}

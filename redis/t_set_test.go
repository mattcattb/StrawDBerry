package redis

import (
	"errors"
	"slices"
	"testing"
)

func setObjectForTest(t *testing.T, members ...string) *RedisObject {
	t.Helper()
	obj := newSetObj()
	if _, err := setTypeAdd(obj, members...); err != nil {
		t.Fatalf("setTypeAdd(%q): %v", members, err)
	}
	return obj
}

func requireStringMembers(t *testing.T, got []string, want ...string) {
	t.Helper()
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("members = %q, want %q", got, want)
	}
}

func replyIntegers(t *testing.T, result CommandResult) []int {
	t.Helper()
	if result.Failed {
		t.Fatalf("command failed: %#v", result.Reply)
	}
	values, ok := result.Reply.Array()
	if !ok {
		t.Fatalf("reply = %#v, want array", result.Reply)
	}

	integers := make([]int, len(values))
	for i, value := range values {
		integer, ok := value.Integer()
		if !ok {
			t.Fatalf("reply element %d = %#v, want integer", i, value)
		}
		integers[i] = integer
	}
	return integers
}

func TestSetTypeMutationsReportActualChanges(t *testing.T) {
	obj := newSetObj()

	added, err := setTypeAdd(obj, "", "go", "go")
	if err != nil {
		t.Fatalf("setTypeAdd: %v", err)
	}
	if added != 2 {
		t.Fatalf("added = %d, want 2", added)
	}

	containsEmpty, err := setTypeContains(obj, "")
	if err != nil {
		t.Fatalf("setTypeContains: %v", err)
	}
	if !containsEmpty {
		t.Fatal("empty string member was not retained")
	}

	removed, err := setTypeRemove(obj, "go", "go", "missing")
	if err != nil {
		t.Fatalf("setTypeRemove: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}

	length, err := setTypeCardinality(obj)
	if err != nil {
		t.Fatalf("setTypeCardinality: %v", err)
	}
	if length != 1 {
		t.Fatalf("cardinality = %d, want 1", length)
	}
}

func TestSetTypeOperationsDistinguishWrongTypeAndInvalidEncoding(t *testing.T) {
	if _, err := setTypeCardinality(newStringObject("value")); !errors.Is(err, ErrWrongType) {
		t.Fatalf("string cardinality error = %v, want ErrWrongType", err)
	}

	invalid := &RedisObject{
		typ:       SetObject,
		encoding:  EncodingSetMap,
		ptr:       hashMapPayload{},
		expiresAt: noExpiration,
	}
	if _, err := setTypeCardinality(invalid); !errors.Is(err, ErrInvalidEncoding) {
		t.Fatalf("invalid payload error = %v, want ErrInvalidEncoding", err)
	}
}

func TestSetTypeAlgebraHandlesMissingAndRepeatedOperands(t *testing.T) {
	first := setObjectForTest(t, "a", "b", "c")
	second := setObjectForTest(t, "b", "c", "d")
	third := setObjectForTest(t, "c", "e")

	difference, err := setTypeDiff(first, second, third)
	if err != nil {
		t.Fatalf("setTypeDiff: %v", err)
	}
	requireStringMembers(t, difference, "a")

	difference, err = setTypeDiff(first, first)
	if err != nil {
		t.Fatalf("setTypeDiff with repeated operand: %v", err)
	}
	requireStringMembers(t, difference)

	intersection, err := setTypeInter(first, nil, second)
	if err != nil {
		t.Fatalf("setTypeInter with missing operand: %v", err)
	}
	requireStringMembers(t, intersection)

	union, err := setTypeUnion(first, nil, second, third)
	if err != nil {
		t.Fatalf("setTypeUnion: %v", err)
	}
	requireStringMembers(t, union, "a", "b", "c", "d", "e")
}

func TestSetAlgebraValidatesAllPresentKeysBeforeComputing(t *testing.T) {
	c := newStringCommandTestClient()
	Set(c, []string{"string", "value"})

	for _, command := range [][]string{
		{"SDIFF", "missing", "string"},
		{"SINTER", "missing", "string"},
		{"SUNION", "missing", "string"},
	} {
		t.Run(command[0], func(t *testing.T) {
			if result := c.HandleCommand(redisCommand(command...)); !result.Failed {
				t.Fatalf("%v unexpectedly succeeded", command)
			}
		})
	}
}

func TestSetCardinalityAndMembershipCommands(t *testing.T) {
	c := newStringCommandTestClient()
	if got := replyInteger(t, c.HandleCommand(redisCommand("SADD", "set", "", "go", "go"))); got != 2 {
		t.Fatalf("SADD count = %d, want 2", got)
	}
	if got := replyInteger(t, c.HandleCommand(redisCommand("SCARD", "set"))); got != 2 {
		t.Fatalf("SCARD = %d, want 2", got)
	}
	if got := replyInteger(t, c.HandleCommand(redisCommand("SISMEMBER", "set", ""))); got != 1 {
		t.Fatalf("SISMEMBER empty string = %d, want 1", got)
	}
	if got := replyInteger(t, c.HandleCommand(redisCommand("SISMEMBER", "missing", "go"))); got != 0 {
		t.Fatalf("missing SISMEMBER = %d, want 0", got)
	}

	if got := replyIntegers(t, c.HandleCommand(redisCommand("SMISMEMBER", "set", "", "missing", "go"))); !slices.Equal(got, []int{1, 0, 1}) {
		t.Fatalf("SMISMEMBER = %v, want [1 0 1]", got)
	}
	if got := replyIntegers(t, c.HandleCommand(redisCommand("SMISMEMBER", "missing", "a", "b"))); !slices.Equal(got, []int{0, 0}) {
		t.Fatalf("missing SMISMEMBER = %v, want [0 0]", got)
	}
}

func TestSetCommandsRejectWrongTypeWithoutMutation(t *testing.T) {
	c := newStringCommandTestClient()
	Set(c, []string{"string", "value"})
	dirtyBefore := c.server.dirty

	for _, command := range [][]string{
		{"SADD", "string", "member"},
		{"SCARD", "string"},
		{"SREM", "string", "member"},
		{"SISMEMBER", "string", "member"},
		{"SMISMEMBER", "string", "member"},
	} {
		t.Run(command[0], func(t *testing.T) {
			if result := c.HandleCommand(redisCommand(command...)); !result.Failed {
				t.Fatalf("%v unexpectedly succeeded", command)
			}
		})
	}

	if c.server.dirty != dirtyBefore {
		t.Fatalf("dirty = %d, want unchanged value %d", c.server.dirty, dirtyBefore)
	}
	result := Get(c, []string{"string"})
	value, ok := result.Reply.BulkString()
	if result.Failed || !ok || value != "value" {
		t.Fatalf("string after rejected set commands = %q, %v, failed=%v", value, ok, result.Failed)
	}
}

func TestSetNoOpWritesDoNotChangeDirtyStateOrAOF(t *testing.T) {
	c := newStringCommandTestClient()
	log := &recordingAof{}
	c.aof = log

	if got := replyInteger(t, c.HandleCommand(redisCommand("SADD", "set", "member"))); got != 1 {
		t.Fatalf("initial SADD count = %d, want 1", got)
	}
	log.commands = nil
	dirtyBefore := c.server.dirty

	if got := replyInteger(t, c.HandleCommand(redisCommand("SADD", "set", "member", "member"))); got != 0 {
		t.Fatalf("duplicate SADD count = %d, want 0", got)
	}
	if got := replyInteger(t, c.HandleCommand(redisCommand("SREM", "set", "missing", "missing"))); got != 0 {
		t.Fatalf("missing SREM count = %d, want 0", got)
	}

	if c.server.dirty != dirtyBefore {
		t.Fatalf("dirty = %d, want unchanged value %d", c.server.dirty, dirtyBefore)
	}
	if len(log.commands) != 0 {
		t.Fatalf("AOF contains %d no-op writes, want 0", len(log.commands))
	}
}

func TestSetCommandsMapCorruptEncodingToInternalError(t *testing.T) {
	c := newStringCommandTestClient()
	c.db.setKey("broken", &RedisObject{
		typ:       SetObject,
		encoding:  EncodingSetMap,
		ptr:       hashMapPayload{},
		expiresAt: noExpiration,
	})

	result := c.HandleCommand(redisCommand("SCARD", "broken"))
	if !result.Failed {
		t.Fatal("SCARD unexpectedly accepted a corrupt set payload")
	}
	if result.Reply.str != "ERR internal server error" {
		t.Fatalf("SCARD error = %q, want internal error", result.Reply.str)
	}
}

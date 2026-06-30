package redis

import "testing"

func TestParseCommandUppercasesCommandAndKeepsArgs(t *testing.T) {
	gotCommand, gotArgs, ok := ParseCommand(Array([]Value{
		BulkString("get"),
		BulkString("name"),
	}))
	if !ok {
		t.Fatal("ParseCommand() ok = false, want true")
	}
	if gotCommand != "GET" {
		t.Fatalf("command = %q, want %q", gotCommand, "GET")
	}
	if len(gotArgs) != 1 {
		t.Fatalf("len(args) = %d, want 1", len(gotArgs))
	}

	gotKey, ok := gotArgs[0].BulkString()
	if !ok || gotKey != "name" {
		t.Fatalf("arg[0] = %q, %v; want %q, true", gotKey, ok, "name")
	}
}

func TestParseCommandRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		in   Value
	}{
		{"not array", BulkString("GET")},
		{"empty array", Array(nil)},
		{"command is not bulk string", Array([]Value{SimpleString("GET")})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, args, ok := ParseCommand(tt.in)
			if ok {
				t.Fatalf("ParseCommand() = %q, %#v, true; want invalid", command, args)
			}
		})
	}
}

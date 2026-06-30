package redis

import "testing"

func TestParseCommandUppercasesCommandAndKeepsArgs(t *testing.T) {
	gotCommand, gotArgs, err := ParseCommand(Array([]Value{
		BulkString("get"),
		BulkString("name"),
	}))
	if err != nil {
		t.Fatalf("ParseCommand() error = %v, want nil", err)
	}
	if gotCommand != "GET" {
		t.Fatalf("command = %q, want %q", gotCommand, "GET")
	}
	if len(gotArgs) != 1 {
		t.Fatalf("len(args) = %d, want 1", len(gotArgs))
	}

	gotKey := gotArgs[0]
	if gotKey != "name" {
		t.Fatalf("arg[0] = %q, want %q", gotKey, "name")
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
			command, args, err := ParseCommand(tt.in)
			if err == nil {
				t.Fatalf("ParseCommand() = %q, %#v, nil; want error", command, args)
			}
		})
	}
}

package redis

import "testing"

func TestParseCommandUppercasesCommandAndKeepsArgs(t *testing.T) {
	got, err := ParseCommand(Array([]Value{
		BulkString("get"),
		BulkString("name"),
	}))
	if err != nil {
		t.Fatalf("ParseCommand() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(tokens) = %d, want 2", len(got))
	}
	if got[0] != "GET" {
		t.Fatalf("command = %q, want %q", got[0], "GET")
	}

	gotKey := got[1]
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
			tokens, err := ParseCommand(tt.in)
			if err == nil {
				t.Fatalf("ParseCommand() = %#v, nil; want error", tokens)
			}
		})
	}
}

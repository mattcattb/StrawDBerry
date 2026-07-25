package redis

import (
	"errors"
	"testing"
)

func TestCommandTableResolveSubcommands(t *testing.T) {
	table := NewCommandTable()
	table.registerCommand("COMMAND", Command{
		subcommands: map[string]Command{
			"COUNT": {
				Arity:   0,
				Flags:   CmdRead,
				Group:   ManagementGroup,
				Handler: CommandCount,
			},
			"LIST": {
				Arity:   1,
				Flags:   CmdAdmin,
				Group:   ManagementGroup,
				Handler: CommandList,
			},
		},
	})

	t.Run("returns the selected leaf command", func(t *testing.T) {
		resolved, err := table.Resolve([]string{"COMMAND", "COUNT", "extra"})
		if err != nil {
			t.Fatalf("Resolve() error = %v, want nil", err)
		}
		got := resolved.Spec
		if got.Handler == nil {
			t.Fatal("resolved command has no handler")
		}
		if got.Arity != 0 || got.Group != ManagementGroup || got.Flags != CmdRead {
			t.Fatalf("resolved command = %#v, want COMMAND COUNT metadata", got)
		}
		if resolved.Name != "COMMAND|COUNT" {
			t.Fatalf("resolved name = %q, want COMMAND|COUNT", resolved.Name)
		}
		if len(resolved.Args) != 1 || resolved.Args[0] != "extra" {
			t.Fatalf("resolved args = %#v, want [extra]", resolved.Args)
		}
	})

	t.Run("matches the subcommand case-insensitively", func(t *testing.T) {
		resolved, err := table.Resolve([]string{"command", "list"})
		if err != nil {
			t.Fatalf("Resolve() error = %v, want nil", err)
		}
		got := resolved.Spec
		if got.Arity != 1 || got.Group != ManagementGroup || got.Flags != CmdAdmin {
			t.Fatalf("resolved command = %#v, want COMMAND LIST metadata", got)
		}
	})

	for _, tt := range []struct {
		name     string
		args     []string
		wantName string
		wantErr  error
	}{
		{"missing selector", []string{"COMMAND"}, "COMMAND", ErrWrongArgs},
		{"unknown selector", []string{"COMMAND", "UNKNOWN"}, "COMMAND|UNKNOWN", ErrUnknownCommand},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := table.Resolve(tt.args)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Resolve(%#v) error = %v, want %v", tt.args, err, tt.wantErr)
			}
			if resolved.Name != tt.wantName {
				t.Fatalf("Resolve(%#v) name = %q, want %q", tt.args, resolved.Name, tt.wantName)
			}
		})
	}
}

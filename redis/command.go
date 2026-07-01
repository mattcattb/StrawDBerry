package redis

import (
	"log"
	"strings"
)

type CommandGroup string

const (
	StringGroup     CommandGroup = "string"
	ConnGroup       CommandGroup = "connection"
	HashGroup       CommandGroup = "hash"
	SetGroup        CommandGroup = "set"
	ExpiryGroup     CommandGroup = "expiry"
	GenericGroup    CommandGroup = "generic"
	TxGroup         CommandGroup = "transaction"
	PubsubGroup     CommandGroup = "pubsub"
	ManagementGroup CommandGroup = "management"
)

type CommandFlags uint32

const (
	CmdRead CommandFlags = 1 << iota
	CmdWrite
	CmdAllowedInPubsub // function can be done in a pubsub
	CmdBlocking        // command translates to blokcing mode
	CmdAdmin           // Restricts the command from regular users (primarily internal usage)
	CmdNoScript        // cannot be ran in lua script
	CmdNoMulti         // command cannot be run inside a transaction
)

type Command struct {
	Name    string
	Arity   int // + is exact, negative is min
	Handler func(*Client, []string) CommandResult
	Group   CommandGroup
	Flags   CommandFlags
}

type CommandResult struct {
	Reply  Value
	Failed bool
}

func Result(reply Value) CommandResult {
	return CommandResult{Reply: reply}
}

func Failed(reply Value) CommandResult {
	return CommandResult{Reply: reply, Failed: true}
}

func ParseCommand(v Value) (string, []string, error) {
	values, ok := v.Array()
	if !ok || len(values) == 0 {
		return "", nil, ErrWrongArgs
	}

	commandName, ok := values[0].BulkString()
	if !ok {
		return "", nil, ErrWrongArgs
	}

	argVals, err := parseBulkStrCommand(values[1:])

	if err != nil {
		return "", nil, ErrWrongArgs
	}

	return strings.ToUpper(commandName), argVals, nil
}

func parseBulkStrCommand(args []Value) ([]string, error) {
	// parse all args into a string array

	stringArgs := make([]string, len(args))

	for i := 0; i < len(args); i += 1 {
		strVal, ok := args[i].BulkString()
		if !ok {
			return stringArgs, ErrWrongType
		}
		stringArgs[i] = strVal
	}

	return stringArgs, nil

}

func validateCommandArity(spec Command, command string, args []string) error {
	// Positive is exact, negative is minimum, zero means handler validates.
	arity := spec.Arity
	if arity == 0 {
		return nil
	}

	exact := !(arity < 0)

	argCount := len(args)
	if exact && arity != argCount {
		// exact arity check
		log.Printf("invalid exact arity: %v", argCount)

		return ErrWrongArgs
	}

	// less than
	if !exact && argCount < -1*arity {
		// minimum args is arity
		log.Printf("invalid arity LT: %v", argCount)

		return ErrWrongArgs
	}

	return nil
}

func validateCommandMode(client *Client, spec Command) error {
	//! needs shaping up

	if client.mode == ModeBlocking && spec.Flags.has(CmdBlocking) {
		// if mode blocking and not blocking flag
		return ErrInvalidState
	}

	switch client.mode {
	case ModeNormal:
		// hmmm... hmmm
		// pubsub requires being in a pubsub maybe?

		break

	case ModeBlocking:
		//! cannot do in blocking
		return ErrInvalidState

	case ModeTx:
		// in transaction mode, should just queue this?
		if spec.Flags.has(CmdNoMulti) {
			return ErrInvalidState
		}

		if spec.Flags.has(CmdBlocking) {
		}
		break

	case ModePubsub:
		//! in pubsub and command NOT ALLOWED in pubsub
		if !spec.Flags.has(CmdAllowedInPubsub) {
			return ErrInvalidState
		}
	}

	return nil
}

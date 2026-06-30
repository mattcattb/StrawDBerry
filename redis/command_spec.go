package redis

import "log"

type SpecHandler struct {
	commandMap map[string]CommandSpec
}

func NewSH() *SpecHandler {
	return &SpecHandler{commandMap: make(map[string]CommandSpec)}
}

type CommandGroup string

const (
	StringGroup  CommandGroup = "string"
	ConnGroup    CommandGroup = "connection"
	HashGroup    CommandGroup = "hash"
	SetGroup     CommandGroup = "set"
	ExpiryGroup  CommandGroup = "expiry"
	GenericGroup CommandGroup = "generic"
	TxGroup      CommandGroup = "transaction"
	PubsubGroup  CommandGroup = "pubsub"
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

type CommandSpec struct {
	name    string
	arity   int // + is exact, negative is min
	handler func(*Client, []string) CommandResult
	group   CommandGroup
	flags   CommandFlags
}

func (sh *SpecHandler) registerCommand(name string, spec CommandSpec) {
	spec.name = name
	sh.commandMap[name] = spec
}

func (f CommandFlags) has(flag CommandFlags) bool {
	return f&flag != 0
}

func (sh *SpecHandler) registerCommandSpecs(specMap map[string]CommandSpec) {

	for name, spec := range specMap {
		spec.name = name
		sh.commandMap[name] = spec
	}

}

func (sh *SpecHandler) getCommandSpec(command string) (CommandSpec, bool) {

	spec, exists := sh.commandMap[command]

	if !exists {
		return CommandSpec{}, false
	}

	return spec, true

}

func validateSpecArity(spec CommandSpec, command string, args []string) error {
	// Positive is exact, negative is minimum, zero means handler validates.
	arity := spec.arity
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

func validateCommandMode(client *Client, spec CommandSpec) error {
	//! needs shaping up

	if client.mode == ModeBlocking && spec.flags.has(CmdBlocking) {
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
		if spec.flags.has(CmdNoMulti) {
			return ErrInvalidState
		}

		if spec.flags.has(CmdBlocking) {
		}
		break

	case ModePubsub:
		//! in pubsub and command NOT ALLOWED in pubsub
		if !spec.flags.has(CmdAllowedInPubsub) {
			return ErrInvalidState
		}
	}

	return nil
}

func shouldQueuePipeline(c *Client, command string, spec CommandSpec) bool {
	if c.mode == ModeTx && spec.group != TxGroup && !spec.flags.has(CmdNoMulti) {
		return true
	}

	return false
}

func shouldAppendAof(spec CommandSpec, dirtyBefore, dirtyAfter uint64) bool {
	return spec.flags.has(CmdWrite) && dirtyAfter != dirtyBefore
}

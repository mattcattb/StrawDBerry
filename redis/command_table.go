package redis

type CommandTable struct {
	commands map[string]Command
	stats    CommandStat
}

func NewCommandTable() *CommandTable {
	return &CommandTable{commands: make(map[string]Command)}
}

func (sh *CommandTable) registerCommand(name string, spec Command) {
	spec.Name = name
	sh.commands[name] = spec
}

func (f CommandFlags) has(flag CommandFlags) bool {
	return f&flag != 0
}

func (sh *CommandTable) registerGroup(group CommandGroup, table map[string]Command) {
	for name, spec := range table {
		spec.Name = name
		spec.Group = group
		sh.commands[name] = spec
	}
}

func (sh *CommandTable) getCommand(command string) (Command, bool) {

	spec, exists := sh.commands[command]

	if !exists {
		return Command{}, false
	}

	return spec, true

}

func shouldQueuePipeline(c *Client, command string, spec Command) bool {
	if c.mode == ModeTx && spec.Group != TxGroup && !spec.Flags.has(CmdNoMulti) {
		return true
	}

	return false
}

func shouldAppendAof(spec Command, dirtyBefore, dirtyAfter uint64) bool {
	return spec.Flags.has(CmdWrite) && dirtyAfter != dirtyBefore
}

type CmdStatsManager struct {
	byCommand map[string]CommandStat

	totalAttempts   uint64
	unknownCommands uint64
	parseErrors     uint64
	errStats        CommandErrorStats
}

type CommandStat struct {
	calls uint64 //
	// usec           uint64 // the total CPU time consumed by these commands
	// usec_per_call  uint64 // the average CPU consumed per command execution
	rejected_calls uint64 // the number of rejected calls (malformed request)
	failed_calls   uint64 // the number of failed calls (internal failure)
}

type CommandErrorStats struct {
	WrongArity     uint64
	InvalidState   uint64
	WrongType      uint64
	InvalidInteger uint64
	Internal       uint64
	NotImplemented uint64
}

// we need to be able to determine from the command result the failure itself
// should be logged outside of the handler maybe
// -1 rejected, -2 failed

func (csm *CmdStatsManager) RecordCmdStat(command string, outcome int) {
	curStats, exists := csm.byCommand[command]

	if !exists {
		curStats = CommandStat{}
	}

	curStats.calls++

	if outcome == -1 {
		// rejected
		curStats.rejected_calls++

	} else if outcome == -2 {
		curStats.failed_calls++
	}

}

func (csm *CommandTable) recordCommandResult(cmdResult CommandResult) {

}

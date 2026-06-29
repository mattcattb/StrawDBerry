package redis

func registerSetCSpec(sh *SpecHandler) {

	setSpecs := map[string]CommandSpec{

		"SADD": {
			handler: SAdd,
			group:   SetGroup,
			flags:   CmdWrite,
		},
		"SCARD": {
			handler: SCard,
			group:   SetGroup,
			flags:   CmdRead,
		},
		"SREM": {
			handler: SRem,
			group:   SetGroup,
			flags:   CmdWrite,
		},
		"SISMEM": {
			handler: SIsMem,
			group:   SetGroup,
			flags:   CmdWrite,
		},
	}

	sh.registerCommandSpecs(setSpecs)

}

func SAdd(c *CommandContext, args []Value) CommandResult {
	return Failed(Error("ERR not implemented"))
}

func SCard(c *CommandContext, args []Value) CommandResult {
	return Failed(Error("ERR not implemented"))
}

func SRem(c *CommandContext, args []Value) CommandResult {
	return Failed(Error("ERR not implemented"))
}

func SIsMem(c *CommandContext, args []Value) CommandResult {
	return Failed(Error("ERR not implemented"))
}

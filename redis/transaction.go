package redis

func registerTransactionSpec(cs *CommandTable) {
	txSpec := map[string]Command{
		"EXEC": {
			Arity:   0,
			Handler: Exec,
			Group:   TxGroup,
		},
		"MULTI": {
			Arity:   0,
			Group:   TxGroup,
			Handler: Multi,
			Flags:   CmdNoMulti, // cannot call multi if in multi
		},
		"DISCARD": {
			Arity:   0,
			Group:   TxGroup,
			Handler: Discard,
		},
	}
	cs.registerCommandSpecs(txSpec)
}

func Exec(c *Client, args []string) CommandResult {

	if c.mode != ModeTx {
		return Failed(invalidStateError())
	}
	c.mode = ModeNormal
	replies := make([]Value, len(c.txQueue))

	for i, v := range c.txQueue {
		result := c.HandleCommand(v)
		replies[i] = result.Reply
	}

	c.txQueue = []Value{}
	return Result(Array(replies))
}

func Multi(c *Client, args []string) CommandResult {

	if c.mode == ModeTx {
		// cannot do multi in state here
		return Failed(invalidStateError())
	}
	// make mode multi mode?
	c.mode = ModeTx
	c.txQueue = make([]Value, 0)

	return Result(SimpleString("OK"))
}

// blocked commands if in multi just return the value themselves?

func Discard(c *Client, args []string) CommandResult {
	if c.mode != ModeTx {
		// cannot discard if not in multi
		return Failed(invalidStateError())
	}
	c.mode = ModeNormal
	c.txQueue = make([]Value, 0)

	return Result(SimpleString("OK"))
}

func Watch(c *Client, args []string) CommandResult {
	return Failed(Error("ERR not implemented"))
}

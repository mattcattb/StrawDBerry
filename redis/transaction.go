package redis

func registerTransactionSpec(cs *SpecHandler) {
	txSpec := map[string]CommandSpec{
		"EXEC": {
			arity:   0,
			handler: Exec,
			group:   TxGroup,
		},
		"MULTI": {
			arity:   0,
			group:   TxGroup,
			handler: Multi,
			flags:   CmdNoMulti, // cannot call multi if in multi
		},
		"DISCARD": {
			arity:   0,
			group:   TxGroup,
			handler: Discard,
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

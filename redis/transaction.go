package redis

func Exec(c *Client, args []string) CommandResult {

	if c.mode != ModeTx {
		return Failed(invalidStateError())
	}
	queued := c.txQueue
	c.txQueue = nil
	c.mode = ModeNormal
	if c.txFailed {
		c.txFailed = false
		return Failed(Error("EXECABORT Transaction discarded because of previous errors."))
	}

	replies := make([]Value, len(queued))
	for i, command := range queued {
		result := c.server.executeResolvedLocked(c, command.request, command.resolved)
		replies[i] = result.Reply
	}

	return Result(Array(replies))
}

func Multi(c *Client, args []string) CommandResult {

	if c.mode == ModeTx {
		// cannot do multi in state here
		return Failed(invalidStateError())
	}
	// make mode multi mode?
	c.mode = ModeTx
	c.txQueue = make([]QueuedCommand, 0)
	c.txFailed = false

	return Result(SimpleString("OK"))
}

// blocked commands if in multi just return the value themselves?

func Discard(c *Client, args []string) CommandResult {
	if c.mode != ModeTx {
		// cannot discard if not in multi
		return Failed(invalidStateError())
	}
	c.mode = ModeNormal
	c.txQueue = nil
	c.txFailed = false

	return Result(SimpleString("OK"))
}

func Watch(c *Client, args []string) CommandResult {
	return Failed(Error("ERR not implemented"))
}

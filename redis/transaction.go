package redis

type Transacion struct {
	inMulti bool
	queued  []Value
	dirty   bool
}

func (tx *Transacion) clearMulti() []Value {
	tx.inMulti = false
	vals := make([]Value, len(tx.queued))
	copy(vals, tx.queued)
	tx.queued = make([]Value, 0)
	return vals
}

func (tx *Transacion) initMulti() {
	tx.inMulti = true
	tx.queued = make([]Value, 0)
}

func (tx *Transacion) queueVal(val Value) {
	tx.queued = append(tx.queued, val)
}

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

func Exec(c *Client, args []Value) CommandResult {

	if c.mode != ModeTx {
		return Failed(invalidStateError())
	}
	queuedCommands := c.tx.clearMulti()
	replies := make([]Value, len(queuedCommands))

	for i := 0; i < len(c.tx.queued); i += 1 {
		result := c.HandleCommand(c.tx.queued[i])
		replies[i] = result.Reply
	}

	c.tx.inMulti = false
	c.tx.queued = []Value{}
	return Result(Array(replies))
}

func Multi(c *Client, args []Value) CommandResult {

	if c.mode == ModeTx {
		// cannot do multi in state here
		return Failed(invalidStateError())
	}
	// make mode multi mode?
	c.tx.initMulti()

	return Result(SimpleString("OK"))
}

// blocked commands if in multi just return the value themselves?

func Discard(c *Client, args []Value) CommandResult {
	if c.mode != ModeTx {
		return Failed(invalidStateError())
	}
	c.tx.clearMulti()

	return Result(SimpleString("OK"))
}

func Watch(c *Client, args []Value) CommandResult {
	return Failed(Error("ERR not implemented"))
}

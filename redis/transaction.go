package redis

type Transacion struct {
	inMulti bool
	queued  []Value
	dirty   bool
}

func (tx *Transacion) clearMulti() []Value {
	tx.inMulti = false
	vals := tx.queued
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
			name:    "EXEC",
			minArgs: 1,
			maxArgs: 1,
			handler: Exec,
			group:   TxGroup,
		},
		"MULTI": {
			name:    "MULTI",
			minArgs: 1,
			maxArgs: 1,
			group:   TxGroup,
			handler: Multi,
		},
		"Discard": {
			name:    "DISCARD",
			minArgs: 1,
			maxArgs: 1,
		},
	}
	cs.registerCommandSpecs(txSpec)
}

func Exec(c *Client, args []Value) CommandResult {

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

	// make mode multi mode?
	c.tx.initMulti()

	return Result(SimpleString("OK"))
}

func Discard(c *Client, args []Value) CommandResult {
	c.tx.clearMulti()

	return Result(SimpleString("OK"))
}

func Watch(c *Client, args []Value) CommandResult {
	return Failed(Error("ERR not implemented"))
}

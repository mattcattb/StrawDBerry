package redis

func registerConnectionCommands(ch *SpecHandler) {
	ch.registerCommandSpecs(map[string]CommandSpec{
		"PING": {
			arity:   0,
			handler: Ping,
			group:   ConnGroup,
		}, "ECHO": {
			arity:   1,
			handler: Echo,
			group:   ConnGroup,
		},
	})
}

func Ping(c *CommandContext, args []Value) CommandResult {
	if len(args) > 1 {
		return Failed(wrongArgs("PING"))
	}

	if len(args) == 0 {
		return Result(SimpleString("pong"))
	}

	argMsg, _ := args[0].BulkString()

	return Result(BulkString(argMsg))
}

func Echo(c *CommandContext, args []Value) CommandResult {
	argMsg, ok := args[0].BulkString()
	if !ok {
		return Failed(syntaxError())
	}

	return Result(BulkString(argMsg))
}

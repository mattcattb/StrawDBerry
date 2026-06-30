package redis

func registerConnectionCommands(ch *SpecHandler) {
	ch.registerCommandSpecs(map[string]CommandSpec{
		"PING": {
			arity:   0,
			handler: Ping,
			group:   ConnGroup,
			flags:   CmdAllowedInPubsub,
		}, "ECHO": {
			arity:   1,
			handler: Echo,
			group:   ConnGroup,
		},
	})
}

func Ping(c *Client, args []string) CommandResult {

	if len(args) == 0 {
		return Result(SimpleString("PONG"))
	}

	return Result(BulkString(args[0]))
}

func Echo(c *Client, args []string) CommandResult {
	return Result(BulkString(args[0]))
}

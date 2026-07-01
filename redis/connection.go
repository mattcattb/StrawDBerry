package redis

func registerConnectionCommands(ch *CommandTable) {
	ch.registerCommandSpecs(map[string]Command{
		"PING": {
			Arity:   0,
			Handler: Ping,
			Group:   ConnGroup,
			Flags:   CmdAllowedInPubsub,
		}, "ECHO": {
			Arity:   1,
			Handler: Echo,
			Group:   ConnGroup,
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

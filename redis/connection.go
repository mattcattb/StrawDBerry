package redis

func Ping(c *Client, args []string) CommandResult {

	if len(args) == 0 {
		return Result(SimpleString("PONG"))
	}

	return Result(BulkString(args[0]))
}

func Echo(c *Client, args []string) CommandResult {
	return Result(BulkString(args[0]))
}

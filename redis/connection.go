package redis

func Ping(c *Client, args []string) CommandResult {

	if len(args) == 0 {
		return Result(SimpleString("PONG"))
	}
	if len(args) > 1 {
		return Failed(wrongArgs("PING"))
	}

	return Result(BulkString(args[0]))
}

func Echo(c *Client, args []string) CommandResult {
	return Result(BulkString(args[0]))
}

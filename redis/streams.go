package redis

type SMessage struct {
	Id   string
	data map[string]string
}

type Stream struct {
}

func XAdd(c *Client, args []Value) CommandResult

func XRange(c *Client, args []Value)

func XRead(c *Client, args []Value)

// COUNT and BLOCK

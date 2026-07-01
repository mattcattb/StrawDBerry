package redis

// TODO uhhh prefixing behavior (COMMAND _)

type infoParams struct {
}

func Info(c *Client, args []string) CommandResult {

	/*
		a map of info fields, one field per line in the form of <field>:<value> where the value can be a comma
		separated map like <key>=<val>. Also contains section header lines starting with # and blank lines.
		 Lines can contain a section name (starting with a # character) or a property. All the properties are
		 in the form of field:value terminated by \r\n.
	*/
	// TODO!!!
	// sfieldArray := make([]Value, 0)

	// serverStats := c.server.

	return Failed(Error("ERR not implemented"))
}

func CommandList(c *Client, args []string) CommandResult {
	// Array reply: a list of command names.

	respArr := make([]Value, 0)

	for name, _ := range c.server.sh.commands {
		respArr = append(respArr, BulkString(name))
	}

	return Result(Array(respArr))
}

/*
func CommandInfo(c *Client, args []string) CommandResult {

}

func ConfigGet(c *Client, args []string) CommandResult {

}*/

func CommandCount(c *Client, args []string) CommandResult {
	n := len(c.server.sh.commands)
	return Result(Integer(n))
}

func DbSize(c *Client, args []string) CommandResult {
	// int keys in database
	snapshot := c.db.StatsSnapshot()

	return Result(Integer(int(snapshot.keys)))
}

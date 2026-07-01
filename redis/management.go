package redis

type Management struct{}

// TODO uhhh prefixing behavior (COMMAND _)

type infoParams struct {
}

func Info(c *Client, args []string) CommandResult {

	/*

	 */

	return Failed(Error("ERR not implemented"))
}

/*
func CommandList(c *Client, args []string) CommandResult {

}

func CommandInfo(c *Client, args []string) CommandResult {

}

func CommandCount(c *Client, args []string) CommandResult {

}

func ConfigGet(c *Client, args []string) CommandResult {

}

func DbSize(c *Client, args []string) CommandResult {
	// int keys in database
}
*/

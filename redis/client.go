package redis

import "net"

type ClientMode uint8

const (
	ModeNormal ClientMode = iota
	ModeTx
	ModeBlocking
	ModePubsub
)

type PersistanceLog interface {
	Append(v Value) error
}

type Client struct {
	out    chan Value
	server *Server
	conn   net.Conn
	reader *Resp
	writer *Writer
	tx     *Transacion
	db     *RedisDb
	aof    PersistanceLog
	mode   ClientMode
}

func NewClient(conn net.Conn, server *Server) *Client {
	return &Client{server: server, conn: conn, reader: NewResp(conn), writer: NewWriter(conn), tx: &Transacion{}, db: server.exec.db, aof: server.aof, mode: ModeNormal}
}

func (c *Client) ListenToPublishing() {
	go func() {
		for outVal := range c.out {
			c.writer.Write(outVal)
		}
	}()
}

func (c *Client) CloseListener() {
	close(c.out)
}

/*
handle command VS replay AOF

BLOCKING COMMAND FLAG: switch to blocking mode
*/
func (c *Client) HandleCommand(req Value) CommandResult {

	/*
		Handle Command takes a req val, parses the command and arguments, looks up command spec,
		validates arg size, validates flags for mode, executes command,
		handles AOF behaviors for command result (writing command to log if mutation occured)
	*/

	command, args, valid := ParseCommand(req)
	if !valid {
		return Failed(syntaxError())
	}
	spec, ok := c.server.sh.getCommandSpec(command)
	if !ok {
		return Failed(unknownCommand(command))
	}

	// validate args
	err := validateCommandArity(spec, command, args)

	if err != nil {
		return Failed(wrongArgs(command))
	}

	// validate mode, make sure command can be done in current client mode
	err = validateCommandMode(c, spec)

	if err != nil {
		return Failed(invalidStateError())
	}

	// queue transaction depends on the mode
	// run command

	if shouldQueuePipeline(c, command, spec) {
		// queue this only
		c.tx.queueVal(req)
		return Result(SimpleString("OK"))
	}

	dirtyBefore := c.server.dirty

	result := spec.handler(c, args)
	// aof write
	// ? make this jsut so spec is a write one
	if shouldAppendAof(spec, dirtyBefore, c.server.dirty) {
		c.aof.Append(req)
	}
	return result
}

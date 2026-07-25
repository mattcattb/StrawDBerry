package redis

import (
	"net"
)

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
	out     chan Value
	server  *Server
	conn    net.Conn
	reader  *Resp
	writer  *Writer
	txQueue []Value
	db      *RedisDb
	aof     PersistanceLog
	mode    ClientMode
}

func NewClient(conn net.Conn, server *Server) *Client {
	return &Client{server: server, conn: conn, reader: NewResp(conn), writer: NewWriter(conn), db: server.db, aof: server.aof, mode: ModeNormal}
}

func (c *Client) ListenToPublishing(bufSize int) {

	c.out = make(chan Value, bufSize)

	go func() {
		// add check if blocking, do not do anything
		for outVal := range c.out {
			c.writer.Write(outVal)
		}
	}()

}

func (c *Client) close() {

	// close pubsub and disconnect client
	close(c.out)
	c.conn.Close()

}

func (c *Client) Send(val Value) bool {
	c.out <- val

	return true
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

	tokens, err := ParseCommand(req)
	if err != nil {
		return Failed(syntaxError())
	}

	rCmd, err := c.server.sh.Resolve(tokens)
	if err != nil {
		return Failed(Error(err.Error()))
	}

	if err = validateCommandArity(rCmd.Spec, rCmd.Name, rCmd.Args); err != nil {
		return Failed(wrongArgs(rCmd.Name))
	}
	// validate mode, make sure command can be done in current client mode
	if err = validateCommandMode(c, rCmd.Spec); err != nil {
		return Failed(invalidStateError())
	}

	if shouldQueuePipeline(c, rCmd.Name, rCmd.Spec) {
		// queue this only
		c.txQueue = append(c.txQueue, req)
		return Result(SimpleString("OK"))
	}

	dirtyBefore := c.server.dirty

	result := rCmd.Spec.Handler(c, rCmd.Args)
	// aof write
	// ? make this jsut so spec is a write one
	if shouldAppendAof(rCmd.Spec, dirtyBefore, c.server.dirty) {
		c.aof.Append(req)
	}
	return result
}

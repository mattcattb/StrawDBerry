package redis

import (
	"net"
	"sync"
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

type QueuedCommand struct {
	request  Value
	resolved ResolvedCommand
}

type Client struct {
	server    *Server
	conn      net.Conn
	outbox    chan Value
	done      chan struct{}
	closeOnce sync.Once
	txFailed  bool
	reader    *Resp
	writer    *Writer
	txQueue   []QueuedCommand
	db        *RedisDb
	aof       PersistanceLog
	mode      ClientMode
}

func NewClient(conn net.Conn, server *Server) *Client {
	return &Client{
		server: server,
		conn:   conn,
		reader: NewResp(conn),
		writer: NewWriter(conn),
		db:     server.db,
		aof:    server.aof,
		mode:   ModeNormal,
		outbox: make(chan Value, 256),
		done:   make(chan struct{}),
	}
}

func (c *Client) startWriter() {
	go func() {
		for {
			select {
			case val := <-c.outbox:
				if err := c.writer.Write(val); err != nil {
					c.server.RequestDisconnect(c)
					return
				}
			case <-c.done:
				for {
					select {
					case val := <-c.outbox:
						if err := c.writer.Write(val); err != nil {
							c.server.RequestDisconnect(c)
							return
						}
					default:
						if c.conn != nil {
							_ = c.conn.Close()
						}
						return
					}
				}
			}
		}
	}()
}

func (c *Client) enqueue(val Value) bool {
	select {
	case <-c.done:
		return false
	default:
	}

	select {
	case c.outbox <- val:
		return true
	default:
		return false
	}
}

func (c *Client) finish() {
	c.closeOnce.Do(func() {
		close(c.done)
	})
}

func (c *Client) stop() {
	c.finish()
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

/*
handle command VS replay AOF

BLOCKING COMMAND FLAG: switch to blocking mode
*/

func (c *Client) HandleCommand(req Value) CommandResult {
	return c.server.execute(c, req, false)
}

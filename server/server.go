package server

import (
	"go-redis/resp"
	"net"
)

type Client struct {
	conn    net.Conn
	reader  *resp.Resp
	writer  *resp.Writer
	inMulti bool
	queued  []resp.Value
	dirty   bool
}

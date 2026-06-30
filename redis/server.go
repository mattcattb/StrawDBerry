package redis

import (
	"net"
)

type Server struct {
	clients []*Client
	exec    *CommandContext
	pubsub  *PubSub
	aof     *Aof
	sh      *SpecHandler
	dirty   uint64
}

func NewServer(exec *CommandContext, aof *Aof, sh *SpecHandler) *Server {

	return &Server{exec: exec, aof: aof, sh: sh}

}

func (s *Server) RegisterAllCommandHandlers() {

	sh := s.sh
	registerConnectionCommands(sh)
	registerGenericCommands(sh)
	registerTransactionSpec(sh)
	registerPubsubCommands(sh)
	registerTHashCommandSpec(sh)
	registerTstringCommands(sh)
	registerSetCSpec(sh)
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	client := &Client{server: s, conn: conn, reader: NewResp(conn), writer: NewWriter(conn), tx: &Transacion{}, db: s.exec.db, aof: s.aof}

	s.clients = append(s.clients, client)

	client.ListenToPublishing()
	for {
		req, err := client.reader.Read()
		if err != nil {
			client.CloseListener()
			return
		}

		result := client.HandleCommand(req)

		if err := client.writer.Write(result.Reply); err != nil {
			client.CloseListener()
			return
		}

	}
}

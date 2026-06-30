package redis

import (
	"net"
)

type Server struct {
	clients []*Client
	ps      *PubSubServer
	aof     *Aof
	db      *RedisDb
	sh      *SpecHandler
	dirty   uint64
}

func NewServer(db *RedisDb, aof *Aof, sh *SpecHandler) *Server {

	ps := &PubSubServer{}

	return &Server{aof: aof, sh: sh, ps: ps, db: db}

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

	client := &Client{server: s, conn: conn, reader: NewResp(conn), writer: NewWriter(conn), tx: &Transacion{}, db: s.db, aof: s.aof}

	s.clients = append(s.clients, client)

	client.ListenToPublishing(0)

	defer client.Disconnect()

	for {
		req, err := client.reader.Read()
		if err != nil {
			return
		}

		result := client.HandleCommand(req)

		if err := client.writer.Write(result.Reply); err != nil {
			return
		}

	}
}

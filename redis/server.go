package redis

import (
	"log"
	"net"
)

type SConfig struct {
	maxmemory int64
}

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
	defer func() {
		log.Printf("closing client %s", conn.RemoteAddr())
		conn.Close()
	}()

	client := &Client{
		server: s,
		conn:   conn,
		reader: NewResp(conn),
		writer: NewWriter(conn),
		db:     s.db,
		aof:    s.aof}

	s.clients = append(s.clients, client)

	client.ListenToPublishing(0)

	defer client.Disconnect()

	for {
		req, err := client.reader.Read()
		if err != nil {
			log.Printf(`read error %s: %#v`, conn.RemoteAddr(), req)
			return
		}

		log.Printf("read request %q", string(req.Marshal()))

		result := client.HandleCommand(req)

		if err := client.writer.Write(result.Reply); err != nil {
			log.Printf("write error to %s: %#v", conn.RemoteAddr(), err)

			return
		}

		log.Printf("wrote reply to %s", conn.RemoteAddr())

	}
}

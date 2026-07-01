package redis

import (
	"fmt"
	"log"
	"net"
	"time"
)

type SConfig struct {
	maxmemory  uint64
	maxclients uint64
	Addr       string
	tcpPort    uint64
}

type ServerStats struct {
	startedAt                time.Time
	connectedClients         uint64
	totalConnectionsAccepted uint64
	totalConnectionsClosed   uint64
	rejectedConnections      uint64
	readErrors               uint64
	writeErrors              uint64
	peakConnectedClients     uint64
}

type Server struct {
	config  SConfig
	clients map[*Client]struct{}
	ps      *PubSubServer
	aof     *Aof
	db      *RedisDb
	sh      *CommandTable
	dirty   uint64
	sStats  *ServerStats
}

func NewServer(db *RedisDb, aof *Aof, sh *CommandTable) *Server {

	ps := NewPubsubServer()

	return &Server{aof: aof, sh: sh, ps: ps, db: db,
		clients: make(map[*Client]struct{}, 0),
		sStats: &ServerStats{
			startedAt: time.Now(),
		}}

}

func (s *Server) RegisterCMDTable() {
	s.sh.registerGroup(StringGroup, StringCmdTable)
	s.sh.registerGroup(HashGroup, HashCmdTable)
	s.sh.registerGroup(SetGroup, SetCmdTable)
	s.sh.registerGroup(ManagementGroup, ManagementCMDTable)
	s.sh.registerGroup(GenericGroup, GenericCMDTable)
	s.sh.registerGroup(ConnGroup, ConnectionCmdTable)
	s.sh.registerGroup(PubsubGroup, PubsubCmdTable)
	s.sh.registerGroup(TxGroup, TxCmdTable)
}

/*
The client socket is put in the non-blocking state since Redis uses multiplexing and non-blocking I/O.
The TCP_NODELAY option is set in order to ensure that there are no delays to the connection.
A readable file event is created so that Redis is able to collect the client queries as soon as new data is available to read on the socket
*/

// After the client is initialized, Redis checks if it is already at the limit configured for the number of simultaneous clients (configured using the maxclients configuration directive, see the next section of this document for further information).

func (s *Server) HandleConnection(conn net.Conn) {

	client := NewClient(conn, s)
	client.ListenToPublishing(0)
	s.AddClient(client)

	defer s.DisconnectClient(client)

	for {
		req, err := client.reader.Read()
		if err != nil {
			log.Printf(`read error %s: %#v`, conn.RemoteAddr(), req)
			s.sStats.readErrors++
			continue
		}

		log.Printf("read request %q", string(req.Marshal()))

		result := client.HandleCommand(req)

		if err := client.writer.Write(result.Reply); err != nil {
			log.Printf("write error to %s: %#v", conn.RemoteAddr(), err)
			s.sStats.writeErrors++
			continue
		}

		log.Printf("wrote reply to %s", conn.RemoteAddr())
	}
}

func (s *Server) AddClient(c *Client) {
	s.clients[c] = struct{}{}
	s.sStats.connectedClients++

}

func (s *Server) DisconnectClient(c *Client) {
	c.close()
	s.ps.disconnClient(c)
	delete(s.clients, c)
	s.sStats.connectedClients--

	// s.clients
}

func (s *Server) SocketListenLoop(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			s.sStats.rejectedConnections++
			continue
		}

		go s.HandleConnection(conn)
	}
}

func (s *Server) StatsSnapshot() {

}

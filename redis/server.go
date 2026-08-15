package redis

import (
	"fmt"
	"log"
	"net"
	"sync"
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
	execMu    sync.Mutex // exec behaviors
	clientsMu sync.Mutex // connection counter + client map
	config    SConfig
	clients   map[*Client]struct{}
	ps        *PubSubServer
	aof       *Aof
	db        *RedisDb
	sh        *CommandTable
	dirty     uint64
	sStats    *ServerStats
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
	client.startWriter()
	s.AddClient(client)

	defer s.FinishClient(client)

	for {
		req, err := client.reader.Read()
		if err != nil {
			log.Printf(`read error %s: %#v`, conn.RemoteAddr(), req)
			s.sStats.readErrors++
			return
		}

		log.Printf("read request %q", string(req.Marshal()))

		s.execute(client, req, true)
	}
}

func (s *Server) AddClient(c *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	s.clients[c] = struct{}{}
	s.sStats.connectedClients++

}

func (s *Server) DisconnectClient(c *Client) {
	s.execMu.Lock()
	defer s.execMu.Unlock()
	s.disconnectClientLocked(c, true)
}

func (s *Server) RequestDisconnect(c *Client) {
	s.DisconnectClient(c)
}

func (s *Server) FinishClient(c *Client) {
	s.execMu.Lock()
	defer s.execMu.Unlock()
	s.disconnectClientLocked(c, false)
}

func (s *Server) disconnectClientLocked(c *Client, force bool) {
	s.ps.disconnClient(c)
	if force {
		c.stop()
	} else {
		c.finish()
	}

	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	if _, exists := s.clients[c]; !exists {
		return
	}
	delete(s.clients, c)
	s.sStats.connectedClients--
}

func (s *Server) enqueueLocked(c *Client, value Value) {
	if !c.enqueue(value) {
		s.disconnectClientLocked(c, true)
	}
}

func (s *Server) prepare(request Value) (ResolvedCommand, CommandResult) {
	tokens, err := ParseCommand(request)
	if err != nil {
		return ResolvedCommand{}, Failed(syntaxError())
	}

	command, err := s.sh.Resolve(tokens)
	if err != nil {
		return ResolvedCommand{}, Failed(Error(err.Error()))
	}

	if err := validateCommandArity(command.Spec, command.Name, command.Args); err != nil {
		return ResolvedCommand{}, Failed(wrongArgs(command.Name))
	}

	return command, CommandResult{}
}

func (s *Server) execute(c *Client, request Value, deliver bool) CommandResult {
	command, preparation := s.prepare(request)

	s.execMu.Lock()
	defer s.execMu.Unlock()

	result := preparation
	if !result.Failed {
		result = s.executeResolvedLocked(c, request, command)
	} else if c.mode == ModeTx {
		c.txFailed = true
	}

	if deliver {
		s.enqueueLocked(c, result.Reply)
	}

	return result
}

func (s *Server) executeResolvedLocked(c *Client, request Value, command ResolvedCommand) CommandResult {
	if err := validateCommandMode(c, command.Spec); err != nil {
		if c.mode == ModeTx && command.Spec.Group != TxGroup {
			c.txFailed = true
		}
		return Failed(invalidStateError())
	}

	if shouldQueuePipeline(c, command.Name, command.Spec) {
		c.txQueue = append(c.txQueue, QueuedCommand{request: request, resolved: command})
		return Result(SimpleString("QUEUED"))
	}

	dirtyBefore := s.dirty
	result := command.Spec.Handler(c, command.Args)
	if shouldAppendAof(command.Spec, dirtyBefore, s.dirty) && s.aof != nil {
		if err := s.aof.Append(request); err != nil {
			log.Printf("AOF append failed: %v", err)
		}
	}
	return result
}
func (s *Server) LoadAOF() error {
	if s.aof == nil {
		return nil
	}

	// set loading as true here

	defer func() {
		// s.loading = false
	}()

	client := NewClient(nil, s)

	return s.aof.Replay(client, func(req Value) error {
		r := s.execute(client, req, false)

		if r.Failed {
			return fmt.Errorf("AOF command failed")
		}

		return nil
	})
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

package redis

import (
	"sync"
)

type PubSubServer struct {
	channels  map[string]map[*Client]struct{}
	connCount map[*Client]int
	mu        sync.Mutex
	stats     PubsubStats
}
type PubsubStats struct {
	pubsub_channels uint64
	pubsub_clients  uint64
}

// client pointer vs... hmmm

func NewPubsubServer() *PubSubServer {
	return &PubSubServer{
		channels:  make(map[string]map[*Client]struct{}),
		connCount: make(map[*Client]int),
		mu:        sync.Mutex{},
	}

}

func (ps *PubSubServer) disconnClient(client *Client) int {
	ps.mu.Lock()
	channels := make([]string, 0, len(ps.channels))
	for channel := range ps.channels {
		channels = append(channels, channel)
	}
	ps.mu.Unlock()

	c := 0
	for _, channel := range channels {
		c += ps.unsub(channel, client)
	}

	return c
}

func (ps *PubSubServer) sub(client *Client, channel string) int {
	// For instance, to subscribe to channels "channel11" and "ch:00" the client issues a SUBSCRIBE providing the names of the channels:
	// SUBSCRIBE channel11 ch:00

	// number of channels we are currently subscribed to

	ps.mu.Lock()
	defer ps.mu.Unlock()

	chanMap, exists := ps.channels[channel]

	if !exists {
		chanMap = make(map[*Client]struct{})
		ps.channels[channel] = chanMap
	}

	_, alreadySubscribed := chanMap[client]

	conCount, conCountExists := ps.connCount[client]

	if !conCountExists {
		conCount = 0
	}

	if !alreadySubscribed {
		chanMap[client] = struct{}{}
		conCount = conCount + 1
		ps.connCount[client] = conCount
	}

	return conCount

}

func (ps *PubSubServer) getSubed(channel string) (conClients []*Client) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	clients, chExists := ps.channels[channel]

	if !chExists {
		return conClients
	}

	for client := range clients {
		conClients = append(conClients, client)
	}
	return conClients
}

func (ps *PubSubServer) unsub(channel string, client *Client) int {

	// means that we successfully unsubscribed from the channel given as second element in the reply.
	//  The third argument represents the number of channels we are currently subscribed to. When the last argument is zero,
	//  we are no longer subscribed to any channel, and the client can issue any kind of Redis command as we are outside the Pub/Sub state

	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch, exists := ps.channels[channel]

	if !exists {
		return ps.connCount[client]
	}

	_, isSubscribed := ch[client]

	connectedChannels, _ := ps.connCount[client]

	if !isSubscribed {
		return connectedChannels
	}

	if connectedChannels != 0 {
		delete(ch, client)
		connectedChannels -= 1
		ps.connCount[client] = connectedChannels
		if connectedChannels == 0 {
			delete(ps.connCount, client)
		}
	}
	if len(ch) == 0 {
		delete(ps.channels, channel)
	}

	return connectedChannels
}

func Subscribe(c *Client, args []string) CommandResult {
	channel := args[0]
	c.mode = ModePubsub
	n := c.server.ps.sub(c, channel)
	return Result(Array([]Value{BulkString("subscribe"), BulkString(channel), Integer(n)}))
}

func Unsubscribe(c *Client, args []string) CommandResult {
	unsubChannel := args[0]

	n := c.server.ps.unsub(unsubChannel, c)
	if n == 0 {
		// no longer normal mode
		c.mode = ModeNormal
	}

	return Result(Array([]Value{BulkString("unsubscribe"), BulkString(unsubChannel), Integer(n)}))
}

func Publish(c *Client, args []string) CommandResult {
	// number subscribers recieved message
	channel, message := args[0], args[1]

	clients := c.server.ps.getSubed(channel)

	for _, targetClient := range clients {
		sendMessage := Array([]Value{BulkString("message"), BulkString(channel), BulkString(message)})
		c.server.enqueueLocked(targetClient, sendMessage)
	}

	return Result(Integer(len(clients)))
}

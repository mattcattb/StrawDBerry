package redis

import (
	"sync"
)

type PubSubServer struct {
	channels  map[string]map[*Client]struct{}
	connCount map[*Client]int
	mu        *sync.Mutex
}

// client pointer vs... hmmm
func registerPubsubCommands(sh *SpecHandler) {
	pubsubRegistry := map[string]CommandSpec{
		"SUBSCRIBE": {
			handler: Subscribe,
			arity:   1,
			group:   PubsubGroup,
			flags:   CmdAllowedInPubsub & CmdNoMulti,
		}, "UNSUBSCRIBE": {
			handler: Unsubscribe,
			arity:   1,
			group:   PubsubGroup,
			flags:   CmdAllowedInPubsub & CmdNoMulti,
		},
		"PUBLISH": {
			handler: Publish,
			arity:   2,
			group:   PubsubGroup,
		},
	}

	sh.registerCommandSpecs(pubsubRegistry)
}

func (ps *PubSubServer) disconnClient(client *Client) int {
	c := 0
	for ch, _ := range ps.channels {
		c += ps.unsub(ch, client)
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
		return 0
	}

	_, isSubscribed := ch[client]

	connectedChannels, _ := ps.connCount[client]

	if !isSubscribed {
		return connectedChannels
	}

	if connectedChannels != 0 {
		// connected exists
		connectedChannels -= 1
		ps.connCount[client] -= 1
		if connectedChannels == 0 {
			// clear it, now its 0
			delete(ps.connCount, client)
		}
	}

	return connectedChannels
}

func Subscribe(c *Client, args []Value) CommandResult {
	channels, err := parseBulkRespStringCommands(args)
	if err != nil {
		return Failed(wrongArgs("SUBSCRIBE"))
	}
	channel := channels[0]
	c.mode = ModePubsub
	n := c.server.ps.sub(c, channel)
	return Result(Array([]Value{BulkString("subscribe"), BulkString(channel), Integer(n)}))
}

func Unsubscribe(c *Client, args []Value) CommandResult {
	channels, err := parseBulkRespStringCommands(args)

	if err != nil {
		return Failed(wrongArgs("UNSUBSCRIBE"))
	}

	unsubChannel := channels[0]

	n := c.server.ps.unsub(unsubChannel, c)
	if n == 0 {
		// no longer normal mode
		c.mode = ModeNormal
	}

	return Result(Array([]Value{BulkString("unsubscribe"), BulkString(unsubChannel), Integer(n)}))
}

func Publish(c *Client, args []Value) CommandResult {

	// number subscribers recieved message
	vals, err := parseBulkRespStringCommands(args)

	if err != nil {
		return Failed(wrongArgs("PUBLISH"))
	}
	channel, message := vals[0], vals[1]

	clients := c.server.ps.getSubed(channel)

	for _, targetClient := range clients {
		sendMessage := Array([]Value{BulkString("message"), BulkString(channel), BulkString(message)})
		targetClient.Send(sendMessage)
	}

	return Result(Integer(len(clients)))
}

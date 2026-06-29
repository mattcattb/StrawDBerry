package redis

import "sync"

type PubSub struct {
	channels  map[string]map[*Client]struct{}
	connCount map[*Client]int
	mu        *sync.Mutex
}

func registerPubsubCommands(sh *SpecHandler) {
	pubsubRegistry := map[string]CommandSpec{
		"SUBSCRIBE": {
			name:    "SUBSCRIBE",
			minArgs: 1,
			handler: Subscribe,
			group:   PubsubGroup,
			flags:   CmdPubSubOnly,
		}, "UNSUBSCRIBE": {
			handler: Unsubscribe,
			group:   PubsubGroup,
			flags:   CmdPubSubOnly,
		},
	}

	sh.registerCommandSpecs(pubsubRegistry)
}

func (ps *PubSub) sub(client *Client, channel string) int {
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

func (ps *PubSub) getSubed(channel string) (conClients []*Client) {
	clients, chExists := ps.channels[channel]

	if !chExists {
		return conClients
	}

	for client := range clients {
		conClients = append(conClients, client)
	}
	return conClients
}

func (ps *PubSub) unsub(channel string, client *Client) int {

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
	return Failed(Error("ERR not implemented"))
}

func Unsubscribe(c *Client, args []Value) CommandResult {
	return Failed(Error("ERR not implemented"))
}

func Publish(c *Client, args []Value) CommandResult {
	return Failed(Error("ERR not implemented"))
}

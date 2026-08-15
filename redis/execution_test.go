package redis

import (
	"net"
	"testing"
	"time"
)

func newExecutionTestServer() *Server {
	db := NewDb()
	server := NewServer(db, nil, NewCommandTable())
	server.RegisterCMDTable()
	return server
}

func newExecutionTestClient(server *Server) *Client {
	client := NewClient(nil, server)
	client.aof = &DummyAofLog{}
	return client
}

func TestClientOutboxWritesRepliesInOrder(t *testing.T) {
	serverConn, peerConn := net.Pipe()
	defer peerConn.Close()

	server := newExecutionTestServer()
	client := NewClient(serverConn, server)
	defer client.stop()
	client.startWriter()

	if !client.enqueue(SimpleString("first")) || !client.enqueue(SimpleString("second")) {
		t.Fatal("could not enqueue replies")
	}

	if err := peerConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	reader := NewResp(peerConn)
	for _, want := range []string{"first", "second"} {
		value, err := reader.Read()
		if err != nil {
			t.Fatalf("read reply: %v", err)
		}
		got, ok := value.SimpleString()
		if !ok || got != want {
			t.Fatalf("reply = %q, %v; want %q, true", got, ok, want)
		}
	}
}

func TestConnectionRepliesThroughClientOutbox(t *testing.T) {
	serverConn, peerConn := net.Pipe()
	defer peerConn.Close()

	server := newExecutionTestServer()
	go server.HandleConnection(serverConn)

	if err := peerConn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := peerConn.Write(redisCommand("PING").Marshal()); err != nil {
		t.Fatalf("write request: %v", err)
	}

	value, err := NewResp(peerConn).Read()
	if err != nil {
		t.Fatalf("read reply: %v", err)
	}
	got, ok := value.SimpleString()
	if !ok || got != "PONG" {
		t.Fatalf("reply = %q, %v; want PONG, true", got, ok)
	}
}

func TestExecDoesNotExposeIntermediateTransactionState(t *testing.T) {
	server := newExecutionTestServer()
	started := make(chan struct{})
	release := make(chan struct{})
	observerStarted := make(chan struct{})

	server.sh.registerCommand("TESTWAITSET", Command{
		Arity: 2,
		Flags: CmdWrite,
		Handler: func(c *Client, args []string) CommandResult {
			close(started)
			<-release
			return Set(c, args)
		},
	})
	server.sh.registerCommand("TESTOBSERVE", Command{
		Arity: 1,
		Flags: CmdRead,
		Handler: func(c *Client, args []string) CommandResult {
			close(observerStarted)
			return Get(c, args)
		},
	})

	transaction := newExecutionTestClient(server)
	observer := newExecutionTestClient(server)

	if result := transaction.HandleCommand(redisCommand("MULTI")); result.Failed {
		t.Fatalf("MULTI failed: %#v", result.Reply)
	}
	if result := transaction.HandleCommand(redisCommand("TESTWAITSET", "first", "one")); result.Failed {
		t.Fatalf("queue first write: %#v", result.Reply)
	}
	if result := transaction.HandleCommand(redisCommand("SET", "second", "two")); result.Failed {
		t.Fatalf("queue second write: %#v", result.Reply)
	}

	execDone := make(chan CommandResult, 1)
	go func() {
		execDone <- transaction.HandleCommand(redisCommand("EXEC"))
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("transaction did not start")
	}

	observeDone := make(chan CommandResult, 1)
	go func() {
		observeDone <- observer.HandleCommand(redisCommand("TESTOBSERVE", "second"))
	}()

	select {
	case <-observerStarted:
		t.Fatal("another client ran while EXEC was still in progress")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	if result := <-execDone; result.Failed {
		t.Fatalf("EXEC failed: %#v", result.Reply)
	}

	result := <-observeDone
	got, ok := result.Reply.BulkString()
	if result.Failed || !ok || got != "two" {
		t.Fatalf("observer reply = %#v; want bulk string two", result.Reply)
	}
}

func TestSlowSubscriberIsRemovedInsteadOfBlockingPublish(t *testing.T) {
	server := newExecutionTestServer()
	subscriber := newExecutionTestClient(server)
	subscriber.outbox = make(chan Value, 1)
	publisher := newExecutionTestClient(server)

	if result := subscriber.HandleCommand(redisCommand("SUBSCRIBE", "events")); result.Failed {
		t.Fatalf("SUBSCRIBE failed: %#v", result.Reply)
	}
	if result := publisher.HandleCommand(redisCommand("PUBLISH", "events", "first")); result.Failed {
		t.Fatalf("first PUBLISH failed: %#v", result.Reply)
	}
	if result := publisher.HandleCommand(redisCommand("PUBLISH", "events", "second")); result.Failed {
		t.Fatalf("second PUBLISH failed: %#v", result.Reply)
	}

	select {
	case <-subscriber.done:
	case <-time.After(time.Second):
		t.Fatal("slow subscriber was not disconnected")
	}

	if got := replyInteger(t, publisher.HandleCommand(redisCommand("PUBLISH", "events", "third"))); got != 0 {
		t.Fatalf("PUBLISH recipient count = %d, want 0 after disconnection", got)
	}
}

func TestPubSubDeliversAndAllowsUnsubscribe(t *testing.T) {
	server := newExecutionTestServer()
	subscriber := newExecutionTestClient(server)
	publisher := newExecutionTestClient(server)

	if result := subscriber.HandleCommand(redisCommand("SUBSCRIBE", "events")); result.Failed {
		t.Fatalf("SUBSCRIBE failed: %#v", result.Reply)
	}
	if got := replyInteger(t, publisher.HandleCommand(redisCommand("PUBLISH", "events", "ready"))); got != 1 {
		t.Fatalf("PUBLISH recipient count = %d, want 1", got)
	}

	select {
	case message := <-subscriber.outbox:
		parts, ok := message.Array()
		if !ok || len(parts) != 3 {
			t.Fatalf("published message = %#v, want three-part array", message)
		}
		payload, ok := parts[2].BulkString()
		if !ok || payload != "ready" {
			t.Fatalf("published payload = %q, %v; want %q, true", payload, ok, "ready")
		}
	default:
		t.Fatal("subscriber did not receive the published message")
	}

	if result := subscriber.HandleCommand(redisCommand("UNSUBSCRIBE", "events")); result.Failed {
		t.Fatalf("UNSUBSCRIBE failed: %#v", result.Reply)
	}
	if subscriber.mode != ModeNormal {
		t.Fatalf("client mode = %v, want normal", subscriber.mode)
	}
	if got := replyInteger(t, publisher.HandleCommand(redisCommand("PUBLISH", "events", "again"))); got != 0 {
		t.Fatalf("PUBLISH after unsubscribe recipient count = %d, want 0", got)
	}
}

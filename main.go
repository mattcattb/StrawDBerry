package main

import (
	"fmt"
	"go-redis/redis"
	"net"
	"time"
)

type ServerConfig struct {
	Addr string
}

type Config struct {
	aof    redis.AofConfig
	server ServerConfig
}

func defaultConfig() Config {
	return Config{
		aof: redis.AofConfig{
			DataDir:       "appendonly.aof",
			FPolicy:       redis.FsAlways,
			SnapshotEvery: time.Second,
		},
		server: ServerConfig{Addr: ":6379"},
	}
}

func main() {

	cfg := defaultConfig()
	db := redis.NewDb()
	exec := redis.NewExec(db)
	aof, err := redis.OpenAof(cfg.aof)

	if err != nil {
		// no aof here...
		fmt.Printf(err.Error())
		panic(err)
	}

	defer aof.Close()
	sh := redis.NewSH()

	srv := redis.NewServer(exec, aof, sh)
	srv.RegisterAllCommandHandlers()
	aof.ReplayAOF(redis.NewClient(nil, srv))

	fmt.Println("Listening on port :6379")

	// Create a new server
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		// Listen for connections
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go srv.HandleConnection(conn)
	}

}

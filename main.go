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
		server: ServerConfig{Addr: ":6479"},
	}
}

func main() {

	cfg := defaultConfig()
	db := redis.NewDb()
	aof, err := redis.OpenAof(cfg.aof)

	if err != nil {
		// no aof here...
		fmt.Print(err.Error())
		panic(err)
	}

	defer aof.Close()
	sh := redis.NewSH()

	srv := redis.NewServer(db, aof, sh)
	srv.RegisterAllCommandHandlers()
	aof.ReplayAOF(redis.NewClient(nil, srv))

	fmt.Println("Listening on port ", cfg.server.Addr)

	// Create a new server
	l, err := net.Listen("tcp", cfg.server.Addr)

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

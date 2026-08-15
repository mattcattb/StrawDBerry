package main

import (
	"fmt"
	"github.com/mattcattb/StrawDBerry/redis"
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
			FilePath:      "appendonly.aof",
			FSyncPolicy:   redis.FsAlways,
			FsyncInterval: time.Second,
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
	sh := redis.NewCommandTable()

	srv := redis.NewServer(db, aof, sh)
	srv.RegisterCMDTable()
	aof.ReplayAOF(redis.NewClient(nil, srv))

	fmt.Println("Listening on port ", cfg.server.Addr)

	// Create a new server
	l, err := net.Listen("tcp", cfg.server.Addr)

	if err != nil {
		fmt.Println(err)
		return
	}

	srv.SocketListenLoop(l)
}

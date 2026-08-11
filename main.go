package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"go-redis/redis"
)

type ServerConfig struct {
	Bind string
	Port uint16
}

type Config struct {
	aof    redis.AofConfig
	server ServerConfig
}

func loadConfig() (Config, error) {
	cfg := Config{
		aof: redis.AofConfig{
			Enabled:       true,
			FilePath:      "appendonly.aof",
			FSyncPolicy:   redis.FsAlways,
			FsyncInterval: time.Second,
		},
		server: ServerConfig{Bind: "0.0.0.0", Port: 6479},
	}

	if value, ok := os.LookupEnv("REDIS_BIND"); ok {
		cfg.server.Bind = value
	}
	if value, ok := os.LookupEnv("REDIS_PORT"); ok {
		port, err := strconv.ParseUint(value, 10, 16)
		if err != nil || port == 0 {
			return Config{}, fmt.Errorf("REDIS_PORT must be between 1 and 65535")
		}
		cfg.server.Port = uint16(port)
	}
	if value, ok := os.LookupEnv("REDIS_APPENDONLY"); ok {
		switch strings.ToLower(value) {
		case "yes", "true", "1":
			cfg.aof.Enabled = true
		case "no", "false", "0":
			cfg.aof.Enabled = false
		default:
			return Config{}, fmt.Errorf("REDIS_APPENDONLY must be yes or no")
		}
	}
	if value, ok := os.LookupEnv("REDIS_APPENDFILENAME"); ok {
		if value == "" {
			return Config{}, fmt.Errorf("REDIS_APPENDFILENAME cannot be empty")
		}
		cfg.aof.FilePath = value
	}
	if value, ok := os.LookupEnv("REDIS_APPENDFSYNC"); ok {
		switch strings.ToLower(value) {
		case "always":
			cfg.aof.FSyncPolicy = redis.FsAlways
		case "everysec":
			cfg.aof.FSyncPolicy = redis.FsEverySecond
		case "no":
			cfg.aof.FSyncPolicy = redis.FsNo
		default:
			return Config{}, fmt.Errorf("REDIS_APPENDFSYNC must be always, everysec, or no")
		}
	}

	return cfg, nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db := redis.NewDb()
	var persistence redis.PersistanceLog = &redis.DummyAofLog{}
	var aof *redis.Aof
	if cfg.aof.Enabled {
		aof, err = redis.OpenAof(cfg.aof)
		if err != nil {
			log.Fatal(err)
		}
		defer aof.Close()
		persistence = aof
	}

	sh := redis.NewCommandTable()

	srv := redis.NewServer(db, persistence, sh)
	srv.RegisterCMDTable()
	if aof != nil {
		if err := aof.ReplayAOF(redis.NewClient(nil, srv)); err != nil {
			log.Fatal(err)
		}
	}

	addr := net.JoinHostPort(cfg.server.Bind, strconv.Itoa(int(cfg.server.Port)))
	fmt.Println("Listening on", addr)

	// Create a new server
	l, err := net.Listen("tcp", addr)

	if err != nil {
		fmt.Println(err)
		return
	}

	srv.SocketListenLoop(l)
}

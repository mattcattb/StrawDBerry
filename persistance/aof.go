package persistance

import "time"

// all commands written to aof file

type Aof struct {
}

type Config struct {
	DataDir       string
	FlushDuration time.Duration
	SnapshotEvery time.Duration
}

type Engine struct {
	config Config
	store  *Store
}

func AppendCommand() {

}

func Replay() {

}

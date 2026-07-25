package redis

import (
	"bufio"
	"os"
	"sync"
	"time"
)

// all commands written to aof file

type FsyncPolicy uint8

const (
	FsAlways FsyncPolicy = iota << 1
	FsEverySecond
	FsNo
)

type Aof struct {
	writer   *bufio.Writer
	file     *os.File
	mu       sync.Mutex
	fsPolicy FsyncPolicy
	ticker   *time.Ticker
	config   AofConfig
}

type AofConfig struct {
	Enabled       bool
	FilePath      string
	FSyncPolicy   FsyncPolicy
	FsyncInterval time.Duration
}

/*
	APPEND     serialize and append one command to AOF
	FLUSH     flush bufio.Writer into the OS file
	SYNC     ask OS to flush file contents to disk
	CLOSE     flush/sync if needed, then close file
	REPLAY read AOF commands and execute them into a DB
*/
// consider persistnace interface

func OpenAof(config AofConfig) (a *Aof, err error) {
	f, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	writer := bufio.NewWriter(f)
	a = &Aof{file: f, writer: writer, fsPolicy: config.FSyncPolicy}

	if a.fsPolicy == FsEverySecond {
		a.startFsyncInterval(config.FsyncInterval)
	}

	return a, nil
}

func (a *Aof) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ticker != nil {
		a.ticker.Stop()
	}

	if err := a.writer.Flush(); err != nil {
		return err
	}

	if err := a.file.Sync(); err != nil {
		return err
	}
	return a.file.Close()
}

func (a *Aof) startFsyncInterval(interval time.Duration) {
	a.ticker = time.NewTicker(interval)

	go func() {
		for range a.ticker.C {
			a.mu.Lock()
			a.file.Sync()
			a.mu.Unlock()
		}
	}()
}

func (a *Aof) Append(v Value) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a == nil || a.file == nil {
		return nil
	}
	bytes := v.Marshal()

	a.writer.Write(bytes)

	err := a.writer.Flush()
	if err != nil {
		return err
	}

	if a.fsPolicy == FsAlways {
		return a.file.Sync()
	}

	return nil
}

func (a *Aof) ReplayAOF(executor *Client) error {
	// replay AOF

	// we need to modify the client executor here

	executor.aof = &DummyAofLog{}

	bufReader := bufio.NewReader(a.file)

	resp := Resp{reader: bufReader}

	for {
		v, err := resp.Read()
		if err != nil {
			return err
		}

		_ = executor.HandleCommand(v)
		// uhhhh uhhh uhhh hmmmm
	}

	return nil

}

type DummyAofLog struct {
}

func (a *DummyAofLog) Append(v Value) error {
	return nil
}

func (a *Aof) CompactLog(v Value) error {
	// we need to do a few things...
	return nil
}

type AofStats struct {
	fCurrentSize uint64 // AOF current file size
	fBaseSize    uint64 // AOF file size on latest startup or rewrite
	bufferLength uint64 // AOF buffer size
	fsyncCount   uint64
	fsyncErrors  uint64

	// aof_enabled
	//todo rewrites added
	// aof_rewrites
	// aof_rewrite_buffer_length
	// aof_pending_rewrite bool
	// aof_rewrite_in_progress
}

/*
loading_start_time: Epoch-based timestamp of the start of the load operation
loading_total_bytes: Total file size
loading_rdb_used_mem: The memory usage of the server that had generated the RDB file at the time of the file's creation
loading_loaded_bytes: Number of bytes already loaded
loading_loaded_perc: Same value expressed as a percentage
loading_eta_seconds: ETA in seconds for the load to be complete
*/

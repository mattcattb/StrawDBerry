package persistance

import (
	"go-redis/redis"
	"go-redis/resp"
	"os"
	"sync"
	"time"
)

// all commands written to aof file

type Aof struct {
	file *os.File
	path string
	mu   sync.Mutex
}

type Config struct {
	DataDir       string
	FlushDuration time.Duration
	SnapshotEvery time.Duration
}

/*

	APPEND     serialize and append one command to AOF


	FLUSH     flush bufio.Writer into the OS file

	SYNC     ask OS to flush file contents to disk

	CLOSE     flush/sync if needed, then close file

	REPLAY read AOF commands and execute them into a DB

*/

func (a *Aof) Open() error

func (a *Aof) Append(v resp.Value) error {
	bytes := v.Marshal()
	_, err := a.file.Write(bytes)
	if err != nil {
		return err
	}

	return a.file.Sync()
}

func (a *Aof) Sync(command string, args []resp.Value) {

}

func ReplayAOF(path string, executor *redis.CommandExecutor) {
	// replay AOF
}

# go-redis

My implementation of redis in golang from scratch.

This is a learning project focused on the lower-level pieces: RESP parsing and writing, command execution, an in-memory keyspace, and basic persistence ideas. The code is still in progress and is not intended to be a production Redis replacement.

## Features

- RESP protocol parsing and serialization
- core string, hash, set, and list general commands
- Multi, Exec, + Discard transactional commands
- AOF memory persistance (log rewrite in progress)
- Pubsub Publish, Subscribe and Unsubscribe with client pubsub blocking mode on subscriptions
- General + Connection commmands

## In Progress:

- AUTH
- Pubsub \* channel mutli subscriptions
- Zset commands
- Blocking behaviors
- AOF file rewrites
- redis stream commands
- memory management
- lru caching for memory limits
- commandstats
- Service Managment Config + limits

#### management

- Info Command ()
- Memory tracking + max memory
- LRU caching
- Hotkey Management (detection of high frequency "hotkeys")

Run tests with:

```sh
go test ./...
```

## Configuration

The server can be configured with environment variables named after the equivalent Redis directives:

| Environment variable | Default | Values |
| --- | --- | --- |
| `REDIS_BIND` | `0.0.0.0` | IP address or hostname to listen on |
| `REDIS_PORT` | `6479` | `1` through `65535` |
| `REDIS_APPENDONLY` | `yes` | `yes` or `no` (also accepts boolean forms) |
| `REDIS_APPENDFILENAME` | `appendonly.aof` | AOF file path |
| `REDIS_APPENDFSYNC` | `always` | `always`, `everysec`, or `no` |

For example:

```sh
REDIS_BIND=127.0.0.1 REDIS_PORT=6379 REDIS_APPENDFSYNC=everysec go run .
```

With Docker, publish the same container port selected by `REDIS_PORT` and mount `/data` to persist the AOF:

```sh
docker run --rm -p 6379:6379 -v redis-data:/data \
  -e REDIS_PORT=6379 -e REDIS_APPENDFSYNC=everysec go-redis
```

## Larger Features

- key frequencies
- Streams
- Latency Behaviors
- Replication
- ACL
- Fine tuned memory tracking
- RDB
- Monitoring
- Blocking
- Geo Indexes
- watch + unwatch mutli exec
- Lua Scripting

# StrawDBerry

<p align="center">
  <img src="assets/strawdberry-icon.png" alt="StrawDBerry logo" width="220" />
</p>

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

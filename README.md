# go-redis

A small Redis-inspired server written in Go.

This is a learning project focused on the lower-level pieces: RESP parsing and writing, command execution, an in-memory keyspace, and basic persistence ideas. The code is still in progress and is not intended to be a production Redis replacement.

Current areas of work:

- RESP protocol parsing and serialization
- string and hash command handling
- a shared in-memory keyspace for Redis-style objects
- append-only persistence experiments

Run tests with:

```sh
go test ./...
```

package redis

/* TODO!!!! type serverStats struct {
	redis_version string;
	redis_git_dirty int32;  //Git dirty flag
	process_id uint64 // PID of the server process
  tcp_port uint16 // TCP/IP listen port
	// server_time_usec  Epoch-based system time with microsecond precision
	uptime_in_seconds uint64 // Number of seconds since Redis server start
	uptime_in_days  uint16 // Same value expressed in days
	hz uint64  // The server's current frequency setting
// configured_hz: The server's configured frequency setting
// lru_clock: Clock incrementing every minute, for LRU management
// executable: The path to the server's executable
// config_file: The path to the config file
// io_threads_active: Flag indicating if I/O threads are active
// //no replication shutdown_in_milliseconds: The maximum time remaining for replicas to catch up the replication before completing the shutdown sequence. This field is only present during shutdown
}

type clientStats struct {
	connected_clients uint64 // : Number of client connections (excluding connections from replicas)
	maxclients uint64 // The value of the maxclients configuration directive. This is the upper limit for the sum of connected_clients, connected_slaves and cluster_connections.
	// client_recent_max_input_buffer: Biggest input buffer among current client connections
	// client_recent_max_output_buffer: Biggest output buffer among current client connections
// todo if blocking added	blocked_clients uint64 // Number of clients pending on a blocking call (BLPOP, BRPOP, BRPOPLPUSH, BLMOVE, BZPOPMIN, BZPOPMAX)
//! tracking_clients: Number of clients being tracked (CLIENT TRACKING)
	pubsub_clients uint64 // Number of clients in pubsub mode (SUBSCRIBE, PSUBSCRIBE, SSUBSCRIBE). Added in Redis 7.4
//!	watching_clients uint64 Number of clients in watching mode (WATCH). Added in Redis 7.4
// clients_in_timeout_table: Number of clients in the clients timeout table
//! total_watched_keys: Number of watched keys. Added in Redis 7.4.
//todo when blocking added total_blocking_keys: Number of blocking keys. Added in Redis 7.2.
// total_blocking_keys_on_nokey: Number of blocking keys that one or more clients that would like to be unblocked when the key is deleted. Added in Redis 7.2
}

// based on the command type
type commandStats struct {
	calls uint64 //
	usec uint64 // the total CPU time consumed by these commands
	usec_per_call uint64 // the average CPU consumed per command execution
	rejected_calls uint64 // the number of rejected calls
	failed_calls uint64  // the number of failed calls
}

type MemoryStats struct {
	used_memory uint64 // Total number of bytes allocated by Redis using its allocator (either standard libc, jemalloc, or an alternative allocator such as tcmalloc)
	used_memory_human uint64 // Human readable representation of previous value
 // used_memory_rss: Number of bytes that Redis allocated as seen by the operating system (a.k.a resident set size). This is the number reported by tools such as top(1) and ps(1)
// used_memory_rss_human: Human readable representation of previous value
 used_memory_peak uint64 // Peak memory consumed by Redis (in bytes)
// used_memory_peak_human: Human readable representation of previous value
used_memory_peak_time uint64 //  Time when peak memory was recorded
used_memory_peak_perc uint64 // The percentage of used_memory out of used_memory_peak
//? used_memory_overhead: The sum in bytes of all overheads that the server allocated for managing its internal data structures
//? used_memory_startup: Initial amount of memory consumed by Redis at startup in bytes
used_memory_dataset uint64 // The size in bytes of the dataset (used_memory_overhead subtracted from used_memory)
used_memory_dataset_perc: The percentage of used_memory_dataset out of the net memory usage (used_memory minus used_memory_startup)
total_system_memory: The total amount of memory that the Redis host has
total_system_memory_human: Human readable representation of previous value
//! used_memory_lua: Number of bytes used by the Lua engine for EVAL scripts. Deprecated in Redis 7.0, renamed to used_memory_vm_eval
//! used_memory_vm_eval: Number of bytes used by the script VM engines for EVAL framework (not part of used_memory). Added in Redis 7.0
//! used_memory_lua_human: Human readable representation of previous value. Deprecated in Redis 7.0
//! used_memory_scripts_eval: Number of bytes overhead by the EVAL scripts (part of used_memory). Added in Redis 7.0
//! number_of_cached_scripts: The number of EVAL scripts cached by the server. Added in Redis 7.0
//!  number_of_functions uint16 // The number of functions. Added in Redis 7.0
//! number_of_libraries uint16 // The number of libraries. Added in Redis 7.0
// used_memory_functions: Number of bytes overhead by Function scripts (part of used_memory). Added in Redis 7.0
//! used_memory_scripts: used_memory_scripts_eval + used_memory_functions (part of used_memory). Added in Redis 7.0
//! used_memory_scripts_human: Human readable representation of previous value
maxmemory uint64 //: The value of the maxmemory configuration directive
// maxmemory_human: Human readable representation of previous value
maxmemory_policy uint64 //: The value of the maxmemory-policy configuration directive
// mem_fragmentation_ratio: Ratio between used_memory_rss and used_memory. Note that this doesn't only includes fragmentation, but also other process overheads (see the allocator_* metrics), and also overheads like code, shared libraries, stack, etc.
// mem_fragmentation_bytes: Delta between used_memory_rss and used_memory. Note that when the total fragmentation bytes is low (few megabytes), a high ratio (e.g. 1.5 and above) is not an indication of an issue.
/*
allocator_frag_ratio:: Ratio between allocator_active and allocator_allocated. This is the true (external) fragmentation metric (not mem_fragmentation_ratio).
allocator_frag_bytes Delta between allocator_active and allocator_allocated. See note about mem_fragmentation_bytes.
allocator_rss_ratio: Ratio between allocator_resident and allocator_active. This usually indicates pages that the allocator can and probably will soon release back to the OS.
allocator_rss_bytes: Delta between allocator_resident and allocator_active
rss_overhead_ratio: Ratio between used_memory_rss (the process RSS) and allocator_resident. This includes RSS overheads that are not allocator or heap related.
rss_overhead_bytes: Delta between used_memory_rss (the process RSS) and allocator_resident
allocator_allocated: Total bytes allocated form the allocator, including internal-fragmentation. Normally the same as used_memory.
allocator_active: Total bytes in the allocator active pages, this includes external-fragmentation.
allocator_resident: Total bytes resident (RSS) in the allocator, this includes pages that can be released to the OS (by MEMORY PURGE, or just waiting).
allocator_muzzy: Total bytes of 'muzzy' memory (RSS) in the allocator. Muzzy memory is memory that has been freed, but not yet fully returned to the operating system. It can be reused immediately when needed or reclaimed by the OS when system pressure increases.
mem_not_counted_for_evict: Used memory that's not counted for key eviction. This is basically transient replica and AOF buffers.
mem_clients_slaves: Memory used by replica clients - Starting Redis 7.0, replica buffers share memory with the replication backlog, so this field can show 0 when replicas don't trigger an increase of memory usage.
mem_clients_normal: Memory used by normal clients
mem_cluster_links: Memory used by links to peers on the cluster bus when cluster mode is enabled.
mem_cluster_slot_migration_output_buffer: Memory usage of the migration client's output buffer. Redis writes incoming changes to this buffer during the migration process.
mem_cluster_slot_migration_input_buffer: Memory usage of the accumulated replication stream buffer on the importing node.
mem_cluster_slot_migration_input_buffer_peak: Peak accumulated repl buffer size on the importing side.
todo aof mem_aof_buffer: Transient memory used for AOF and AOF rewrite buffers
! mem_replication_backlog: Memory used by replication backlog
? mem_total_replication_buffers: Total memory consumed for replication buffers - Added in Redis 7.0.
mem_allocator: Memory allocator, chosen at compile time.
mem_overhead_db_hashtable_rehashing: Temporary memory overhead of database dictionaries currently being rehashed - Added in 7.4.
active_defrag_running: When activedefrag is enabled, this indicates whether defragmentation is currently active, and the CPU percentage it intends to utilize.
lazyfree_pending_objects: The number of objects waiting to be freed (as a result of calling UNLINK, or FLUSHDB and FLUSHALL with the ASYNC option)
lazyfreed_objects: The number of objects that have been lazy freed
*/

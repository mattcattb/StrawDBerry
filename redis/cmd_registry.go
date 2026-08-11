package redis

var StringCmdTable map[string]Command = map[string]Command{
	"GET": {
		Arity:   1,
		Flags:   CmdRead,
		Handler: Get,
	},
	"SET": {
		Arity:   -2,
		Flags:   CmdWrite,
		Handler: Set,
	},
	"MGET": {
		Arity:   -1,
		Flags:   CmdRead,
		Handler: MGet,
	},

	"MSET": {
		Arity:   -2,
		Flags:   CmdWrite,
		Handler: MSet,
	},
	"INCR": {
		Arity:   1,
		Handler: Incr,
		Flags:   CmdWrite,
	},
	"DECR": {
		Arity:   1,
		Handler: Decr,
		Flags:   CmdWrite,
	},

	"INCRBY": {
		Arity:   2,
		Handler: IncrBy,
		Flags:   CmdWrite,
	},

	"DECRBY": {
		Arity:   2,
		Handler: DecrBy,
		Flags:   CmdWrite,
	},
	"STRLEN": {
		Arity:   1,
		Handler: StrLen,
		Group:   StringGroup,
		Flags:   CmdRead,
	},
	"LCS": {
		Arity:   -2,
		Handler: Lcs,
		Flags:   CmdRead,
	},
}

var HashCmdTable map[string]Command = map[string]Command{
	"HSET": {
		Arity:   -3,
		Flags:   CmdWrite,
		Handler: HSet,
	},
	"HGET": {
		Arity:   2,
		Flags:   CmdRead,
		Handler: HGet,
	},

	"HDEL": {
		Arity:   -2,
		Flags:   CmdWrite,
		Handler: HDel},
	"HGETALL": {
		Arity:   1,
		Flags:   CmdRead,
		Handler: HGetAll,
	},
	"HEXISTS": {
		Arity:   2,
		Flags:   CmdRead,
		Handler: HExists,
	}}

var SetCmdTable map[string]Command = map[string]Command{"SADD": {
	Handler: SAdd,
	Group:   SetGroup,
	Flags:   CmdWrite,
	Arity:   -2,
},
	"SCARD": {
		Handler: SCard,
		Group:   SetGroup,
		Flags:   CmdRead,
		Arity:   1,
	},
	"SREM": {
		Handler: SRem,
		Group:   SetGroup,
		Flags:   CmdWrite,
		Arity:   -2,
	},
	"SMISMEMBER": {
		Handler: SMIsMem,
		Group:   SetGroup,
		Flags:   CmdRead,
		Arity:   -2,
	},
	"SISMEMBER": {
		Handler: SIsMem,
		Group:   SetGroup,
		Flags:   CmdRead,
		Arity:   2,
	},
	"SMEMBERS": {
		Handler: SMembers,
		Group:   SetGroup,
		Flags:   CmdRead,
		Arity:   1,
	},
	"SDIFF": {
		Handler: SDiff,
		Arity:   -1,
		Group:   SetGroup,
		Flags:   CmdRead,
	},
	"SINTER": {
		Handler: SInter,
		Arity:   -1,
		Group:   SetGroup,
		Flags:   CmdRead,
	},
	"SUNION": {
		Handler: SUnion,
		Arity:   -1,
		Group:   SetGroup,
		Flags:   CmdRead,
	}}

var ManagementCMDTable map[string]Command = map[string]Command{
	"INFO": {
		Handler: Info,
		Group:   ManagementGroup,
		Flags:   CmdRead,
		Arity:   -1,
	},
	"DBSIZE": {
		Arity:   0,
		Handler: DbSize,
		Group:   ManagementGroup,
	},

	"COMMAND": {
		Handler: CommandList,
		Arity:   0,
		subcommands: map[string]Command{
			"LIST":  {Arity: 0, Handler: CommandList, Flags: CmdRead},
			"COUNT": {Arity: 0, Handler: CommandCount, Flags: CmdRead},
		},
	},
}

var GenericCMDTable map[string]Command = map[string]Command{
	"KEYS": {
		Arity:   1,
		Handler: Keys,
		Flags:   CmdRead,
		Group:   GenericGroup,
	},
	"SCAN": {
		Arity:   -1,
		Handler: Scan,
		Flags:   CmdRead,
		Group:   GenericGroup,
	},
	"COPY": {
		Arity:   -2,
		Handler: Copy,
		Flags:   CmdWrite,
		Group:   GenericGroup,
	},
	"EXISTS": {
		Handler: Exists,
		Arity:   -1,
		Group:   GenericGroup,
		Flags:   CmdRead,
	},
	"TYPE": {
		Arity:   1,
		Group:   GenericGroup,
		Flags:   CmdRead,
		Handler: Type,
	},
	"TTL": {
		Arity:   1,
		Group:   GenericGroup,
		Flags:   CmdRead,
		Handler: Ttl,
	},
	"PERSIST": {
		Arity:   1,
		Group:   GenericGroup,
		Handler: Persist,
		Flags:   CmdWrite,
	},
	"DEL": {
		Arity:   -1,
		Group:   GenericGroup,
		Flags:   CmdWrite,
		Handler: Del,
	},
	"DUMP": {
		Handler: Dump,
		Arity:   1,
	},
	"RESTORE": {
		Handler: Restore,
		Arity:   1,
	},

	"OBJECT": {
		subcommands: map[string]Command{
			"ENCODING": {
				Arity:   1,
				Handler: ObjectEncodingCmd,
				Group:   GenericGroup,
				Flags:   CmdRead,
			},
		},
	},

	"FLUSHALL": {
		Arity:   0,
		Handler: FlushAll,
		Group:   GenericGroup,
		Flags:   CmdWrite,
	},
}

var ConnectionCmdTable map[string]Command = map[string]Command{
	"PING": {
		Arity:   0,
		Handler: Ping,
		Group:   ConnGroup,
		Flags:   CmdAllowedInPubsub,
	}, "ECHO": {
		Arity:   1,
		Handler: Echo,
		Group:   ConnGroup,
	}}

var PubsubCmdTable map[string]Command = map[string]Command{
	"SUBSCRIBE": {
		Handler: Subscribe,
		Arity:   1,
		Group:   PubsubGroup,
		Flags:   CmdAllowedInPubsub | CmdNoMulti,
	}, "UNSUBSCRIBE": {
		Handler: Unsubscribe,
		Arity:   1,
		Group:   PubsubGroup,
		Flags:   CmdAllowedInPubsub | CmdNoMulti,
	},
	"PUBLISH": {
		Handler: Publish,
		Arity:   2,
		Group:   PubsubGroup,
	},
}

var TxCmdTable map[string]Command = map[string]Command{
	"EXEC": {
		Arity:   0,
		Handler: Exec,
		Group:   TxGroup,
	},
	"MULTI": {
		Arity:   0,
		Group:   TxGroup,
		Handler: Multi,
		Flags:   CmdNoMulti, // cannot call multi if in multi
	},
	"DISCARD": {
		Arity:   0,
		Group:   TxGroup,
		Handler: Discard,
	},
}

// Zset commands are still in progress. Keep them out of the default command
// table until the ranking/range implementation is complete.

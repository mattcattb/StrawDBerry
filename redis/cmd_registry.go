package redis

func registerTstringCommands(sh *CommandTable) {

	specMap := map[string]Command{
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

	for k, v := range specMap {
		v.Group = StringGroup
		v.Name = k
		specMap[k] = v
	}

	sh.registerCommandSpecs(specMap)
}

func registerTHashCommandSpec(sh *CommandTable) {

	hSpecs := map[string]Command{
		"HSET": {
			Arity:   -3,
			Flags:   CmdWrite,
			Handler: HSet,
		},

		"HDEL": {
			Arity:   2,
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
		},
	}

	for k, v := range hSpecs {
		v.Group = HashGroup
		v.Name = k
		hSpecs[k] = v
	}
	sh.registerCommandSpecs(hSpecs)

}

func registerSetCSpec(sh *CommandTable) {

	setSpecs := map[string]Command{

		"SADD": {
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
		"SDIFF": {
			Handler: SDiff,
			Arity:   -1,
			Group:   SetGroup,
			Flags:   CmdRead,
		},
	}

	sh.registerCommandSpecs(setSpecs)

}

func registerGenericCommands(sh *CommandTable) {
	sh.registerCommandSpecs(map[string]Command{
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
			Flags:   CmdWrite,
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
		"OBJECT": {
			Arity:   2,
			Handler: ObjCommand,
			Flags:   CmdRead,
			Group:   GenericGroup,
		},
	})

}

// ! if we have all of the sub commands under a single command will be treated as same in logging
func registerMangementSpec(sh *CommandTable) {
	mSpecs := map[string]Command{

		"INFO": {
			Handler: Info,
			Group:   ManagementGroup,
			Flags:   CmdRead,
			Arity:   -1,
		},
	}

	for k, v := range mSpecs {
		v.Group = ManagementGroup
		v.Name = k
		mSpecs[k] = v
	}

	sh.registerCommandSpecs(mSpecs)
}

func (sh *CommandTable) RegisterFullCommandTable() {
	registerConnectionCommands(sh)
	registerGenericCommands(sh)
	registerTransactionSpec(sh)
	registerPubsubCommands(sh)
	registerTHashCommandSpec(sh)
	registerTstringCommands(sh)
	registerSetCSpec(sh)
}

package redis

func registerTstringCommands(sh *SpecHandler) {

	specMap := map[string]CommandSpec{
		"GET": {
			arity:   1,
			flags:   CmdRead,
			handler: Get,
		},
		"SET": {
			arity:   -2,
			flags:   CmdWrite,
			handler: Set,
		},
		"MGET": {
			arity:   -1,
			flags:   CmdRead,
			handler: MGet,
		},

		"MSET": {
			arity:   -2,
			flags:   CmdWrite,
			handler: MSet,
		},
		"INCR": {
			arity:   1,
			handler: Incr,
			flags:   CmdWrite,
		},
		"DECR": {
			arity:   1,
			handler: Decr,
			flags:   CmdWrite,
		},

		"INCRBY": {
			arity:   2,
			handler: IncrBy,
			flags:   CmdWrite,
		},

		"DECRBY": {
			arity:   2,
			handler: DecrBy,
			flags:   CmdWrite,
		},
		"STRLEN": {
			arity:   1,
			handler: StrLen,
			group:   StringGroup,
			flags:   CmdRead,
		},
		"LCS": {
			arity:   -2,
			handler: Lcs,
			flags:   CmdRead,
		},
	}

	for k, v := range specMap {
		v.group = StringGroup
		v.name = k
		specMap[k] = v
	}

	sh.registerCommandSpecs(specMap)
}

func registerTHashCommandSpec(sh *SpecHandler) {

	hSpecs := map[string]CommandSpec{
		"HSET": {
			arity:   -3,
			flags:   CmdWrite,
			handler: HSet,
		},

		"HDEL": {
			arity:   2,
			flags:   CmdWrite,
			handler: HDel},
		"HGETALL": {
			arity:   1,
			flags:   CmdRead,
			handler: HGetAll,
		},
		"HEXISTS": {
			arity:   2,
			flags:   CmdRead,
			handler: HExists,
		},
	}

	for k, v := range hSpecs {
		v.group = HashGroup
		v.name = k
		hSpecs[k] = v
	}
	sh.registerCommandSpecs(hSpecs)

}

func registerSetCSpec(sh *SpecHandler) {

	setSpecs := map[string]CommandSpec{

		"SADD": {
			handler: SAdd,
			group:   SetGroup,
			flags:   CmdWrite,
			arity:   -2,
		},
		"SCARD": {
			handler: SCard,
			group:   SetGroup,
			flags:   CmdRead,
			arity:   1,
		},
		"SREM": {
			handler: SRem,
			group:   SetGroup,
			flags:   CmdWrite,
			arity:   -2,
		},
		"SMISMEMBER": {
			handler: SMIsMem,
			group:   SetGroup,
			flags:   CmdRead,
			arity:   -2,
		},
		"SISMEMBER": {
			handler: SIsMem,
			group:   SetGroup,
			flags:   CmdRead,
			arity:   2,
		},
		"SDIFF": {
			handler: SDiff,
			arity:   -1,
			group:   SetGroup,
			flags:   CmdRead,
		},
	}

	sh.registerCommandSpecs(setSpecs)

}

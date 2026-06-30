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

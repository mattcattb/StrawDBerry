package command

import "go-redis/resp"

func (ce *CommandExecutor) Set(args []resp.Value) resp.Value {
	if len(args) <= 2 {
		return resp.ErrorValue("insufficient args for set")
	}

	key, ok := args[0].BulkString()
	if !ok {
		return resp.ErrorValue("ERR invalid key")
	}
	val, ok := args[1].BulkString()

	if !ok {
		return resp.ErrorValue("ERR invalid val")
	}

	// [EX seconds | PX milliseconds


	ce.store.

}

func (ce *CommandExecutor) Get(args []resp.Value) resp.Value {

}

func (ce *CommandExecutor) Del(args []resp.Value) resp.Value {

}

func (ce *CommandExecutor) Incr(args []resp.Value) resp.Value {

}

func (ce *CommandExecutor) MGet(args []resp.Value) resp.Value {

}

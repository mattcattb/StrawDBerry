package redis

import "go-redis/resp"

func newHashRObject(val map[string]string) *RedisObject {
	return &RedisObject{
		typ:      HashObject,
		encoding: EncodingMap,
		ptr:      val,
	}
}

func hashObjValue(obj *RedisObject) (map[string]string, error) {

	var newMap map[string]string
	if obj.typ != HashObject {
		return newMap, ErrWrongType
	}

	switch obj.encoding {
	case EncodingMap:
		val, ok := obj.ptr.(map[string]string)
		if !ok {
			return newMap, ErrInvalidEncoding
		}

		return val, nil
	}

	return newMap, ErrInvalidEncoding
}

func setHashObjValue(obj *RedisObject) {
	obj.encoding = EncodingMap
	obj.ptr = obj
}

func (ce *CommandExecutor) HGet(args []resp.Value) resp.Value {
	// HGET key field
	// returns resp.Array field, value

}

func (ce *CommandExecutor) HSet(args []resp.Value) resp.Value {
	// HSET key field value [field value ...]
	// returns int of set fields

}

func (ce *CommandExecutor) HDel(args []resp.Value) resp.Value {
	// key field [field ...]

}
func (ce *CommandExecutor) HGetAll(args []resp.Value) resp.Value {
	// HGETALL key
	// returns resp.Array field, value

}

func (ce *CommandExecutor) HExists(args []resp.Value) resp.Value {
	// HEXISTS key field

}

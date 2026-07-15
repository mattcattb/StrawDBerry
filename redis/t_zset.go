//go:build zset

package redis

import "github.com/huandu/skiplist"

// skiplist encoding

func zsetObjValue(o *RedisObject) (skipListPayload, error) {

	if o.typ != ZSetObject {
		return skipListPayload{}, ErrWrongType
	}

	if o.encoding != EncodingSkipList {
		return skipListPayload{}, ErrWrongType
	}

	p, ok := o.ptr.(skipListPayload)

	if !ok {
		return skipListPayload{}, ErrWrongType
	}

	return p, nil
}

func newZsetValue() *skipListPayload {
	return &skipListPayload{
		scores: map[string]int{},
		sl:     skiplist.New(skiplist.Int),
	}
}

func rankingZsetValue(obj *RedisObject, member string) (int, bool) {
	// uhhhh hmmmm
}

func rangeZsetValues(obj *RedisObject, count int, top bool) {
	// get members either top or bottom from the zset
}

func removeZsetMember(obj *RedisObject, member string) (int, bool) {
	// remove a member from the zset
}

func Zadd(c *Client, args []string) CommandResult {
	// > ZADD racer_scores 8 "Sam-Bodden" 10 "Royce" 6 "Ford" 14 "Prickett"

	key, score, member := args[0], args[1], args[2]

	scoreMemberPairs := args[1:]

}

func ZRank(c *Client, args []string) CommandResult {
	key, member := args[0], args[1]

	withScore := len(args) > 2

	if withScore && args[2] != "WITHSCORE" {
		return Failed(wrongArgs("ZRANK"))
	}

	rObj, ok := c.db.lookupKey(key)

	if !ok {
		return Result(Null())
	}

}

func ZRange(c *Client, args []string) CommandResult {
	// > ZADD racer_scores 8 "Sam-Bodden" 10 "Royce" 6 "Ford" 14 "Prickett"

}

func ZRevRange(c *Client, args []string) CommandResult {
	// > ZADD racer_scores 8 "Sam-Bodden" 10 "Royce" 6 "Ford" 14 "Prickett"

}

func ZRangeByScore(c *Client, args []string) CommandResult {

}

func ZRem(c *Client, args []string) CommandResult {}

func ZRemRangeByScore(c *Client, args []string) CommandResult {}

package redis

import (
	"math"
	"strconv"
	"strings"
)

type zsetEntry struct {
	Member string
	Score  float64
}

func zsetTypeLen(obj *RedisObject) (uint64, error) {

	switch obj.encoding {
	case EncodingSkipList:
		zset, err := zsetSkiplistValue(obj)

		if err != nil {
			return 0, err
		}
		return uint64(zset.ordered.len), nil

	// case for encoding listpack

	default:
		return 0, ErrInvalidEncoding

	}
}

func zsetTypeScore(obj *RedisObject, member string) (score float64, found bool, err error) {
	switch obj.encoding {
	case EncodingSkipList:
		zs, err := zsetSkiplistValue(obj)

		if err != nil {
			return score, found, err
		}

		score, found = zs.score(member)

		return score, found, nil

	default:
		return score, found, ErrInvalidEncoding
	}
}

type zsetAddOptions struct {
	NX   bool
	XX   bool
	GT   bool
	LT   bool
	Incr bool
	CH   bool
}

func parseZaddOptions(args []string) (za zsetAddOptions, pairStart int) {

	for pairStart < len(args) {
		switch strings.ToUpper(args[pairStart]) {
		case "NX":
			za.NX = true
			za.XX = false
		case "XX":
			za.XX = true
			za.NX = false

		case "GT":
			za.GT = true
			za.LT = false
		case "LT":
			za.LT = true
			za.GT = false

		case "CH":
			za.CH = true
		case "INCR":
			za.Incr = true
		default:
			break
		}
		pairStart++
	}

	return za, pairStart
}

type zsetAddResult struct {
	Added   bool
	Updated bool
	Score   float64
}

func zsetTypeAdd(obj *RedisObject, entry zsetEntry, options zsetAddOptions) (result zsetAddResult, err error) {
	switch obj.encoding {

	case EncodingSkipList:

		zs, err := zsetSkiplistValue(obj)

		if err != nil {
			return result, err
		}

		curNode, exists := zs.lookup(entry.Member)

		if !exists {
			// we can just insert like normal?

			if options.XX {
				//! cannot add new members
				return zsetAddResult{Added: false, Updated: false}, nil
			}

			if err := zs.insertNew(entry.Member, entry.Score); err != nil {
				return zsetAddResult{}, err
			}

			return zsetAddResult{Added: true, Updated: false, Score: entry.Score}, nil
		}

		if options.NX {
			// prevent updating
			return zsetAddResult{Added: false, Updated: false}, nil
		}

		if options.GT && !(entry.Score > curNode.score) {
			return zsetAddResult{Added: false, Updated: false, Score: curNode.score}, nil
		}

		if options.LT && !(entry.Score < curNode.score) {
			return zsetAddResult{Added: false, Updated: false, Score: curNode.score}, nil
		}

		updateScore := entry.Score

		if options.Incr {
			updateScore += curNode.score
		}

		updated, err := zs.updateScore(curNode, updateScore)

		if err != nil {
			return zsetAddResult{}, err
		}

		return zsetAddResult{Added: false, Updated: updated, Score: updateScore}, nil

	default:
		return result, ErrInvalidEncoding
	}
}

func zsetTypeRemove(obj *RedisObject, member string) (removed bool, err error) {
	switch obj.encoding {
	case EncodingSkipList:
		zs, err := zsetSkiplistValue(obj)
		if err != nil {
			return removed, err
		}

		removed := zs.remove(member)
		return removed, nil
	default:
		return removed, ErrInvalidEncoding
	}
}

func zsetTypeRank(obj *RedisObject, member string) (rank uint64, found bool, err error) {
	switch obj.encoding {
	case EncodingSkipList:
		zs, err := zsetSkiplistValue(obj)
		if err != nil {
			return rank, found, err
		}

		rank, found := zs.rank(member)

		return uint64(rank), found, nil
	default:
		return rank, found, ErrInvalidEncoding
	}
}

type zrangeMode uint8

const (
	zrangeByRank zrangeMode = iota
	zrangeByScore
	zrangeByLex
)

type zrangeOptions struct {
	mode       zrangeMode
	reverse    bool
	withScores bool
	limit      *zrangeLimit
}

type zrangeLimit struct {
	offset int64
	count  int64
}

type rankRange struct {
	start int64
	stop  int64
}

type scoreBound struct {
	value     float64
	exclusive bool
}

type scoreRange struct {
	min scoreBound
	max scoreBound
}

func zsetTypeRangeByRank(obj *RedisObject, rankRange rankRange, rev bool) ([]zsetEntry, error) {

	switch obj.encoding {
	case EncodingSkipList:
		zs, err := zsetSkiplistValue(obj)

		if err != nil {
			return nil, err
		}

		zsNodes := zs.ordered.rangeByRank(rankRange, rev)

		entries := make([]zsetEntry, len(zsNodes))

		for i, n := range zsNodes {
			entries[i] = zsetEntry{Member: n.member, Score: n.score}
		}

		return entries, nil

	default:
		return nil, ErrInvalidEncoding
	}

}

func zsetTypeRangeByScore(obj *RedisObject, r scoreRange) ([]zsetEntry, error) {
	switch obj.encoding {
	case EncodingSkipList:
		zs, err := zsetSkiplistValue(obj)
		if err != nil {
			return nil, err
		}

		zslNodes := zs.ordered.rangeByScore(r)

		entries := make([]zsetEntry, len(zslNodes))
		for i, n := range zslNodes {
			entries[i] = zsetEntry{Member: n.member, Score: n.score}
		}

		return entries, nil

	default:
		return nil, ErrInvalidEncoding
	}
}

func getZsetForWrite(db *RedisDb, key string) (*zset, error) {

	rObj, exists := db.lookupKey(key)

	if !exists {
		rObj = newZsetObject()
		db.setKey(key, rObj)
	}

	zset, err := zsetSkiplistValue(rObj)

	if err != nil {
		return nil, err
	}

	return zset, nil
}

func getZsetForRead(db *RedisDb, key string) (*zset, bool, error) {
	rObj, exists := db.lookupKey(key)
	if !exists {
		return nil, false, nil
	}

	zset, err := zsetSkiplistValue(rObj)

	if err != nil {
		return nil, false, err
	}

	return zset, true, nil
}

func newZsetObject() *RedisObject {
	return &RedisObject{
		typ:       ZSetObject,
		encoding:  EncodingSkipList,
		ptr:       createZset(),
		expiresAt: noExpiration,
	}
}

type zPopOptions struct {
	popMin bool
	count  int
}

func zsetTypePop(o *RedisObject, min bool, count int) (poppedElements []zsetEntry, err error) {
	switch o.encoding {
	case EncodingSkipList:
		z, err := zsetSkiplistValue(o)

		if err != nil {
			return poppedElements, err
		}

		var entries []zsetEntry

		if min {
			entries = z.popMin(count)
		} else {
			entries = z.popMax(count)
		}

		return entries, nil

	default:
		return poppedElements, ErrInvalidEncoding
	}
}

func formatScore(score float64) string {
	return strconv.FormatFloat(score, 'g', -1, 64)
}

func parseZsetPairs(args []string) (pairs []struct {
	score  float64
	member string
}, err error) {

	if len(args)%2 != 0 {
		return pairs, ErrWrongArgs
	}

	for i := 0; i < len(args); i += 2 {
		s, m := args[i], args[i+1]

		score, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return pairs, ErrInvalidEncoding
		}

		pairs = append(pairs, struct {
			score  float64
			member string
		}{score: score, member: m})
	}

	return pairs, nil
}

func Zadd(c *Client, args []string) CommandResult {
	// > ZADD racer_scores 8 "Sam-Bodden" 10 "Royce" 6 "Ford" 14 "Prickett"

	// returns number of added (not changed) members\

	// ! FLAGS (todo, in progress)
	// nx/xx (only add new members, dont update exsisting or only update existing, dont add new)
	// gt vs lt (only update existing if new score gt current score)
	// ch: change the return to number of changed members
	// incr: increment members score by score rather the setting, but only allows 1 pair
	key := args[0]

	zops, sIdx := parseZaddOptions(args[1:])

	pairs, err := parseZsetPairs(args[sIdx:])

	if err != nil {
		return Failed(Error(err.Error()))
	}

	added := 0

	robj, exists := c.db.lookupKey(key)

	if !exists {
		robj = newZsetObject()
		c.db.setKeyLocked(key, robj)
	}

	for _, vals := range pairs {
		addRes, err := zsetTypeAdd(robj, zsetEntry{Member: vals.member, Score: vals.score}, zops)

		if err != nil {
			return Failed(Error(err.Error()))
		}

		if addRes.Added {
			added += 1
		}
		if zops.CH && addRes.Updated {
			// allow for ch to have this be the total changed members
			added += 1
		}
	}

	// scoreMemberPairs := args[1:]
	// zadd adds in the set, or replaces it!

	return Result(Integer(added))

}

func ZIncrBy(c *Client, args []string) CommandResult {
	key, incR, member := args[0], args[1], args[2]

	incr, err := strconv.ParseFloat(incR, 64)

	if err != nil {
		return Failed(syntaxError())
	}

	o, exists := c.db.lookupKey(key)

	if !exists {
		o = newZsetObject()
	}

	addRes, err := zsetTypeAdd(o, zsetEntry{Member: member, Score: incr}, zsetAddOptions{Incr: true})

	if err != nil {
		return Failed(Error(err.Error()))
	}

	return Result(BulkString(formatScore(addRes.Score)))

}

func ZCard(c *Client, args []string) CommandResult {
	key := args[0]

	count := 0
	rObj, exists := c.db.lookupKey(key)

	if exists {

		len, err := zsetTypeLen(rObj)

		if err != nil {
			return Failed(Error(err.Error()))
		}
		count = int(len)
	}

	return Result(Integer(count))
}

func ZCount(c *Client, args []string) CommandResult {
	// ZCOUNT key min max
	key, minR, maxR := args[0], args[1], args[2]
	sr, err := parseScoreRange(minR, maxR)

	if err != nil {
		return Failed(Error(err.Error()))
	}

	o, exists := c.db.lookupKey(key)

	count := 0

	if exists {
		entries, err := zsetTypeRangeByScore(o, sr)

		if err != nil {
			return Failed(Error(err.Error()))
		}

		count = len(entries)

	}

	return Result(Integer(count))

}

func ZRank(c *Client, args []string) CommandResult {
	key, member := args[0], args[1]

	withScore := len(args) > 2

	if withScore && args[2] != "WITHSCORE" {
		return Failed(wrongArgs("ZRANK"))
	}

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Null())
	}

	rank, found, err := zsetTypeRank(obj, member)

	if err != nil {
		return Failed(Error(err.Error()))
	}

	if !found {
		return Result(Null())
	}

	return Result(Integer(int(rank)))

}

func parseScoreRange(startRaw string, endRaw string) (sr scoreRange, err error) {
	// parse floats AND inf potentially later with the (val inclusivity

	if startRaw == "-inf" {
		sr.min.value = -math.MaxFloat64
	} else {
		start, err := strconv.ParseFloat(startRaw, 64)
		if err != nil {
			return sr, err
		}
		sr.min.value = start
	}

	if endRaw == "+inf" {
		sr.max.value = math.MaxFloat64
	} else {
		end, err := strconv.ParseFloat(endRaw, 64)

		if err != nil {
			return sr, err
		}

		sr.max.value = end
	}

	return sr, nil
}

func parseRankRange(startRaw string, endRaw string) (rr rankRange, err error) {
	// parses the range values here for zrange

	start, err := strconv.ParseInt(startRaw, 10, 64)

	if err != nil {
		return rr, err
	}

	end, err := strconv.ParseInt(endRaw, 10, 64)

	if err != nil {
		return rr, err
	}
	rr.start = start
	rr.stop = end
	return rr, err

}

func zNodesToMembersArray(node []*zslNode, withScores bool) []Value {

	vArray := make([]Value, 0)

	for _, n := range node {
		vArray = append(vArray, BulkString(n.member))
		if withScores {
			vArray = append(vArray, BulkString(formatScore(n.score)))
		}
	}
	return vArray
}

func parseZrangeOptions(optionalArgs []string) (options zrangeOptions, err error) {

	for i, arg := range optionalArgs {

		switch strings.ToUpper(arg) {
		case "BYSCORE":
			options.mode = zrangeByScore

		case "REV":
			options.reverse = true
		case "LIMIT":

			if i+2 > len(optionalArgs) {
				return options, ErrWrongArgs
			}

			offsetStr := optionalArgs[i+1]

			offset, err := strconv.ParseInt(offsetStr, 10, 64)

			if err != nil {
				return options, ErrWrongArgs
			}

			count, err := strconv.ParseInt(optionalArgs[i+2], 10, 64)

			if err != nil {
				return options, ErrWrongArgs
			}

			if options.limit == nil {
				options.limit = &zrangeLimit{}
			}

			options.limit.offset = offset
			options.limit.count = count

			i += 2

		case "WITHSCORES":
			options.withScores = true
		default:
			return options, ErrWrongArgs
		}

	}

	return options, err

}

func ZRange(c *Client, args []string) CommandResult {
	// key start stop [REV] [LIMIT offset count]
	// [WITHSCORES]

	key, startRaw, stopRaw, optionalArgs := args[0], args[1], args[2], args[3:]

	options, err := parseZrangeOptions(optionalArgs)

	if err != nil {
		return Failed(Error(err.Error()))
	}

	o, exists := c.db.lookupKey(key)

	var responseEntries []zsetEntry

	if exists {

		switch options.mode {
		case zrangeByRank:
			rr, err := parseRankRange(startRaw, stopRaw)
			if err != nil {
				return Failed(Error(err.Error()))
			}

			responseEntries, err = zsetTypeRangeByRank(o, rr, options.reverse)
		case zrangeByScore:
			sr, err := parseScoreRange(startRaw, stopRaw)
			if err != nil {
				return Failed(Error(err.Error()))
			}

			responseEntries, err = zsetTypeRangeByScore(o, sr)

		default:
			return Failed(internalError())
		}
	}

	if err != nil {
		return Failed(Error(err.Error()))
	}

	return Result(Array(serializeZsetEntries(responseEntries, options.withScores)))
}

func ZRem(c *Client, args []string) CommandResult {

	key, members := args[0], args[1:]

	o, exists := c.db.lookupKey(key)

	remCount := 0
	if exists {
		for _, m := range members {

			removed, err := zsetTypeRemove(o, m)
			if err != nil {
				return Failed(Error(err.Error()))
			}
			if removed {
				remCount++
			}
		}
	}
	return Result(Integer(remCount))
}

func ZRemRangeByRank(c *Client, args []string) CommandResult {
	key, startR, stopR := args[0], args[1], args[2]
	rr, err := parseRankRange(startR, stopR)
	if err != nil {
		return Failed(Error(err.Error()))
	}
	o, exists := c.db.lookupKey(key)
	membersRemoved := 0

	if exists {
		entries, err := zsetTypeRangeByRank(o, rr, false)

		if err != nil {
			return Failed(Error(err.Error()))
		}

		for _, n := range entries {
			removed, err := zsetTypeRemove(o, n.Member)
			if err != nil {
				return Failed(Error(err.Error()))
			}
			if removed {
				membersRemoved++
			}
		}

	}
	return Result(Integer(membersRemoved))
}

func ZRemRangeByScore(c *Client, args []string) CommandResult {

	key, minR, maxR := args[0], args[1], args[2]

	sr, err := parseScoreRange(minR, maxR)

	if err != nil {
		return Failed(Error(err.Error()))
	}

	o, exists := c.db.lookupKey(key)

	membersRemoved := 0

	if exists {
		entries, err := zsetTypeRangeByScore(o, sr)

		if err != nil {
			return Failed(Error(err.Error()))
		}

		for _, e := range entries {
			removed, err := zsetTypeRemove(o, e.Member)

			if err != nil {
				return Failed(Error(err.Error()))
			}

			if removed {
				membersRemoved++
			}
		}

	}
	return Result(Integer(membersRemoved))
}

func ZScore(c *Client, args []string) CommandResult {
	key, member := args[0], args[1]

	rObj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Null())
	}

	score, found, err := zsetTypeScore(rObj, member)

	if err != nil {
		return Failed(Error(err.Error()))
	}

	if !found {
		return Result(Null())
	}

	return Result(BulkString(formatScore(score)))

}

func serializeZsetEntries(entries []zsetEntry, withScores bool) []Value {

	respArraySize := len(entries)

	if withScores {
		respArraySize *= 2
	}

	vArray := make([]Value, respArraySize)

	for i, n := range entries {
		mR := BulkString(n.Member)
		sR := BulkString(formatScore(n.Score))

		if withScores {
			vArray[i*2] = mR
			vArray[i*2+1] = sR
		} else {
			vArray[i] = mR
		}
	}
	return vArray
}

func ZPopMax(c *Client, args []string) CommandResult {
	// list of popped elements and scores
	key := args[0]
	var count uint64 = 1

	if len(args) > 1 {
		c, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return Failed(Error(err.Error()))
		}
		count = c
	}
	var poppedEntires []zsetEntry
	o, exists := c.db.lookupKey(key)

	if exists {
		r, err := zsetTypePop(o, false, int(count))

		if err != nil {
			return Failed(Error(err.Error()))
		}
		poppedEntires = r
	}

	return Result(Array(serializeZsetEntries(poppedEntires, true)))
}

func ZPopMin(c *Client, args []string) CommandResult {
	// list of popped elements and scores
	key := args[0]

	var count uint64 = 1

	if len(args) > 1 {
		c, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return Failed(Error(err.Error()))
		}

		count = c
	}

	var popped []zsetEntry
	o, exists := c.db.lookupKey(key)
	if exists {
		r, err := zsetTypePop(o, true, int(count))

		if err != nil {
			return Failed(Error(err.Error()))
		}

		popped = r
	}

	return Result(Array(serializeZsetEntries(popped, true)))
}

func ZMPop(c *Client, args []string) CommandResult {
	nKeys := args[0]
	nk, err := strconv.ParseInt(nKeys, 10, 64)

	if err != nil {
		return Failed(invalidInteger())
	}

	numKeys := int(nk)

	if numKeys <= 0 || len(args) < numKeys+2 {
		return Failed(wrongArgs("ZMPOP"))
	}

	//! deduplicate same keys maybe?
	keys := args[1 : numKeys+1]

	modifier := strings.ToUpper(args[numKeys+1])

	if modifier != "MIN" && modifier != "MAX" {
		return Failed(wrongArgs("ZMPOP"))
	}

	popMax := modifier == "MAX"

	var count uint64 = 1

	restArgs := args[numKeys+2:]
	if len(restArgs) >= 2 && strings.ToUpper(restArgs[0]) == "COUNT" {
		// check the next count maybe
		c, err := strconv.ParseUint(restArgs[1], 10, 64)
		if err != nil {
			return Failed(Error(err.Error()))
		}
		count = c

	}

	for _, k := range keys {
		o, exists := c.db.lookupKey(k)

		if !exists {
			continue
		}

		centries, err := zsetTypePop(o, !popMax, int(count))

		if err != nil {
			return Failed(Error(err.Error()))
		}
		if len(centries) > 0 {
			return Result(Array([]Value{BulkString(k), Array(serializeZsetEntries(centries, true))}))
		}
	}

	return Result(Null())
}

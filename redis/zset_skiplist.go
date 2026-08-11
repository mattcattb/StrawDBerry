package redis

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
)

const MAX_ZSL_HEIGHT = 32

// zset is the skiplist encoding of a sorted set. It owns both indexes:
// byMember provides member lookup, while ordered provides score/rank traversal.
type zset struct {
	byMember map[string]*zslNode
	ordered  *zskiplist
}

func createZset() *zset {
	return &zset{
		byMember: make(map[string]*zslNode),
		ordered:  createZskiplist(),
	}
}

// zsetSkiplistValue validates and extracts the skiplist-specific payload.
func zsetSkiplistValue(o *RedisObject) (*zset, error) {
	if o.typ != ZSetObject {
		return nil, ErrWrongType
	}
	if o.encoding != EncodingSkipList {
		return nil, ErrWrongType
	}

	zs, ok := o.ptr.(*zset)
	if !ok {
		return nil, ErrWrongType
	}
	return zs, nil
}

type zslNode struct {
	score    float64
	member   string
	backward *zslNode
	levels   []zslLevel
}

type zslLevel struct {
	forward *zslNode
	span    int // how many nodes skipped to get to the new area
}

type zskiplist struct {
	header *zslNode
	tail   *zslNode
	len    int
	height int
}

func createZskiplist() *zskiplist {
	return &zskiplist{
		header: &zslNode{
			levels: make([]zslLevel, MAX_ZSL_HEIGHT),
		},
		height: 1,
	}
}

func (z *zset) popMin(count int) (pEntries []zsetEntry) {
	totalRemoved := 0
	for i := 0; i < count; i++ {
		if z.ordered.len <= 0 {
			break
		}

		head := z.ordered.header.levels[0].forward
		if head == nil {
			break
		}
		pEntries = append(pEntries, zsetEntry{Member: head.member, Score: head.score})
		z.remove(head.member)
		totalRemoved++
	}
	return pEntries
}

func (z *zset) popMax(count int) (pEntries []zsetEntry) {
	totalRemoved := 0

	for i := 0; i < count; i++ {
		if z.ordered.len <= 0 || z.ordered.tail == nil {
			break
		}
		target := z.ordered.tail
		z.remove(target.member)
		pEntries = append(pEntries, zsetEntry{Member: target.member, Score: target.score})
		totalRemoved++
	}
	return pEntries
}

var errMemberExists error = errors.New("error exists")

var errNanScore error = errors.New("invalid score")

// only allow insertation of a new node
func (z *zset) insertNew(member string, score float64) error {

	if math.IsNaN(score) {
		return errNanScore
	}

	if _, exists := z.byMember[member]; exists {
		return errMemberExists
	}

	node := z.ordered.insert(score, member)

	if node == nil {
		return ErrInvalidEncoding
	}

	z.byMember[member] = node
	return nil
}

func (z *zset) updateScore(node *zslNode, newScore float64) (bool, error) {
	if math.IsNaN(newScore) {
		return false, errNanScore
	}

	current, exists := z.byMember[node.member]
	if !exists || current != node {
		return false, fmt.Errorf("node not owned")
	}

	if node.score == newScore {
		return false, nil
	}
	z.ordered.unlinkNode(node)
	node.score = newScore
	z.ordered.linkNode(node)
	return true, nil
}

func (z *zset) remove(member string) bool {
	node, exists := z.byMember[member]
	if !exists {
		return false
	}
	z.ordered.unlinkNode(node)
	delete(z.byMember, member)
	return true
}

func (z *zset) lookup(member string) (*zslNode, bool) {
	node, exists := z.byMember[member]

	return node, exists
}

func (z *zset) score(member string) (float64, bool) {
	node, exists := z.byMember[member]
	if !exists {
		return 0, false
	}
	return node.score, true
}

// rank is currently the skiplist's internal one-based rank.
func (z *zset) rank(member string) (int, bool) {
	node, exists := z.byMember[member]
	if !exists {
		return 0, false
	}
	return z.ordered.rankOf(node), true
}

func randomHeight() int {
	height := 1

	for height < MAX_ZSL_HEIGHT && rand.Float64() < 0.25 {
		height++
	}

	return height
}

func (a *zslNode) cmp(b *zslNode) int {
	// -1: [a, b], 0: same, 1: [b, a]
	cmp := a.score - b.score

	if cmp < 0 {
		return -1
	} else if cmp > 0 {
		return 1
	}
	// cmp == 0, lex compare
	return strings.Compare(a.member, b.member)
}

func (a *zslNode) lt(b *zslNode) bool {
	return a.cmp(b) == -1
}

func normalizeRankRange(length int64, rr rankRange) (uint64, uint64, bool) {
	// 0 <= startIndex <= stopIndex < length
	// negative start index subtracts

	start, stop := rr.start, rr.stop

	if start < 0 {
		start = length + start
	}

	if stop < 0 {
		stop = length + stop
	}

	if stop < 0 {
		stop = 0
	}

	if start < 0 {
		start = 0
	}

	if stop > length {
		stop = length
	}

	if start > length {
		start = length
	}

	return uint64(start), uint64(stop), start <= stop
}

func (z *zskiplist) rangeByRank(rRange rankRange, rev bool) []*zslNode {

	// first normalize positions

	start, stop, ok := normalizeRankRange(int64(z.len), rRange)

	foundNodes := make([]*zslNode, 0)

	if !ok {
		return foundNodes
	}

	startIdx := start

	if rev {
		//
		startIdx = uint64(z.len) - startIdx
	}

	startNode := z.nodeAtRank(uint64(startIdx) + 1)

	// 1, 5 normal: start at rank 1 go forward until 5
	// 1, 5 rev: actual ranks 4, 0 start at 4, go backwards
	for start <= stop {
		if startNode == nil {
			break
		}
		foundNodes = append(foundNodes, startNode)
		if rev {
			startNode = startNode.backward
		} else {
			startNode = startNode.levels[0].forward
		}
		start++
	}

	return foundNodes

}

func (z *zskiplist) rangeByScore(sRange scoreRange) []*zslNode {
	// min can be -inf, max can be +inf

	nodes := make([]*zslNode, 0)

	cur := z.firstInScoreRange(sRange.min.value)
	for cur.score <= sRange.max.value {
		if cur == nil {
			break
		}
		nodes = append(nodes, cur)
		cur = cur.levels[0].forward
	}

	return nodes
}

func (z *zskiplist) firstInScoreRange(score float64) *zslNode {

	// get the first element
	cur := z.header

	for l := z.height - 1; l >= 0; l-- {
		for cur.levels[l].forward != nil && cur.levels[l].forward.score < score {
			cur = cur.levels[l].forward
		}
	}

	return cur.levels[0].forward
}

func (z *zskiplist) insert(score float64, member string) *zslNode {

	nodeHeight := randomHeight()

	if nodeHeight > z.height {
		for level := z.height; level < nodeHeight; level++ {
			z.header.levels[level].span = z.len
		}
		z.height = nodeHeight
	}

	node := zslNode{score: score, member: member, levels: make([]zslLevel, nodeHeight)}

	z.linkNode(&node)

	return &node
}

func (z *zskiplist) linkNode(node *zslNode) {
	predecessors, rankBefore := z.findPredecessors(node.score, node.member)

	if existing := predecessors[0].levels[0].forward; existing != nil && existing.cmp(node) == 0 {
		//! node with member + score already exists
		return
	}

	rankNew := rankBefore[0] + 1

	// span split algorithm
	for level := 0; level < z.height; level++ {

		if level >= len(node.levels) {
			// points beyond new node height, so just add to its span
			predecessors[level].levels[level].span++
			continue
		}

		prev := predecessors[level]
		spanPrev, nxt := prev.levels[level].span, prev.levels[level].forward
		rankPrev := rankBefore[level]
		// prev -> (p_span)  after
		// prev -> (n_p_span) new -> (n_i_span) after

		distanceToNew := rankNew - rankPrev - 1
		prev.levels[level] = zslLevel{span: distanceToNew, forward: node}
		node.levels[level].span = spanPrev - distanceToNew // new node level span
		node.levels[level].forward = nxt
	}

	// update the backward for the nxt node
	node.backward = predecessors[0]
	if node.levels[0].forward != nil {
		node.levels[0].forward.backward = node
	} else {
		z.tail = node
	}

	z.len++

}

func (z *zskiplist) updateScore(node *zslNode, newScore float64) {
	z.unlinkNode(node)

	node.score = newScore
	z.linkNode(node)

}

func (z *zskiplist) findPredecessors(score float64, member string) ([]*zslNode, []int) {
	preds := make([]*zslNode, z.height)
	prevRanks := make([]int, z.height)

	cur := z.header
	curRank := 0
	temp := zslNode{member: member, score: score}

	for level := z.height - 1; level >= 0; level-- {
		for nxt := cur.levels[level].forward; nxt != nil && nxt.cmp(&temp) < 0; {
			curRank += 1 + cur.levels[level].span
			cur = nxt
		}
		preds[level] = cur
		prevRanks[level] = curRank
	}

	return preds, prevRanks
}

func (z *zskiplist) top(n int) []*zslNode {
	// get the top ranked members

	topNodes := make([]*zslNode, 0)

	cur := z.header.levels[0].forward
	for i := 0; i < n && i < z.len; i++ {

		if cur == nil {
			break
		}

		topNodes = append(topNodes, cur)
		cur = cur.levels[0].forward
	}
	return topNodes
}

func (z *zskiplist) rankOf(node *zslNode) int {
	_, prevRank := z.findPredecessors(node.score, node.member)
	return prevRank[0] + 1
}

func (z *zskiplist) findOrdered(score float64, member string) (bool, *zslNode) {
	cur := z.header

	tmp := &zslNode{score: score, member: member}

	for i := z.height - 1; i >= 0; i-- {
		for cur.levels[i].forward != nil && tmp.cmp(cur.levels[i].forward) >= 0 {
			cur = cur.levels[i].forward
		}
	}

	isEqual := cur.cmp(tmp) == 0

	return isEqual, cur
}

func (z *zskiplist) unlinkNode(node *zslNode) {

	preds, _ := z.findPredecessors(node.score, node.member)

	for i := len(preds) - 1; i >= 0; i-- {
		if i >= len(node.levels) {
			// we need to just subtract the span of this prev and do nothing
			preds[i].levels[i].span--
			continue
		}
		// we need to preform the split operation here
		preds[i].levels[i].span += node.levels[i].span
		preds[i].levels[i].forward = node.levels[i].forward
		if i == 0 && node.levels[i].forward != nil {
			node.levels[i].forward.backward = preds[i]
		}
	}

	if preds[0].levels[0].forward == nil {
		if preds[0] == z.header {
			z.tail = nil
		} else {
			z.tail = preds[0]
		}
	}

	// we need to remove height that just links to null

	for z.height > 1 && z.header.levels[z.height-1].forward == nil {
		z.header.levels[z.height-1] = zslLevel{}
		z.height--
	}
	z.len--

}

func (z *zskiplist) nodeAtRank(rank uint64) *zslNode {

	cur := z.header

	curRank := 0

	for level := z.height - 1; level >= 0; level-- {
		for curLvl := cur.levels[level]; curLvl.forward != nil && curRank+1+curLvl.span <= int(rank); {
			curRank += 1 + curLvl.span
			cur = curLvl.forward
		}
	}
	return cur
}

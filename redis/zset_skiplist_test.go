package redis

import "testing"

// orderedScores reads only level 0. Level 0 is the complete ordered list;
// higher levels are implementation details used to accelerate searches.
func orderedScores(list *zskiplist) []float64 {
	if list == nil || list.header == nil || len(list.header.levels) == 0 {
		return nil
	}

	scores := make([]float64, 0, list.len)
	for node := list.header.levels[0].forward; node != nil; node = node.levels[0].forward {
		scores = append(scores, node.score)
	}
	return scores
}

func TestBasicSkiplistStartsEmpty(t *testing.T) {
	list := createZskiplist()
	if list == nil {
		t.Fatal("create returned nil")
	}

	if list.len != 0 {
		t.Fatalf("length = %d, want 0", list.len)
	}
	if got := orderedScores(list); len(got) != 0 {
		t.Fatalf("ordered scores = %v, want empty", got)
	}
	if found, _ := list.findOrdered(10, "ten"); found {
		t.Fatal("findOrdered on empty list reported a match")
	}
}

func TestBasicSkiplistKeepsLevelZeroOrderedByScore(t *testing.T) {
	list := createZskiplist()

	list.insert(30, "thirty")
	list.insert(10, "ten")
	list.insert(20, "twenty")

	want := []float64{10, 20, 30}
	got := orderedScores(list)
	if len(got) != len(want) {
		t.Fatalf("ordered scores = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ordered scores = %v, want %v", got, want)
		}
	}
	if list.len != 3 {
		t.Fatalf("length = %d, want 3", list.len)
	}
	if list.tail == nil || list.tail.score != 30 {
		t.Fatalf("tail = %#v, want score 30", list.tail)
	}
}

func TestBasicSkiplistOrdersEqualScoresByMember(t *testing.T) {
	list := createZskiplist()
	first := list.insert(20, "alpha")
	second := list.insert(20, "beta")

	if first == nil || second == nil || second == first {
		t.Fatal("equal scores with different members should create two nodes")
	}
	if list.len != 2 {
		t.Fatalf("length after equal-score inserts = %d, want 2", list.len)
	}
	if first.levels[0].forward != second {
		t.Fatal("equal scores were not ordered by member")
	}
}

func TestBasicSkiplistFindOrderedFindsExactKey(t *testing.T) {
	list := createZskiplist()
	list.insert(10, "ten")
	list.insert(20, "twenty")
	list.insert(30, "thirty")

	for _, entry := range []struct {
		score  float64
		member string
	}{{10, "ten"}, {20, "twenty"}, {30, "thirty"}} {
		found, node := list.findOrdered(entry.score, entry.member)
		if !found || node == nil {
			t.Fatalf("findOrdered(%v, %q) did not find the node", entry.score, entry.member)
		}
		if node.score != entry.score || node.member != entry.member {
			t.Fatalf("findOrdered(%v, %q) = %#v", entry.score, entry.member, node)
		}
	}

	if found, _ := list.findOrdered(25, "twenty-five"); found {
		t.Fatal("findOrdered reported a missing key as present")
	}
}

func TestBasicSkiplistUnlinkNodeOnlyRemovesTheRequestedNode(t *testing.T) {
	list := createZskiplist()
	list.insert(10, "ten")
	twenty := list.insert(20, "twenty")
	list.insert(30, "thirty")

	list.unlinkNode(twenty)

	want := []float64{10, 30}
	got := orderedScores(list)
	if len(got) != len(want) {
		t.Fatalf("ordered scores after remove = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ordered scores after remove = %v, want %v", got, want)
		}
	}
	if list.len != 2 {
		t.Fatalf("length after remove = %d, want 2", list.len)
	}
	if found, _ := list.findOrdered(20, "twenty"); found {
		t.Fatal("unlinked node was still found")
	}
}

func TestBasicSkiplistRemoveLastScoreResetsEmptyState(t *testing.T) {
	list := createZskiplist()
	node := list.insert(10, "ten")

	list.unlinkNode(node)
	if list.len != 0 {
		t.Fatalf("length after removing last node = %d, want 0", list.len)
	}
	if got := orderedScores(list); len(got) != 0 {
		t.Fatalf("ordered scores after removing last node = %v, want empty", got)
	}
	if list.tail != nil {
		t.Fatalf("tail after removing last node = %#v, want nil", list.tail)
	}
}

func TestZsetOwnsTheMemberIndex(t *testing.T) {
	set := createZset()

	if err := set.insertNew("sam", 8); err != nil {
		t.Fatalf("insertNew(sam, 8): %v", err)
	}
	node := set.byMember["sam"]
	changed, err := set.updateScore(node, 10)
	if err != nil || !changed {
		t.Fatalf("updateScore(sam, 10) = (%t, %v), want (true, nil)", changed, err)
	}

	if len(set.byMember) != 1 || set.ordered.len != 1 {
		t.Fatalf("member and ordered indexes diverged: members=%d ordered=%d", len(set.byMember), set.ordered.len)
	}
	if score, found := set.score("sam"); !found || score != 10 {
		t.Fatalf("score(sam) = (%v, %t), want (10, true)", score, found)
	}
}

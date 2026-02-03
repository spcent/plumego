package cache

import (
	"math/rand"
	"time"
)

const (
	maxLevel     = 32   // Maximum skip list level
	probability  = 0.25 // Probability for level promotion (matching Redis)
	maxRandLevel = maxLevel - 1
)

// init initializes the random seed for level generation
func init() {
	rand.Seed(time.Now().UnixNano())
}

// skipListNode represents a node in the skip list
type skipListNode struct {
	member   string          // Member name
	score    float64         // Score value
	backward *skipListNode   // Backward pointer for reverse traversal
	level    []skipListLevel // Forward pointers and spans at each level
}

// skipListLevel represents the forward pointer and span at a specific level
type skipListLevel struct {
	forward *skipListNode // Forward pointer to next node
	span    int64         // Distance to the next node (for rank calculation)
}

// skipList represents a skip list data structure for sorted sets
type skipList struct {
	header *skipListNode // Header node (sentinel)
	tail   *skipListNode // Tail node (last node)
	length int64         // Number of nodes in the list
	level  int32         // Current maximum level
}

// newSkipList creates a new skip list
func newSkipList() *skipList {
	return &skipList{
		header: newSkipListNode(maxLevel, "", 0),
		tail:   nil,
		length: 0,
		level:  1,
	}
}

// newSkipListNode creates a new skip list node with the specified level
func newSkipListNode(level int32, member string, score float64) *skipListNode {
	return &skipListNode{
		member:   member,
		score:    score,
		backward: nil,
		level:    make([]skipListLevel, level),
	}
}

// randomLevel generates a random level for a new node
// Uses P=0.25 probability (same as Redis)
func randomLevel() int32 {
	level := int32(1)
	for rand.Float64() < probability && level < maxLevel {
		level++
	}
	return level
}

// insert adds or updates a member in the skip list
// Returns the inserted/updated node
func (sl *skipList) insert(member string, score float64) *skipListNode {
	update := make([]*skipListNode, maxLevel)
	rank := make([]int64, maxLevel)

	x := sl.header

	// Find insertion position and track ranks
	for i := sl.level - 1; i >= 0; i-- {
		// Store rank that is crossed to reach the insert position
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score && x.level[i].forward.member < member)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	// Generate random level for the new node
	level := randomLevel()

	// If new level is higher than current max level, initialize new levels
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			rank[i] = 0
			update[i] = sl.header
			update[i].level[i].span = sl.length
		}
		sl.level = level
	}

	// Create the new node
	x = newSkipListNode(level, member, score)

	// Update forward pointers and spans
	for i := int32(0); i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		// Update span covered by update[i] as x is inserted here
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	// Increment span for untouched levels
	for i := level; i < sl.level; i++ {
		update[i].level[i].span++
	}

	// Set backward pointer
	if update[0] == sl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}

	// Update tail if needed
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		sl.tail = x
	}

	sl.length++
	return x
}

// delete removes a node with the specified member and score
// Returns true if the node was found and deleted
func (sl *skipList) delete(member string, score float64) bool {
	update := make([]*skipListNode, maxLevel)
	x := sl.header

	// Find the node to delete
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score && x.level[i].forward.member < member)) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	// x now points to the node before the target (or header)
	x = x.level[0].forward

	// Check if we found the target node
	if x != nil && x.score == score && x.member == member {
		sl.deleteNode(x, update)
		return true
	}

	return false
}

// deleteNode removes a node from the skip list
func (sl *skipList) deleteNode(x *skipListNode, update []*skipListNode) {
	// Update forward pointers and spans
	for i := int32(0); i < sl.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}

	// Update backward pointer
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		sl.tail = x.backward
	}

	// Update level if needed
	for sl.level > 1 && sl.header.level[sl.level-1].forward == nil {
		sl.level--
	}

	sl.length--
}

// getScore returns the score for a member, or false if not found
func (sl *skipList) getScore(member string) (float64, bool) {
	x := sl.header.level[0].forward

	// Search through level 0 (all nodes)
	for x != nil {
		if x.member == member {
			return x.score, true
		}
		x = x.level[0].forward
	}

	return 0, false
}

// getRank returns the rank (0-based) of a member
// Returns -1 if member not found
// If desc is true, returns rank in descending order
func (sl *skipList) getRank(member string, score float64, desc bool) int64 {
	rank := int64(0)
	x := sl.header

	// Find the node and accumulate rank
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score && x.level[i].forward.member <= member)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		// Found the member
		if x.member == member {
			if desc {
				return sl.length - rank
			}
			return rank - 1 // 0-based rank
		}
	}

	return -1
}

// getNodeByRank returns the node at the specified rank (1-based)
func (sl *skipList) getNodeByRank(rank int64) *skipListNode {
	if rank <= 0 || rank > sl.length {
		return nil
	}

	traversed := int64(0)
	x := sl.header

	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}

		if traversed == rank {
			return x
		}
	}

	return nil
}

// getRange returns nodes in the specified rank range
// start and stop are 0-based indices
// If desc is true, returns in descending order
func (sl *skipList) getRange(start, stop int64, desc bool) []*skipListNode {
	// Handle negative indices
	if start < 0 {
		start = sl.length + start
	}
	if stop < 0 {
		stop = sl.length + stop
	}

	// Validate range
	if start < 0 {
		start = 0
	}
	if stop >= sl.length {
		stop = sl.length - 1
	}

	rangeLen := stop - start + 1
	if rangeLen <= 0 {
		return nil
	}

	// Find starting node
	var x *skipListNode
	if desc {
		// Start from tail for descending order
		x = sl.tail
		if start > 0 {
			x = sl.getNodeByRank(sl.length - start)
		}
	} else {
		// Start from head for ascending order
		x = sl.header.level[0].forward
		if start > 0 {
			x = sl.getNodeByRank(start + 1)
		}
	}

	// Collect results
	results := make([]*skipListNode, 0, rangeLen)
	for i := int64(0); i < rangeLen && x != nil; i++ {
		results = append(results, x)
		if desc {
			x = x.backward
		} else {
			x = x.level[0].forward
		}
	}

	return results
}

// getRangeByScore returns nodes with scores in the specified range [min, max]
// If desc is true, returns in descending order
func (sl *skipList) getRangeByScore(min, max float64, desc bool) []*skipListNode {
	if min > max {
		return nil
	}

	var results []*skipListNode

	if desc {
		// Start from tail and go backwards
		x := sl.tail
		for x != nil && x.score > max {
			x = x.backward
		}

		// Collect nodes while score >= min
		for x != nil && x.score >= min {
			if x.score <= max {
				results = append(results, x)
			}
			x = x.backward
		}
	} else {
		// Start from header and go forward
		x := sl.header.level[0].forward
		for x != nil && x.score < min {
			x = x.level[0].forward
		}

		// Collect nodes while score <= max
		for x != nil && x.score <= max {
			results = append(results, x)
			x = x.level[0].forward
		}
	}

	return results
}

// count returns the number of nodes with scores in [min, max]
func (sl *skipList) count(min, max float64) int64 {
	if min > max {
		return 0
	}

	count := int64(0)
	x := sl.header.level[0].forward

	// Skip nodes with score < min
	for x != nil && x.score < min {
		x = x.level[0].forward
	}

	// Count nodes with min <= score <= max
	for x != nil && x.score <= max {
		count++
		x = x.level[0].forward
	}

	return count
}

// deleteRangeByRank removes nodes in the specified rank range [start, stop]
// Returns the number of nodes deleted
func (sl *skipList) deleteRangeByRank(start, stop int64) int64 {
	// Handle negative indices
	if start < 0 {
		start = sl.length + start
	}
	if stop < 0 {
		stop = sl.length + stop
	}

	// Validate range
	if start < 0 {
		start = 0
	}
	if stop >= sl.length {
		stop = sl.length - 1
	}

	if start > stop || sl.length == 0 {
		return 0
	}

	update := make([]*skipListNode, maxLevel)
	traversed := int64(0)
	x := sl.header

	// Find the position just before start
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && traversed+x.level[i].span < start+1 {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	// Delete nodes in range
	deleted := int64(0)
	x = x.level[0].forward

	for x != nil && traversed < stop+1 {
		next := x.level[0].forward
		sl.deleteNode(x, update)
		deleted++
		traversed++
		x = next
	}

	return deleted
}

// deleteRangeByScore removes nodes with scores in [min, max]
// Returns the number of nodes deleted
func (sl *skipList) deleteRangeByScore(min, max float64) int64 {
	if min > max {
		return 0
	}

	update := make([]*skipListNode, maxLevel)
	x := sl.header

	// Find the position just before min
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && x.level[i].forward.score < min {
			x = x.level[i].forward
		}
		update[i] = x
	}

	// Delete nodes in range
	deleted := int64(0)
	x = x.level[0].forward

	for x != nil && x.score <= max {
		next := x.level[0].forward
		sl.deleteNode(x, update)
		deleted++
		x = next
	}

	return deleted
}

// clear removes all nodes from the skip list
func (sl *skipList) clear() {
	sl.header = newSkipListNode(maxLevel, "", 0)
	sl.tail = nil
	sl.length = 0
	sl.level = 1
}

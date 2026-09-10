package engine

import (
	"math/rand"
)

// DEPRECATED

type ActionQueueOld []action

func (aq ActionQueueOld) Len() int {
	return len(aq)
}

func (aq ActionQueueOld) Less(i, j int) bool {
	if aq[i].prio(nil) < aq[j].prio(nil) {
		return true
	} else if aq[i].prio(nil) > aq[j].prio(nil) {
		return false
	} else if aq[i].speed(nil) < aq[j].speed(nil) {
		return true
	} else if aq[i].speed(nil) > aq[j].speed(nil) {
		return false
	}

	return (rand.Int() % 2) == 0
}

func (aq ActionQueueOld) Swap(i, j int) {
	aq[i], aq[j] = aq[j], aq[i]
}

func (aq *ActionQueueOld) Push(a any) {
	*aq = append(*aq, a.(action))
}

func (aq *ActionQueueOld) Pop() any {
	old := *aq
	n := len(old)
	if n == 0 {
		return nil
	}
	x := old[n-1]
	*aq = old[:n-1]
	return x
}

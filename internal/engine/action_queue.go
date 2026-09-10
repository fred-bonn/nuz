package engine

import (
	"slices"
)

type action interface {
	invoke(bs BattleState)
	prio(bs BattleState) int
	speed(bs BattleState) int
}

type priorityQueue[T any] []T

func (q *priorityQueue[T]) push(a T) {
	*q = append(*q, a)
}

func (q *priorityQueue[T]) pop() (T, bool) {
	l := len(*q)
	if l == 0 {
		return *new(T), false
	}

	a := (*q)[l-1]
	*q = (*q)[:l-1]

	return a, true
}

func (q *priorityQueue[T]) insertAt(a T, cmp func(T, T) bool) {
	for i := 0; i < len(*q); i++ {
		if cmp(a, (*q)[i]) {
			*q = slices.Insert(*q, i, a)
			return
		}
	}

	q.push(a)
}

func (q priorityQueue[T]) sortBy(f func(a, b T) int) bool {
	if f == nil {
		return false
	}

	slices.SortFunc(q, f)

	return true
}

func (q *priorityQueue[T]) fetchBy(f func(T) bool) (T, bool) {
	for i, e := range *q {
		if f(e) {
			*q = append((*q)[:i], (*q)[i+1:]...)
			return e, true
		}
	}

	return *new(T), false
}

type actionQueue struct {
	queue priorityQueue[action]
}

func (a *actionQueue) sort(bs BattleState) {
	cmp := func(b, c action) int {
		if b.prio(bs) > c.prio(bs) {
			return 1
		}
		if b.prio(bs) < c.prio(bs) {
			return -1
		}
		if b.speed(bs) > c.speed(bs) {
			return 1
		}
		if b.speed(bs) < c.speed(bs) {
			return -1
		}
		return rollInt(1, 2)*2 - 1
	}

	a.queue.sortBy(cmp)
}

func (a *actionQueue) containstSwitchTo(mon *Pokemon) bool {
	for _, action := range a.queue {
		if sa, ok := action.(*switchAction); ok && sa.new == mon {
			return true
		}
	}
	return false
}

func (a *actionQueue) getMoveActionBy(mon *Pokemon) *moveAction {
	for _, action := range a.queue {
		if ma, ok := action.(*moveAction); ok && mon == ma.userSlot.mon {
			return ma
		}
	}
	return nil
}
